package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

func CreateTable(dialect string, tbl Table) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	if len(tbl.Columns) == 0 {
		return "", fmt.Errorf("ddl: table has no columns")
	}
	buf.WriteString("CREATE TABLE ")
	if tbl.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, tbl.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, tbl.TableName))
	buf.WriteString(" (")
	var columnWritten bool
	for _, column := range tbl.Columns {
		if column.Ignore {
			continue
		}
		buf.WriteString("\n    ")
		if !columnWritten {
			columnWritten = true
		} else {
			buf.WriteString(",")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, column.ColumnName))
		if column.ColumnType != "" {
			buf.WriteString(" " + column.ColumnType)
		}
		var isGenerated bool
		if column.Identity != "" && dialect != sq.DialectMySQL && dialect != sq.DialectSQLite {
			buf.WriteString(" GENERATED " + column.Identity)
		} else if column.Autoincrement && (dialect == sq.DialectMySQL || dialect == sq.DialectSQLite) {
			switch dialect {
			case sq.DialectMySQL:
				buf.WriteString(" AUTO_INCREMENT")
			case sq.DialectSQLite:
				buf.WriteString(" AUTOINCREMENT")
			}
		} else if column.GeneratedExpr != "" {
			isGenerated = true
			buf.WriteString(" GENERATED ALWAYS AS (" + column.GeneratedExpr + ")") // TODO: c.GeneratedExprStored has to be sanitized and escaped
			// postgres defaults to STORED because it does not support virtual generated columns
			if column.GeneratedExprStored || dialect == sq.DialectPostgres {
				buf.WriteString(" STORED")
			} else {
				buf.WriteString(" VIRTUAL")
			}
		}
		if column.IsNotNull {
			buf.WriteString(" NOT NULL")
		}
		if column.ColumnDefault != "" && !isGenerated {
			buf.WriteString(" DEFAULT (" + column.ColumnDefault + ")") // TODO: c.ColumnDefault has to be sanitized and escaped
		}
		if column.OnUpdateCurrentTimestamp && dialect == sq.DialectMySQL && !isGenerated {
			buf.WriteString(" ON UPDATE CURRENT_TIMESTAMP")
		}
		if column.CollationName != "" {
			switch dialect {
			case sq.DialectPostgres:
				buf.WriteString(` COLLATE "` + sq.EscapeQuote(column.CollationName, '"') + `"`) // postgres collation names need double quotes (idk why)
			default:
				buf.WriteString(" COLLATE " + column.CollationName) // TODO: c.CollationName has to be sanitized and escaped
			}
		}
	}
	if dialect == sq.DialectSQLite {
		buf.WriteString("\n")
		for _, constraint := range tbl.Constraints {
			buf.WriteString("\n    ,CONSTRAINT " + constraint.ConstraintName + " " + constraint.ConstraintType)
			switch constraint.ConstraintType {
			case PRIMARY_KEY, UNIQUE:
				buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ")")
			case FOREIGN_KEY:
				buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ") REFERENCES ")
				if constraint.ReferencesSchema != "" {
					buf.WriteString(constraint.ReferencesSchema + ".")
				}
				buf.WriteString(constraint.ReferencesTable)
				if len(constraint.ReferencesColumns) > 0 {
					buf.WriteString(" (" + strings.Join(constraint.ReferencesColumns, ", ") + ")")
				}
				if constraint.MatchOption != "" {
					buf.WriteString(" " + constraint.MatchOption) // TODO: check for validity
				}
				if constraint.OnUpdate != "" {
					buf.WriteString(" ON UPDATE " + constraint.OnUpdate) // TODO: check for validity
				}
				if constraint.OnDelete != "" {
					buf.WriteString(" ON DELETE " + constraint.OnDelete) // TODO: check for validity
				}
			case CHECK:
				buf.WriteString(" (" + constraint.CheckExpr + ")")
			}
			var canDeferrable bool
			switch dialect {
			case sq.DialectPostgres:
				if constraint.ConstraintType != CHECK {
					canDeferrable = true
				}
			case sq.DialectSQLite:
				if constraint.ConstraintType == FOREIGN_KEY {
					canDeferrable = true
				}
			}
			if canDeferrable && constraint.IsDeferrable {
				buf.WriteString(" DEFERRABLE")
				if constraint.IsInitiallyDeferred {
					buf.WriteString(" INITIALLY DEFERRED")
				} else {
					buf.WriteString(" INITIALLY IMMEDIATE")
				}
			}
		}
	}
	if false && dialect == sq.DialectMySQL { // move this into its own CreateIndex function
		for _, index := range tbl.Indices {
			buf.WriteString("\n    ,")
			switch index.IndexType {
			case "FULLTEXT", "SPATIAL":
				buf.WriteString(index.IndexType + " INDEX " + index.IndexName)
			default:
				if index.IsUnique {
					buf.WriteString("UNIQUE ")
				}
				buf.WriteString("INDEX " + index.IndexName)
				if index.IndexType != "" && !strings.EqualFold(index.IndexType, "BTREE") {
					buf.WriteString(" USING " + index.IndexType)
				}
				buf.WriteString(" (")
				for j, column := range index.Columns {
					if j > 0 {
						buf.WriteString(", ")
					}
					if column != "" {
						buf.WriteString(column)
					} else {
						buf.WriteString("(" + index.Exprs[j] + ")")
					}
				}
				buf.WriteString(")")
			}
		}
	}
	buf.WriteString("\n);")
	return buf.String(), nil
}

func CreateConstraint(dialect string, constraint Constraint) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString("ALTER TABLE ")
	if constraint.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, constraint.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, constraint.TableName) + " ADD CONSTRAINT " + constraint.ConstraintName + " " + constraint.ConstraintType)
	switch constraint.ConstraintType {
	case PRIMARY_KEY, UNIQUE:
		buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ")")
	case FOREIGN_KEY:
		buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ") REFERENCES ")
		if constraint.ReferencesSchema != "" {
			buf.WriteString(constraint.ReferencesSchema + ".")
		}
		buf.WriteString(constraint.ReferencesTable)
		if len(constraint.ReferencesColumns) > 0 {
			buf.WriteString(" (" + strings.Join(constraint.ReferencesColumns, ", ") + ")")
		}
		if constraint.MatchOption != "" {
			buf.WriteString(" " + constraint.MatchOption) // TODO: check for validity
		}
		if constraint.OnUpdate != "" {
			buf.WriteString(" ON UPDATE " + constraint.OnUpdate) // TODO: check for validity
		}
		if constraint.OnDelete != "" {
			buf.WriteString(" ON DELETE " + constraint.OnDelete) // TODO: check for validity
		}
	case CHECK:
		buf.WriteString(" (" + constraint.CheckExpr + ")")
	}
	var deferSupported bool
	switch dialect {
	case sq.DialectPostgres:
		if constraint.ConstraintType != CHECK {
			deferSupported = true
		}
	case sq.DialectSQLite:
		if constraint.ConstraintType == FOREIGN_KEY {
			deferSupported = true
		}
	}
	if deferSupported && constraint.IsDeferrable {
		buf.WriteString(" DEFERRABLE")
		if constraint.IsInitiallyDeferred {
			buf.WriteString(" INITIALLY DEFERRED")
		} else {
			buf.WriteString(" INITIALLY IMMEDIATE")
		}
	}
	buf.WriteString(";")
	return buf.String(), nil
}

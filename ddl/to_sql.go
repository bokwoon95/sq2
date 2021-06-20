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
	var columned bool
	for _, column := range tbl.Columns {
		if column.Ignore {
			continue
		}
		buf.WriteString("\n    ")
		if columned {
			buf.WriteString(",")
		}
		columned = true
		buf.WriteString(sq.QuoteIdentifier(dialect, column.ColumnName))
		if column.ColumnType != "" {
			buf.WriteString(" " + column.ColumnType)
		}
		if column.Identity != "" {
			switch dialect {
			case sq.DialectMySQL, sq.DialectSQLite:
				break // mysql and sqlite do not support identity columns
			default:
				switch column.Identity {
				case BY_DEFAULT_AS_IDENTITY:
					buf.WriteString(" GENERATED BY DEFAULT AS IDENTITY")
				case ALWAYS_AS_IDENTITY:
					buf.WriteString(" GENERATED ALWAYS AS IDENTITY")
				}
			}
		} else if column.Autoincrement {
			switch dialect {
			case sq.DialectMySQL:
				buf.WriteString(" AUTO_INCREMENT")
			case sq.DialectSQLite:
				buf.WriteString(" AUTOINCREMENT")
			}
		} else if column.GeneratedExpr != "" {
			buf.WriteString(" GENERATED ALWAYS AS (" + column.GeneratedExpr + ")") // TODO: c.GeneratedExprStored has to be sanitized and escaped
			if column.GeneratedExprStored {
				buf.WriteString(" STORED")
			} else {
				switch dialect {
				case sq.DialectPostgres:
					break // postgres does not support virtual generated columns
				default:
					buf.WriteString(" VIRTUAL")
				}
			}
		}
		if column.IsNotNull {
			buf.WriteString(" NOT NULL")
		}
		if column.ColumnDefault != "" {
			buf.WriteString(" DEFAULT (" + column.ColumnDefault + ")") // TODO: c.ColumnDefault has to be sanitized and escaped
		}
		if column.CollationName != "" {
			switch dialect {
			case sq.DialectPostgres:
				buf.WriteString(` COLLATE "` + sq.EscapeQuote(column.CollationName, '"') + `"`) // postgres collation names need double quotes (idk why)
			default:
				buf.WriteString(" COLLATE " + column.CollationName) // TODO: c.CollationName has to be sanitized and escaped
			}
		}
		if column.OnUpdateCurrentTimestamp {
			switch dialect {
			case sq.DialectMySQL:
				buf.WriteString(" ON UPDATE CURRENT_TIMESTAMP")
			}
		}
	}
	var newlined bool
	if dialect == sq.DialectSQLite {
		for _, constraint := range tbl.Constraints {
			if !newlined {
				buf.WriteString("\n")
				newlined = true
			}
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
			if constraint.IsDeferrable {
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
		if !newlined {
			buf.WriteString("\n")
			newlined = true
		}
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
	if constraint.IsDeferrable {
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

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
	if tbl.TableName == "" {
		return "", fmt.Errorf("ddl: table has no name")
	}
	if len(tbl.Columns) == 0 {
		return "", fmt.Errorf("ddl: table %s has no columns", tbl.TableName)
	}
	if tbl.VirtualTable != "" {
		if dialect != sq.DialectSQLite {
			return buf.String(), fmt.Errorf("ddl: only SQLite has VIRTUAL TABLE support (table=%s)", tbl.TableName)
		}
		buf.WriteString("CREATE VIRTUAL TABLE ")
	} else {
		buf.WriteString("CREATE TABLE ")
	}
	if tbl.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, tbl.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, tbl.TableName))
	if tbl.VirtualTable != "" {
		buf.WriteString(" USING " + tbl.VirtualTable)
	}
	buf.WriteString(" (")
	var columnWritten bool
	for _, column := range tbl.Columns {
		if column.Ignore {
			continue
		}
		if !columnWritten {
			columnWritten = true
			buf.WriteString("\n    ")
		} else {
			buf.WriteString("\n    ,")
		}
		if strings.EqualFold(tbl.VirtualTable, "fts5") {
			column = Column{ColumnName: column.ColumnName}
		}
		err := createColumn(dialect, buf, column)
		if err != nil {
			return buf.String(), err
		}
	}
	if len(tbl.VirtualTableArgs) > 0 && tbl.VirtualTable == "" {
		return buf.String(), fmt.Errorf("ddl: virtual table arguments present without a virtual table module")
	}
	if tbl.VirtualTable != "" && dialect == sq.DialectSQLite && len(tbl.VirtualTableArgs) > 0 {
		for _, arg := range tbl.VirtualTableArgs {
			if !columnWritten {
				columnWritten = true
				buf.WriteString("\n    ")
			} else {
				buf.WriteString("\n    ,")
			}
			buf.WriteString(arg)
		}
	}
	var newlineWritten bool
	for _, constraint := range tbl.Constraints {
		if dialect == sq.DialectSQLite && constraint.ConstraintType == PRIMARY_KEY && len(constraint.Columns) == 1 {
			continue
		}
		if dialect != sq.DialectSQLite && constraint.ConstraintType == FOREIGN_KEY {
			continue
		}
		if !newlineWritten {
			buf.WriteString("\n")
			newlineWritten = true
		}
		buf.WriteString("\n    ,CONSTRAINT ")
		err := createConstraint(dialect, buf, constraint)
		if err != nil {
			return buf.String(), err
		}
	}
	buf.WriteString("\n);")
	return buf.String(), nil
}

func createColumn(dialect string, buf *bytes.Buffer, column Column) error {
	buf.WriteString(sq.QuoteIdentifier(dialect, column.ColumnName))
	if column.ColumnType != "" {
		buf.WriteString(" " + column.ColumnType)
	}
	if column.IsNotNull {
		buf.WriteString(" NOT NULL")
	}
	isAutoincrement := column.Autoincrement && (dialect == sq.DialectMySQL || dialect == sq.DialectSQLite)
	isIdentity := column.Identity != "" && (dialect != sq.DialectMySQL && dialect != sq.DialectSQLite)
	isGenerated := column.GeneratedExpr != ""
	if column.ColumnDefault != "" && !isAutoincrement && !isIdentity && !isGenerated {
		buf.WriteString(" DEFAULT " + column.ColumnDefault)
	}
	if column.IsPrimaryKey && dialect == sq.DialectSQLite {
		// only SQLite primary key is defined inline, others are defined as separate constraints
		buf.WriteString(" PRIMARY KEY")
	}
	if isAutoincrement && dialect != sq.DialectMySQL && dialect != sq.DialectSQLite {
		return fmt.Errorf("ddl: %s does not support autoincrement columns", dialect)
	}
	if isIdentity && (dialect == sq.DialectMySQL || dialect == sq.DialectSQLite) {
		return fmt.Errorf("ddl: %s does not support identity columns", dialect)
	}
	if isAutoincrement {
		switch dialect {
		case sq.DialectMySQL:
			buf.WriteString(" AUTO_INCREMENT")
		case sq.DialectSQLite:
			buf.WriteString(" AUTOINCREMENT")
		}
	} else if isIdentity {
		buf.WriteString(" GENERATED " + column.Identity)
	} else if isGenerated {
		buf.WriteString(" GENERATED ALWAYS AS (" + column.GeneratedExpr + ")")
		if column.GeneratedExprStored {
			buf.WriteString(" STORED")
		} else {
			if dialect == sq.DialectPostgres {
				return fmt.Errorf("ddl: Postgres does not support VIRTUAL generated columns")
			}
			buf.WriteString(" VIRTUAL")
		}
	}
	if column.OnUpdateCurrentTimestamp {
		if dialect != sq.DialectMySQL {
			return fmt.Errorf("ddl: %s does not support ON UPDATE CURRENT_TIMESTAMP", dialect)
		}
		buf.WriteString(" ON UPDATE CURRENT_TIMESTAMP")
	}
	if column.CollationName != "" {
		switch dialect {
		case sq.DialectPostgres:
			buf.WriteString(` COLLATE "` + sq.EscapeQuote(column.CollationName, '"') + `"`) // postgres collation names need double quotes (idk why)
		default:
			buf.WriteString(" COLLATE " + column.CollationName)
		}
	}
	return nil
}

func CreateColumn(dialect string, column Column) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString("ALTER TABLE ")
	if column.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, column.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, column.TableName) + " ADD COLUMN " + sq.QuoteIdentifier(dialect, column.ColumnName))
	err := createColumn(dialect, buf, column)
	if err != nil {
		return buf.String(), err
	}
	buf.WriteString(";")
	return buf.String(), nil
}

func createConstraint(dialect string, buf *bytes.Buffer, constraint Constraint) error {
	buf.WriteString(constraint.ConstraintName + " " + constraint.ConstraintType)
	switch constraint.ConstraintType {
	case CHECK:
		buf.WriteString(" (" + constraint.CheckExpr + ")")
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
			buf.WriteString(" " + constraint.MatchOption)
		}
		if constraint.OnUpdate != "" {
			buf.WriteString(" ON UPDATE " + constraint.OnUpdate)
		}
		if constraint.OnDelete != "" {
			buf.WriteString(" ON DELETE " + constraint.OnDelete)
		}
	default:
		buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ")")
	}
	var deferSupported bool
	if (dialect == sq.DialectPostgres && constraint.ConstraintType != CHECK) ||
		(dialect == sq.DialectSQLite && constraint.ConstraintType == FOREIGN_KEY) {
		deferSupported = true
	}
	if deferSupported && constraint.IsDeferrable {
		buf.WriteString(" DEFERRABLE")
		if constraint.IsInitiallyDeferred {
			buf.WriteString(" INITIALLY DEFERRED")
		} else {
			buf.WriteString(" INITIALLY IMMEDIATE")
		}
	}
	return nil
}

func CreateConstraint(dialect string, constraint Constraint) (string, error) {
	if dialect == sq.DialectSQLite {
		return "", fmt.Errorf("ddl: SQLite does not allow the creating of constraints separately")
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString("ALTER TABLE ")
	if constraint.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, constraint.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, constraint.TableName) + " ADD CONSTRAINT ")
	err := createConstraint(dialect, buf, constraint)
	if err != nil {
		return buf.String(), err
	}
	buf.WriteString(";")
	return buf.String(), nil
}

func CreateIndex(dialect string, index Index) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString("CREATE")
	isFulltextOrSpatial := index.IndexType == "FULLTEXT" || index.IndexType == "SPATIAL"
	if dialect == sq.DialectMySQL && isFulltextOrSpatial {
		buf.WriteString(" " + index.IndexType)
	} else if index.IsUnique {
		buf.WriteString(" UNIQUE")
	}
	buf.WriteString(" INDEX " + index.IndexName + " ON ")
	if index.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, index.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, index.TableName))
	if index.IndexType != "" && !isFulltextOrSpatial && !strings.EqualFold(index.IndexType, "BTREE") {
		buf.WriteString(" USING " + index.IndexType)
	}
	buf.WriteString(" (")
	for i, column := range index.Columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		if column != "" {
			buf.WriteString(column)
		} else {
			buf.WriteString(index.Exprs[i])
		}
	}
	buf.WriteString(")")
	if len(index.Include) > 0 && dialect == sq.DialectPostgres {
		buf.WriteString(" INCLUDE (" + strings.Join(index.Include, ", ") + ")")
	}
	if index.Predicate != "" && (dialect == sq.DialectPostgres || dialect == sq.DialectSQLite) {
		buf.WriteString(" WHERE " + index.Predicate)
	}
	buf.WriteString(";")
	return buf.String(), nil
}

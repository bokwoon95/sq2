package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Table struct {
	TableSchema      string       `json:",omitempty"`
	TableName        string       `json:",omitempty"`
	Columns          []Column     `json:",omitempty"`
	Constraints      []Constraint `json:",omitempty"`
	Indexes          []Index      `json:",omitempty"`
	Triggers         []Trigger    `json:",omitempty"`
	VirtualTable     string       `json:",omitempty"`
	VirtualTableArgs []string     `json:",omitempty"`
	SQL              string       `json:",omitempty"`
	Comment          string       `json:",omitempty"`
	Ignore           bool         `json:",omitempty"`
	columnCache      map[string]int
	constraintCache  map[string]int
	indexCache       map[string]int
	triggerCache     map[string]int
}

func (tbl *Table) CachedColumnPosition(columnName string) (columnPosition int) {
	if columnName == "" {
		return -1
	}
	columnPosition, ok := tbl.columnCache[columnName]
	if !ok {
		return -1
	}
	if columnPosition < 0 || columnPosition >= len(tbl.Columns) {
		delete(tbl.columnCache, columnName)
		return -1
	}
	column := tbl.Columns[columnPosition]
	if column.ColumnName != columnName || column.Ignore {
		delete(tbl.columnCache, columnName)
		return -1
	}
	return columnPosition
}

func (tbl *Table) AppendColumn(column Column) (columnPosition int) {
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnCache == nil {
		tbl.columnCache = make(map[string]int)
	}
	columnPosition = len(tbl.Columns) - 1
	tbl.columnCache[column.ColumnName] = columnPosition
	return columnPosition
}

func (tbl *Table) RefreshColumnCache() {
	if tbl.columnCache == nil && len(tbl.Columns) > 0 {
		tbl.columnCache = make(map[string]int)
	}
	for i, column := range tbl.Columns {
		if column.Ignore {
			continue
		}
		tbl.columnCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintPosition(constraintName string) (constraintPosition int) {
	if constraintName == "" {
		return -1
	}
	constraintPosition, ok := tbl.constraintCache[constraintName]
	if !ok {
		return -1
	}
	if constraintPosition < 0 || constraintPosition >= len(tbl.Constraints) {
		delete(tbl.constraintCache, constraintName)
		return -1
	}
	constraint := tbl.Constraints[constraintPosition]
	if constraint.ConstraintName != constraintName || constraint.Ignore {
		delete(tbl.constraintCache, constraintName)
		return -1
	}
	return constraintPosition
}

func (tbl *Table) AppendConstraint(constraint Constraint) (constraintPosition int) {
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintCache == nil {
		tbl.constraintCache = make(map[string]int)
	}
	constraintPosition = len(tbl.Constraints) - 1
	tbl.constraintCache[constraint.ConstraintName] = constraintPosition
	return constraintPosition
}

func (tbl *Table) RefreshConstraintCache() {
	if tbl.constraintCache == nil && len(tbl.Constraints) > 0 {
		tbl.constraintCache = make(map[string]int)
	}
	for i, constraint := range tbl.Constraints {
		if constraint.Ignore {
			continue
		}
		tbl.constraintCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexPosition(indexName string) (indexPosition int) {
	if indexName == "" {
		return -1
	}
	indexPosition, ok := tbl.indexCache[indexName]
	if !ok {
		return -1
	}
	if indexPosition < 0 || indexPosition >= len(tbl.Indexes) {
		delete(tbl.indexCache, indexName)
		return -1
	}
	index := tbl.Indexes[indexPosition]
	if index.IndexName != indexName || index.Ignore {
		delete(tbl.indexCache, indexName)
		return -1
	}
	return indexPosition
}

func (tbl *Table) AppendIndex(index Index) (indexPosition int) {
	tbl.Indexes = append(tbl.Indexes, index)
	if tbl.indexCache == nil {
		tbl.indexCache = make(map[string]int)
	}
	indexPosition = len(tbl.Indexes) - 1
	tbl.indexCache[index.IndexName] = indexPosition
	return indexPosition
}

func (tbl *Table) RefreshIndexesCache() {
	if tbl.indexCache == nil && len(tbl.Indexes) > 0 {
		tbl.indexCache = make(map[string]int)
	}
	for i, index := range tbl.Indexes {
		if index.Ignore {
			continue
		}
		tbl.indexCache[index.IndexName] = i
	}
}

func (tbl *Table) CachedTriggerPosition(triggerName string) (triggerPosition int) {
	if triggerName == "" {
		return -1
	}
	triggerPosition, ok := tbl.triggerCache[triggerName]
	if !ok {
		return -1
	}
	if triggerPosition < 0 || triggerPosition >= len(tbl.Triggers) {
		delete(tbl.triggerCache, triggerName)
		return -1
	}
	trigger := tbl.Triggers[triggerPosition]
	if trigger.TriggerName != triggerName || trigger.Ignore {
		delete(tbl.triggerCache, triggerName)
		return -1
	}
	return triggerPosition
}

func (tbl *Table) AppendTrigger(trigger Trigger) (triggerPosition int) {
	tbl.Triggers = append(tbl.Triggers, trigger)
	if tbl.triggerCache == nil {
		tbl.triggerCache = make(map[string]int)
	}
	triggerPosition = len(tbl.Triggers) - 1
	tbl.triggerCache[trigger.TriggerName] = triggerPosition
	return triggerPosition
}

func (tbl *Table) RefreshTriggerCache() {
	if tbl.triggerCache == nil && len(tbl.Triggers) > 0 {
		tbl.triggerCache = make(map[string]int)
	}
	for i, trigger := range tbl.Triggers {
		if trigger.Ignore {
			continue
		}
		tbl.triggerCache[trigger.TriggerName] = i
	}
}

func (tbl *Table) loadIndexConfig(dialect, tableSchema, tableName string, columns []string, config string) error {
	columnNames, modifiers, modifierPositions, err := tokenizeValue(config)
	if err != nil {
		return err
	}
	var indexName string
	if n, ok := modifierPositions["name"]; ok {
		indexName = modifiers[n][1]
	}
	if columnNames != "." && columnNames != "" {
		columns = strings.Split(columnNames, ",")
	}
	if len(columns) == 0 {
		return fmt.Errorf("%s.%s no index columns provided: %s", tableSchema, tableName, config)
	}
	if indexName == "" && len(columns) > 0 {
		indexName = generateName(INDEX, tableName, columns...)
	}
	var index Index
	if n := tbl.CachedIndexPosition(indexName); n >= 0 {
		index = tbl.Indexes[n]
		defer func() { tbl.Indexes[n] = index }()
	} else {
		index = Index{
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			IndexName:   indexName,
			Columns:     columns,
		}
		defer func() { tbl.AppendIndex(index) }()
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "cols":
			continue
		case "unique":
			index.IsUnique = true
		case "using":
			index.IndexType = strings.ToUpper(modifier[1])
		case "where":
			index.Predicate = modifier[1]
		case "include":
			index.IncludeColumns = strings.Split(modifier[1], ",")
		case "ignore":
			if modifier[1] == "" {
				index.Ignore = true
			} else {
				ignoredDialects := strings.Split(modifier[1], ",")
				for _, ignoredDialect := range ignoredDialects {
					if dialect == ignoredDialect {
						index.Ignore = true
						break
					}
				}
			}
		default:
			return fmt.Errorf("invalid modifier 'index.%s'", modifier[0])
		}
	}
	return nil
}

func (tbl *Table) loadConstraintConfig(dialect, constraintType, tableSchema, tableName string, columns []string, config string) error {
	value, modifiers, modifierPositions, err := tokenizeValue(config)
	if err != nil {
		return err
	}
	var constraintName string
	if n, ok := modifierPositions["name"]; ok {
		constraintName = modifiers[n][1]
	}
	var isReferences bool
	if constraintType == "references" {
		isReferences = true
		constraintType = FOREIGN_KEY
	}
	// "references" modifier is treated different from the rest
	// ("primarykey", "foreignkey", "unique") because the value refers to
	// columns of another table instead of columns of the current table. So
	// we must exclude "references" from this block of code.
	if !isReferences && value != "." && value != "" {
		columns = strings.Split(value, ",")
	}
	if len(columns) == 0 {
		return fmt.Errorf("%s.%s no constraint columns provided: %s", tableSchema, tableName, config)
	}
	if constraintName == "" && len(columns) > 0 {
		constraintName = generateName(constraintType, tableName, columns...)
	}
	var constraint Constraint
	if n := tbl.CachedConstraintPosition(constraintName); n >= 0 {
		constraint = tbl.Constraints[n]
		constraint.ConstraintType = constraintType
		defer func() { tbl.Constraints[n] = constraint }()
	} else {
		constraint = Constraint{
			TableSchema:    tableSchema,
			TableName:      tableName,
			ConstraintName: constraintName,
			ConstraintType: constraintType,
			Columns:        columns,
		}
		defer func() { tbl.AppendConstraint(constraint) }()
	}
	if isReferences {
		switch parts := strings.SplitN(value, ".", 3); len(parts) {
		case 1:
			constraint.ReferencesTable = parts[0]
			constraint.ReferencesColumns = columns
		case 2:
			constraint.ReferencesTable = parts[0]
			constraint.ReferencesColumns = strings.Split(parts[1], ",")
		case 3:
			constraint.ReferencesSchema = parts[0]
			constraint.ReferencesTable = parts[1]
			constraint.ReferencesColumns = strings.Split(parts[2], ",")
		}
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "name", "cols":
			continue
		case "references":
			switch parts := strings.SplitN(modifier[1], ".", 3); len(parts) {
			case 1:
				constraint.ReferencesTable = parts[0]
				constraint.ReferencesColumns = columns
			case 2:
				constraint.ReferencesTable = parts[0]
				constraint.ReferencesColumns = strings.Split(parts[1], ",")
			case 3:
				constraint.ReferencesSchema = parts[0]
				constraint.ReferencesTable = parts[1]
				constraint.ReferencesColumns = strings.Split(parts[2], ",")
			}
		case "onupdate":
			switch modifier[1] {
			case "cascade":
				constraint.UpdateRule = CASCADE
			case "restrict":
				constraint.UpdateRule = RESTRICT
			case "noaction":
				constraint.UpdateRule = NO_ACTION
			case "setnull":
				constraint.UpdateRule = SET_NULL
			case "setdefault":
				constraint.UpdateRule = SET_DEFAULT
			default:
				return fmt.Errorf("unknown value '%s' for 'references.onupdate' modifier", modifier[1])
			}
		case "ondelete":
			switch modifier[1] {
			case "cascade":
				constraint.DeleteRule = CASCADE
			case "restrict":
				constraint.DeleteRule = RESTRICT
			case "noaction":
				constraint.DeleteRule = NO_ACTION
			case "setnull":
				constraint.DeleteRule = SET_NULL
			case "setdefault":
				constraint.DeleteRule = SET_DEFAULT
			default:
				return fmt.Errorf("unknown value '%s' for 'references.ondelete' modifier", modifier[1])
			}
		case "deferrable":
			if dialect != sq.DialectMySQL {
				constraint.IsDeferrable = true
			}
		case "deferred":
			if dialect != sq.DialectMySQL {
				constraint.IsDeferrable = true
				constraint.IsInitiallyDeferred = true
			}
		case "ignore":
			if modifier[1] == "" {
				constraint.Ignore = true
			} else {
				ignoredDialects := strings.Split(modifier[1], ",")
				for _, ignoredDialect := range ignoredDialects {
					if dialect == ignoredDialect {
						constraint.Ignore = true
						break
					}
				}
			}
		case "index":
			err = tbl.loadIndexConfig(dialect, tableSchema, tableName, columns, modifier[1])
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid modifier 'check.%s'", modifier[0])
		}
	}
	return nil
}

func (tbl *Table) loadColumnConfig(dialect, columnName, columnType, config string) error {
	qualifiedColumn := tbl.TableSchema + "." + tbl.TableName + "." + columnName
	if tbl.TableSchema == "" {
		qualifiedColumn = qualifiedColumn[1:]
	}
	modifiers, _, err := tokenizeModifiers(config)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
	}
	var column Column
	if n := tbl.CachedColumnPosition(columnName); n >= 0 {
		column = tbl.Columns[n]
		defer func() { tbl.Columns[n] = column }()
	} else {
		column = Column{
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			ColumnName:  columnName,
			ColumnType:  columnType,
		}
		defer func() { tbl.AppendColumn(column) }()
	}
	for _, modifier := range modifiers {
		modifierName := modifier[0]
		if i := strings.IndexByte(modifierName, ':'); i >= 0 {
			modifierDialect := modifierName[:i]
			modifierName = modifierName[i+1:]
			if modifierDialect != dialect {
				continue
			}
		}
		switch modifierName {
		case "type":
			column.ColumnType = modifier[1]
		case "auto_increment":
			if dialect == sq.DialectMySQL {
				column.IsAutoincrement = true
			}
		case "autoincrement":
			if dialect == sq.DialectSQLite {
				column.IsAutoincrement = true
			}
		case "identity":
			if dialect == sq.DialectPostgres {
				column.Identity = BY_DEFAULT_AS_IDENTITY
			}
		case "alwaysidentity":
			if dialect == sq.DialectPostgres {
				column.Identity = ALWAYS_AS_IDENTITY
			}
		case "notnull":
			column.IsNotNull = true
		case "onupdatecurrenttimestamp":
			if dialect == sq.DialectMySQL {
				column.OnUpdateCurrentTimestamp = true
			}
		case "expr":
			column.GeneratedExpr = modifier[1]
			column.GeneratedExprStored = false
			if dialect == sq.DialectPostgres {
				column.GeneratedExprStored = true
			}
		case "storedexpr":
			column.GeneratedExpr = modifier[1]
			column.GeneratedExprStored = true
		case "collate":
			column.CollationName = modifier[1]
		case "default":
			if needsExpressionBrackets(modifier[1]) && dialect != sq.DialectPostgres {
				column.ColumnDefault = "(" + modifier[1] + ")"
			} else {
				column.ColumnDefault = modifier[1]
			}
		case "primarykey":
			err = tbl.loadConstraintConfig(dialect, PRIMARY_KEY, column.TableSchema, column.TableName, []string{column.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "references":
			err = tbl.loadConstraintConfig(dialect, "references", column.TableSchema, column.TableName, []string{column.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "foreignkey":
			err = tbl.loadConstraintConfig(dialect, FOREIGN_KEY, column.TableSchema, column.TableName, []string{column.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "unique":
			err = tbl.loadConstraintConfig(dialect, UNIQUE, column.TableSchema, column.TableName, []string{column.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "index":
			err = tbl.loadIndexConfig(dialect, column.TableSchema, column.TableName, []string{column.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "ignore":
			if modifier[1] == "" {
				column.Ignore = true
			} else {
				ignoredDialects := strings.Split(modifier[1], ",")
				for _, ignoredDialect := range ignoredDialects {
					if dialect == ignoredDialect {
						column.Ignore = true
						break
					}
				}
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedColumn, modifier[0])
		}
	}
	return nil
}

func (tbl *Table) loadTableConfig(dialect, qualifiedTable, tableModifiers string) error {
	modifiers, _, err := tokenizeModifiers(tableModifiers)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
	}
	for _, modifier := range modifiers {
		modifierName := modifier[0]
		if i := strings.IndexByte(modifierName, ':'); i >= 0 {
			modifierDialect := modifierName[:i]
			modifierName = modifierName[i+1:]
			if modifierDialect != dialect {
				continue
			}
		}
		switch modifierName {
		case "virtual":
			if dialect == sq.DialectSQLite {
				virtualTable, submodifiers, _, err := tokenizeValue(modifier[1])
				if err != nil {
					return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
				}
				tbl.VirtualTable = virtualTable
				for _, submodifier := range submodifiers {
					virtualTableArg := submodifier[0]
					if submodifier[1] != "" {
						virtualTableArg += "=" + submodifier[1]
					}
					tbl.VirtualTableArgs = append(tbl.VirtualTableArgs, virtualTableArg)
				}
			}
		case "primarykey":
			err = tbl.loadConstraintConfig(dialect, PRIMARY_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "references":
			err = tbl.loadConstraintConfig(dialect, "references", tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "unique":
			err = tbl.loadConstraintConfig(dialect, UNIQUE, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "index":
			err = tbl.loadIndexConfig(dialect, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "ignore":
			if modifier[1] == "" {
				tbl.Ignore = true
			} else {
				ignoredDialects := strings.Split(modifier[1], ",")
				for _, ignoredDialect := range ignoredDialects {
					if dialect == ignoredDialect {
						tbl.Ignore = true
						break
					}
				}
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedTable, modifier[0])
		}
	}
	return nil
}

type CreateTableCommand struct {
	CreateIfNotExists  bool
	IncludeConstraints bool
	Table              Table
	CreateIndexCmds    []CreateIndexCommand // mysql-only
	Ignore             bool
}

func (cmd CreateTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if cmd.Ignore {
		return nil
	}
	if cmd.Table.TableName == "" {
		return fmt.Errorf("CREATE TABLE: table has no name")
	}
	if cmd.Table.VirtualTable != "" {
		if dialect != sq.DialectSQLite {
			return fmt.Errorf("CREATE TABLE: only SQLite has VIRTUAL TABLE support (table=%s)", cmd.Table.TableName)
		}
		buf.WriteString("CREATE VIRTUAL TABLE ")
	} else {
		buf.WriteString("CREATE TABLE ")
	}
	if cmd.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	if cmd.Table.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Table.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Table.TableName))
	if cmd.Table.VirtualTable != "" {
		buf.WriteString(" USING " + cmd.Table.VirtualTable)
	}
	buf.WriteString(" (")
	var columnWritten bool
	for i, column := range cmd.Table.Columns {
		if column.Ignore {
			continue
		}
		if cmd.Table.VirtualTable != "" {
			continue
		}
		if !columnWritten {
			columnWritten = true
			buf.WriteString("\n    ")
		} else {
			buf.WriteString("\n    ,")
		}
		err := writeColumnDefinition(dialect, buf, column)
		if err != nil {
			return fmt.Errorf("column #%d: %w", i+1, err)
		}
	}
	if len(cmd.Table.VirtualTableArgs) > 0 && cmd.Table.VirtualTable == "" {
		return fmt.Errorf("virtual table arguments present without a virtual table module")
	}
	if cmd.Table.VirtualTable != "" && dialect == sq.DialectSQLite && len(cmd.Table.VirtualTableArgs) > 0 {
		for _, arg := range cmd.Table.VirtualTableArgs {
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
	if cmd.IncludeConstraints {
		for i, constraint := range cmd.Table.Constraints {
			if constraint.Ignore {
				continue
			}
			if dialect == sq.DialectSQLite && constraint.ConstraintType == PRIMARY_KEY && len(constraint.Columns) == 1 {
				// SQLite PRIMARY KEY is always be defined inline with the column,
				// so we don't have to do it here.
				continue
			}
			if dialect != sq.DialectSQLite && constraint.ConstraintType == FOREIGN_KEY {
				// FOREIGN KEYs are always defined after all tables have been
				// created, to avoid referencing non-yet-created tables. SQLite
				// is the exception because constraints cannot be defined
				// outside of CREATE TABLE. However, SQLite foreign keys can be
				// created even if the referencing tables do not yet exist, so
				// it's not an issue.
				// http://sqlite.1065341.n5.nabble.com/Circular-foreign-keys-td14977.html
				continue
			}
			if !newlineWritten {
				buf.WriteString("\n")
				newlineWritten = true
			}
			if constraint.ConstraintName == "" || (dialect == sq.DialectMySQL && constraint.ConstraintType == PRIMARY_KEY) {
				buf.WriteString("\n    ,")
			} else {
				buf.WriteString("\n    ,CONSTRAINT " + sq.QuoteIdentifier(dialect, constraint.ConstraintName) + " ")
			}
			err := writeConstraintDefinition(dialect, buf, constraint)
			if err != nil {
				return fmt.Errorf("constraint #%d: %w", i+1, err)
			}
		}
	}
	if len(cmd.CreateIndexCmds) > 0 {
		if dialect != sq.DialectMySQL {
			return fmt.Errorf("%s does not allow defining indexes inside CREATE TABLE", dialect)
		}
		for i, createIndexCmd := range cmd.CreateIndexCmds {
			if createIndexCmd.Ignore {
				continue
			}
			if !newlineWritten {
				buf.WriteString("\n")
				newlineWritten = true
			}
			buf.WriteString("\n    ,")
			err := createIndexCmd.AppendSQL(dialect, buf, args, params, nil)
			if err != nil {
				return fmt.Errorf("index #%d: %w", i+1, err)
			}
		}
	}
	buf.WriteString("\n)")
	return nil
}

type AlterTableCommand struct {
	AlterIfExists       bool
	TableSchema         string
	TableName           string
	AddColumnCmds       []AddColumnCommand
	AlterColumnCmds     []AlterColumnCommand
	DropColumnCmds      []DropColumnCommand
	AddConstraintCmds   []AddConstraintCommand
	AlterConstraintCmds []AlterConstraintCommand
	DropConstraintCmds  []DropConstraintCommand
	CreateIndexCmds     []CreateIndexCommand // mysql-only
	DropIndexCmds       []DropIndexCommand   // mysql-only
	Ignore              bool
}

func (cmd AlterTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if cmd.Ignore {
		return nil
	}
	buf.WriteString("ALTER TABLE ")
	if cmd.AlterIfExists {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support ALTER TABLE IF EXISTS", dialect)
		}
		buf.WriteString("IF EXISTS ")
	}
	if cmd.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableName))
	if dialect == sq.DialectSQLite {
		columnCmdCount := len(cmd.AddColumnCmds) + len(cmd.AlterColumnCmds) + len(cmd.DropColumnCmds)
		indexCmdCount := len(cmd.CreateIndexCmds) + len(cmd.DropIndexCmds)
		if columnCmdCount > 1 {
			return fmt.Errorf("sqlite ALTER TABLE only supports one column modification")
		}
		if indexCmdCount > 0 {
			return fmt.Errorf("sqlite ALTER TABLE does not support indexes")
		}
	} else if dialect == sq.DialectPostgres {
		indexCmdCount := len(cmd.DropIndexCmds) + len(cmd.CreateIndexCmds)
		if indexCmdCount > 0 {
			return fmt.Errorf("postgres ALTER TABLE does not support indexes")
		}
	}
	var firstLineWritten bool
	writeNewLine := func() {
		if !firstLineWritten {
			firstLineWritten = true
			buf.WriteString("\n    ")
		} else {
			buf.WriteString("\n    ,")
		}
	}
	for _, addColumnCmd := range cmd.AddColumnCmds {
		writeNewLine()
		err := addColumnCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE ADD COLUMN %s: %w", addColumnCmd.Column.ColumnName, err)
		}
	}
	for _, alterColumnCmd := range cmd.AlterColumnCmds {
		writeNewLine()
		err := alterColumnCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE ALTER COLUMN %s: %w", alterColumnCmd.Column.ColumnName, err)
		}
	}
	for _, dropColumnCmd := range cmd.DropColumnCmds {
		writeNewLine()
		err := dropColumnCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE DROP COLUMN %s: %w", dropColumnCmd.ColumnName, err)
		}
	}
	// DROP CONSTRAINT comes before ADD CONSTRAINT because that's the only way
	// MySQL can rename constraints: by dropping and re-adding them in the same
	// command.
	for _, dropConstraintCmd := range cmd.DropConstraintCmds {
		writeNewLine()
		err := dropConstraintCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE DROP CONSTRAINT %s: %w", dropConstraintCmd.ConstraintName, err)
		}
	}
	for _, addConstraintCmd := range cmd.AddConstraintCmds {
		writeNewLine()
		err := addConstraintCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE ADD CONSTRAINT %s: %w", addConstraintCmd.Constraint.ConstraintName, err)
		}
	}
	for _, alterConstraintCmd := range cmd.AlterConstraintCmds {
		writeNewLine()
		err := alterConstraintCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE ALTER CONSTRAINT %s: %w", alterConstraintCmd.ConstraintName, err)
		}
	}
	for _, createIndexCmd := range cmd.CreateIndexCmds {
		writeNewLine()
		if dialect == sq.DialectMySQL {
			buf.WriteString("ADD ")
		}
		err := createIndexCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE INDEX %s: %w", createIndexCmd.Index.IndexName, err)
		}
	}
	for _, dropIndexCmd := range cmd.DropIndexCmds {
		writeNewLine()
		err := dropIndexCmd.AppendSQL(dialect, buf, args, params, nil)
		if err != nil {
			return fmt.Errorf("ALTER TABLE DROP INDEX %s: %w", dropIndexCmd.IndexName, err)
		}
	}
	buf.WriteString("\n")
	return nil
}

func decomposeAlterTableCommandSQLite2(alterTableCmd AlterTableCommand) []AlterTableCommand {
	alterTableCmds := make([]AlterTableCommand, 0, len(alterTableCmd.AddColumnCmds)+len(alterTableCmd.DropColumnCmds))
	for _, addColumnCmd := range alterTableCmd.AddColumnCmds {
		alterTableCmds = append(alterTableCmds, AlterTableCommand{
			TableSchema:   alterTableCmd.TableSchema,
			TableName:     alterTableCmd.TableName,
			AddColumnCmds: []AddColumnCommand{addColumnCmd},
		})
	}
	for _, dropColumnCmd := range alterTableCmd.DropColumnCmds {
		alterTableCmds = append(alterTableCmds, AlterTableCommand{
			TableSchema:    alterTableCmd.TableSchema,
			TableName:      alterTableCmd.TableName,
			DropColumnCmds: []DropColumnCommand{dropColumnCmd},
		})
	}
	return alterTableCmds
}

func decomposeAlterTableCommandSQLite(alterTableCmd *AlterTableCommand) []Command {
	alterTableCmds := make([]AlterTableCommand, 0, len(alterTableCmd.AddColumnCmds)+len(alterTableCmd.DropColumnCmds))
	for _, addColumnCmd := range alterTableCmd.AddColumnCmds {
		alterTableCmds = append(alterTableCmds, AlterTableCommand{
			TableSchema:   alterTableCmd.TableSchema,
			TableName:     alterTableCmd.TableName,
			AddColumnCmds: []AddColumnCommand{addColumnCmd},
		})
	}
	for _, dropColumnCmd := range alterTableCmd.DropColumnCmds {
		alterTableCmds = append(alterTableCmds, AlterTableCommand{
			TableSchema:    alterTableCmd.TableSchema,
			TableName:      alterTableCmd.TableName,
			DropColumnCmds: []DropColumnCommand{dropColumnCmd},
		})
	}
	cmds := make([]Command, len(alterTableCmds))
	for i := range alterTableCmds {
		cmds[i] = &alterTableCmds[i]
	}
	return cmds
}

type RenameTableCommand struct {
	AlterIfExists   bool
	TableSchemas    []string
	TableNames      []string
	RenameToSchemas []string
	RenameToNames   []string
	Ignore          bool
}

func (cmd *RenameTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if cmd.Ignore {
		return nil
	}
	if len(cmd.TableNames) > 1 && dialect != sq.DialectMySQL {
		return fmt.Errorf("%s does not support renaming multiple tables in one command", dialect)
	}
	if dialect == sq.DialectMySQL {
		buf.WriteString("RENAME TABLE ")
		for i, tableSchema := range cmd.TableSchemas {
			if i > 0 {
				buf.WriteString(", ")
			}
			if tableSchema != "" {
				buf.WriteString(sq.QuoteIdentifier(dialect, tableSchema) + ".")
			}
			buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableNames[i]) + " RENAME TO ")
			if renameToSchema := cmd.RenameToSchemas[i]; renameToSchema != "" {
				buf.WriteString(sq.QuoteIdentifier(dialect, renameToSchema) + ".")
			}
			buf.WriteString(sq.QuoteIdentifier(dialect, cmd.RenameToNames[i]))
		}
	} else {
		buf.WriteString("ALTER TABLE ")
		if cmd.AlterIfExists {
			if dialect != sq.DialectPostgres {
				return fmt.Errorf("%s does not support ALTER TABLE IF EXISTS", dialect)
			}
			buf.WriteString("IF EXISTS ")
		}
		if tableSchema := cmd.TableSchemas[0]; tableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, tableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableNames[0]) + " RENAME TO " + sq.QuoteIdentifier(dialect, cmd.RenameToNames[0]))
	}
	return nil
}

type DropTableCommand struct {
	DropIfExists bool
	TableSchemas []string
	TableNames   []string
	DropCascade  bool
	Ignore       bool
}

func (cmd DropTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if cmd.Ignore {
		return nil
	}
	buf.WriteString("DROP TABLE ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS")
		if len(cmd.TableNames) == 1 {
			buf.WriteString(" ")
		}
	}
	for i, tableName := range cmd.TableNames {
		if i > 0 {
			buf.WriteString("\n    ,")
		} else if len(cmd.TableNames) > 1 {
			buf.WriteString("\n    ")
		}
		tableSchema := cmd.TableSchemas[i]
		if tableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, tableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, tableName))
	}
	if cmd.DropCascade {
		if dialect != sq.DialectPostgres && dialect != sq.DialectMySQL {
			return fmt.Errorf("%s does not support DROP TABLE ... CASCADE", dialect)
		}
		if len(cmd.TableNames) > 1 {
			buf.WriteString("\n")
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString("CASCADE")
	}
	return nil
}

func decomposeDropTableCommandSQLite(dropTableCmd *DropTableCommand) []Command {
	dropTableCmds := make([]DropTableCommand, 0, len(dropTableCmd.TableNames))
	for i, tableName := range dropTableCmd.TableNames {
		dropTableCmds = append(dropTableCmds, DropTableCommand{
			DropIfExists: dropTableCmd.DropIfExists,
			TableSchemas: []string{dropTableCmd.TableSchemas[i]},
			TableNames:   []string{tableName},
		})
	}
	cmds := make([]Command, len(dropTableCmd.TableNames))
	for i := range dropTableCmds {
		cmds[i] = &dropTableCmds[i]
	}
	return cmds
}

func decomposeDropTableCommandSQLite2(dropTableCmd DropTableCommand) []DropTableCommand {
	dropTableCmds := make([]DropTableCommand, 0, len(dropTableCmd.TableNames))
	for i, tableName := range dropTableCmd.TableNames {
		dropTableCmds = append(dropTableCmds, DropTableCommand{
			DropIfExists: dropTableCmd.DropIfExists,
			TableSchemas: []string{dropTableCmd.TableSchemas[i]},
			TableNames:   []string{tableName},
		})
	}
	return dropTableCmds
}

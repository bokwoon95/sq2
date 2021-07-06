package ddl3

import "github.com/bokwoon95/sq"

type Catalog struct {
	Dialect         string
	CatalogName     string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
	Schemas         []Schema
	schemasCache    map[string]int
}

func (c *Catalog) CachedSchemaPosition(schemaName string) (schemaPosition int) {
	schemaPosition, ok := c.schemasCache[schemaName]
	if !ok || schemaPosition < 0 || schemaPosition >= len(c.Schemas) {
		delete(c.schemasCache, schemaName)
		return -1
	}
	if c.Schemas[schemaPosition].SchemaName != schemaName {
		delete(c.schemasCache, schemaName)
		return -1
	}
	return schemaPosition
}

func (c *Catalog) AppendSchema(schema Schema) (schemaPosition int) {
	c.Schemas = append(c.Schemas, schema)
	if c.schemasCache == nil {
		c.schemasCache = make(map[string]int)
	}
	schemaPosition = len(c.Schemas) - 1
	c.schemasCache[schema.SchemaName] = schemaPosition
	return schemaPosition
}

func (c *Catalog) RefreshSchemasCache() {
	for i, schema := range c.Schemas {
		if c.schemasCache == nil {
			c.schemasCache = make(map[string]int)
		}
		c.schemasCache[schema.SchemaName] = i
	}
}

func NewCatalog(dialect string) Catalog {
	return Catalog{Dialect: dialect}
}

func (c *Catalog) LoadDB(db sq.Queryer) error {
	return nil
}

func (c *Catalog) LoadTables(tables ...sq.SchemaTable) error {
	return nil
}

func (c *Catalog) loadTable(table sq.SchemaTable) error {
	return nil
}

func (c *Catalog) LoadViews(views ...View) error {
	return nil
}

func (c *Catalog) loadView(view View) error {
	return nil
}

func (c *Catalog) LoadFunctions() error {
	return nil
}

func (c *Catalog) loadFunction() error {
	return nil
}

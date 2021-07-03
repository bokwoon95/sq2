package ddl2

type Metadata struct {
	Dialect         string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
	Schemas         []Schema
	schemasCache    map[string]int
}

func NewMetadata(dialect string) Metadata {
	return Metadata{Dialect: dialect}
}

func (m *Metadata) LoadDB() error {
	return nil
}

func (m *Metadata) CachedSchemaIndex(schemaName string) (schemaIndex int) {
	if m == nil {
		return -1
	}
	schemaIndex, ok := m.schemasCache[schemaName]
	if !ok || schemaIndex < 0 || schemaIndex >= len(m.Schemas) {
		delete(m.schemasCache, schemaName)
		return -1
	}
	if m.Schemas[schemaIndex].SchemaName != schemaName {
		delete(m.schemasCache, schemaName)
		return -1
	}
	return schemaIndex
}

func (m *Metadata) AppendSchema(schema Schema) (schemaIndex int) {
	if m == nil {
		return -1
	}
	m.Schemas = append(m.Schemas, schema)
	if m.schemasCache == nil {
		m.schemasCache = make(map[string]int)
	}
	schemaIndex = len(m.Schemas) - 1
	m.schemasCache[schema.SchemaName] = schemaIndex
	return schemaIndex
}

func (m *Metadata) RefreshSchemaCache() {
	if m == nil {
		return
	}
	for i, schema := range m.Schemas {
		if m.schemasCache == nil {
			m.schemasCache = make(map[string]int)
		}
		m.schemasCache[schema.SchemaName] = i
	}
}

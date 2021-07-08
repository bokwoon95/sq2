package ddl3

type CatalogDiff struct {
	SchemaDiffs      []SchemaDiff
	schemaDiffsCache map[string]int // 8 bytes
}

func DiffCatalog(gotCatalog, wantCatalog Catalog) (CatalogDiff, error) {
	var catalogDiff CatalogDiff
	return catalogDiff, nil
}

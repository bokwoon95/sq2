package ddl3

type Changeset struct {
}

func DiffCatalog(gotCatalog, wantCatalog Catalog) (Changeset, error) {
	var set Changeset
	return set, nil
}

func (set Changeset) Commands() Commands {
	return Commands{}
}

package ddl3

import "github.com/bokwoon95/sq"

type View struct {
	ViewSchema string
	ViewName   string
	Contents   string
}

type DDLView interface {
	sq.SchemaTable
	DDL(dialect string, v *V) sq.Query
}

type V struct {
	CreateOrReplace bool
	IsMaterialized  bool
	IsRecursive     bool
}

// NOTE: I can eventually add a v.Version(versionID string), in order to
// support versioned Views/Functions/Triggers. The main issue with updating to
// a new version is that you have to drop the existing version, which is NOT
// SAFE if there are other applications or nodes that are communicating with
// the DB. DiffCatalog can generate those changes anyway, and it is up to the
// user to remove those DROP VIEW commands themselves. Alternatively, they can
// reach into the Catalog and change the View/Function/Trigger back to
// unversioned (setting VersionID to an empty string) so that DiffCatalog never
// generates those changes in the first place.

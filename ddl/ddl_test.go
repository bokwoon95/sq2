package ddl

import (
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_generateName(t *testing.T) {
	t.Run("long identifier names over 63 bytes get trimmed", func(t *testing.T) {
		gotName := generateName(PRIMARY_KEY, "pm_url_role_capability",
			"site_id",
			"urlpath",
			"plugin",
			"role",
			"capability",
		)
		wantName := "pm_url_role_capability_site_id_urlpath_plugin_role_capabil_pkey"
		if diff := testutil.Diff(gotName, wantName); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
	})
}

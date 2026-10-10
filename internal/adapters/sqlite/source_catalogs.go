package sqlite

import (
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

var _ app.CatalogTx = (*attentionTx)(nil)

func (t *attentionTx) LoadSourceCatalog(platform, owner, mode string) ([]app.SourceCollection, error) {
	rows := []app.SourceCollection{}
	err := t.load("SELECT data FROM attention_source_catalogs WHERE platform=? AND owner_id=? AND access_mode=?", "source_catalog", platform+":"+owner, []any{platform, owner, mode}, &rows)
	return rows, err
}
func (t *attentionTx) SaveSourceCatalog(platform, owner, mode string, rows []app.SourceCollection) error {
	if platform == "" || owner == "" || (mode != "public" && mode != "browser_selected") {
		return conflict("invalid catalog identity")
	}
	for _, row := range rows {
		if row.OwnerID != owner || row.AccessMode != mode || row.ExternalID == "" {
			return conflict("catalog row identity differs from catalog scope")
		}
	}
	data, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_source_catalogs(platform,owner_id,access_mode,data) VALUES(?,?,?,?) ON CONFLICT(platform,owner_id,access_mode) DO UPDATE SET data=excluded.data", platform, owner, mode, string(data))
	return storageError(err)
}

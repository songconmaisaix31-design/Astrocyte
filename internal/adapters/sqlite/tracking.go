package sqlite

import (
	"encoding/json"
	"errors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

var _ app.TrackingTx = (*attentionTx)(nil)

func (t *attentionTx) ListTrackingSources() ([]app.TrackingSource, error) {
	return listJSON[app.TrackingSource](t, "SELECT data FROM attention_tracking_sources ORDER BY platform,source_kind,id")
}
func (t *attentionTx) LoadTrackingSource(id string) (app.TrackingSource, error) {
	var row app.TrackingSource
	err := t.load("SELECT data FROM attention_tracking_sources WHERE id=?", "tracking_source", id, []any{id}, &row)
	return row, err
}
func (t *attentionTx) SaveTrackingSource(row app.TrackingSource, expected int) error {
	if row.Version != expected+1 {
		return conflict("tracking source version must advance exactly once")
	}
	if expected > 0 {
		old, err := t.LoadTrackingSource(row.ID)
		if err != nil {
			return err
		}
		if old.Platform != row.Platform || old.SourceKind != row.SourceKind || old.ExternalID != row.ExternalID || old.OwnerID != row.OwnerID {
			return conflict("public source identity cannot change")
		}
	}
	data, err := encode(row)
	if err != nil {
		return err
	}
	return t.saveHead("attention_tracking_sources", row.ID, expected, data, []string{"platform", "source_kind", "external_id"}, []any{row.Platform, row.SourceKind, row.ExternalID})
}
func (t *attentionTx) ListSourceItems(sourceID string) ([]app.SourceItem, error) {
	rows, err := t.conn.QueryContext(t.ctx, "SELECT data FROM attention_source_items WHERE source_id=? ORDER BY external_id", sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []app.SourceItem{}
	for rows.Next() {
		var data string
		var item app.SourceItem
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (t *attentionTx) SaveSourceItem(item app.SourceItem) error {
	if item.SourceID == "" || item.ExternalID == "" || item.Revision < 1 || item.Metadata.ExternalID != item.ExternalID {
		return conflict("invalid source item identity")
	}
	var old app.SourceItem
	err := t.load("SELECT data FROM attention_source_items WHERE source_id=? AND external_id=?", "source_item", item.ExternalID, []any{item.SourceID, item.ExternalID}, &old)
	if err == nil {
		if item.Revision < old.Revision || item.Revision > old.Revision+1 || (item.Revision == old.Revision && !equalJSON(old.Metadata, item.Metadata)) || (item.Revision == old.Revision+1 && equalJSON(old.Metadata, item.Metadata)) {
			return conflict("metadata revision does not match actual change")
		}
	} else {
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != apierrors.NotFound {
			return err
		}
		if item.Revision != 1 {
			return conflict("initial metadata revision must be one")
		}
	}
	if item.Recommendation != nil && item.Recommendation.MetadataRevision != item.Revision {
		return conflict("recommendation must bind current metadata revision")
	}
	data, err := encode(item)
	if err != nil {
		return err
	}
	_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_source_items(source_id,external_id,revision,data) VALUES(?,?,?,?) ON CONFLICT(source_id,external_id) DO UPDATE SET revision=excluded.revision,data=excluded.data", item.SourceID, item.ExternalID, item.Revision, data)
	return storageError(err)
}

package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

var _ app.TasteProfileTx = (*attentionTx)(nil)

func (t *attentionTx) LoadTasteProfile() (app.TasteProfileDetail, error) {
	result := app.TasteProfileDetail{SchemaVersion: 1}
	var version int
	var data string
	err := t.conn.QueryRowContext(t.ctx, "SELECT version,data FROM attention_taste_profiles WHERE id='attention'").Scan(&version, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	profile, err := decodeTasteProfile(data, "attention", version)
	if err != nil {
		return result, err
	}
	fixed, err := t.LoadTasteProfileRevision(profile.ID, profile.Version)
	if err != nil {
		return result, err
	}
	if !equalJSON(profile, fixed) {
		return result, errors.New("taste profile head differs from its immutable revision")
	}
	result.Configured, result.Profile = true, &profile
	return result, nil
}

func (t *attentionTx) LoadTasteProfileRevision(id string, version int) (app.TasteProfile, error) {
	var data string
	err := t.conn.QueryRowContext(t.ctx, "SELECT data FROM attention_taste_profile_revisions WHERE profile_id=? AND version=?", id, version).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return app.TasteProfile{}, apierrors.NewNotFound("taste_profile_revision", id)
	}
	if err != nil {
		return app.TasteProfile{}, err
	}
	return decodeTasteProfile(data, id, version)
}

func decodeTasteProfile(data, id string, version int) (app.TasteProfile, error) {
	var profile app.TasteProfile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return profile, err
	}
	if err := app.ValidateTasteProfile(profile); err != nil {
		return app.TasteProfile{}, err
	}
	if profile.ID != id || profile.Version != version {
		return app.TasteProfile{}, errors.New("taste profile differs from its stored immutable identity")
	}
	return profile, nil
}

func (t *attentionTx) SaveTasteProfile(profile app.TasteProfile, expected int) error {
	if err := app.ValidateTasteProfile(profile); err != nil {
		return err
	}
	if err := requireVersion(profile.Version, expected); err != nil {
		return err
	}
	current, err := t.LoadTasteProfile()
	if err != nil {
		return err
	}
	if (expected == 0 && current.Configured) || (expected > 0 && (!current.Configured || current.Profile.Version != expected)) {
		return conflict("taste profile version changed")
	}
	data, err := encode(profile)
	if err != nil {
		return err
	}
	if err := t.saveHead("attention_taste_profiles", profile.ID, expected, data, nil, nil); err != nil {
		return err
	}
	// 只插入下一条修订；不存在替换或更新历史修订的代码路径。
	_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_taste_profile_revisions(profile_id,version,data) VALUES(?,?,?)", profile.ID, profile.Version, data)
	return storageError(err)
}

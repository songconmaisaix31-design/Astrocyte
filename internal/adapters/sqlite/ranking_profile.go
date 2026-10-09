package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

var _ app.RankingProfileTx = (*attentionTx)(nil)

func (t *attentionTx) LoadRankingProfile() (app.RankingProfileDetail, error) {
	v := app.RankingProfileDetail{SchemaVersion: 1, Versions: []app.RankingProfile{}}
	var data string
	err := t.conn.QueryRowContext(t.ctx, "SELECT data FROM attention_ranking_profiles WHERE id='attention'").Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return v, nil
	}
	if err != nil {
		return v, err
	}
	err = json.Unmarshal([]byte(data), &v)
	return v, err
}

func (t *attentionTx) SaveRankingProfile(v app.RankingProfileDetail, expected int) error {
	if !v.Configured || v.Profile == nil || v.Profile.ID != "attention" {
		return conflict("ranking profile requires configured singleton attention")
	}
	if err := requireVersion(v.Profile.Version, expected); err != nil {
		return err
	}
	if len(v.Versions) != v.Profile.Version || !equalJSON(*v.Profile, v.Versions[len(v.Versions)-1]) {
		return conflict("ranking profile history must end at the current head")
	}
	old, err := t.LoadRankingProfile()
	if err != nil {
		return err
	}
	if len(old.Versions) > len(v.Versions) {
		return conflict("ranking profile history cannot be removed")
	}
	for i, version := range v.Versions {
		if version.ID != "attention" || version.Version != i+1 {
			return conflict("ranking profile versions must be sequential")
		}
		if i < len(old.Versions) && !equalJSON(old.Versions[i], version) {
			return conflict("ranking profile revision cannot be overwritten")
		}
	}
	data, err := encode(v)
	if err != nil {
		return err
	}
	if err = t.saveHead("attention_ranking_profiles", "attention", expected, data, nil, nil); err != nil {
		return err
	}
	revision, err := encode(*v.Profile)
	if err != nil {
		return err
	}
	_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_ranking_profile_revisions(profile_id,version,data) VALUES ('attention',?,?)", v.Profile.Version, revision)
	return storageError(err)
}

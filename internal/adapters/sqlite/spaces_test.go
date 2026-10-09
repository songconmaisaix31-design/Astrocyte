package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestHumanClassificationAndFixedReferenceSurviveRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	r := NewAttentionRepository(db)
	material := materialFixture(t)
	material.Material.DomainIDs = []string{"daily-study"}
	domain := app.MaterialDomain{ID: "daily-study", Version: 1, Title: "日常学习"}
	space := app.ProjectSpace{ID: "project", Version: 1, Title: "当前研究", MaterialRefs: []app.SourceRef{{MaterialID: material.Material.ID, Revision: 1, Locator: material.Revisions[0].SourceLocator}}}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		if err := tx.SaveMaterialDomain(domain, 0); err != nil {
			return err
		}
		if err := tx.SaveMaterial(material, 0); err != nil {
			return err
		}
		return tx.SaveProjectSpace(space, 0)
	}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db = openAttentionDB(t, path)
	r = NewAttentionRepository(db)
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		m, err := tx.LoadMaterial("m1")
		if err != nil {
			return err
		}
		s, err := tx.LoadProjectSpace("project")
		if err != nil {
			return err
		}
		d, err := tx.LoadMaterialDomain("daily-study")
		if err != nil {
			return err
		}
		if d.Title != "日常学习" || len(m.Material.DomainIDs) != 1 || len(m.Revisions) != 1 || s.MaterialRefs[0].Revision != 1 {
			t.Fatal("classification/reference moved or copied original")
		}
		s.Version++
		s.MaterialRefs = []app.SourceRef{}
		return tx.SaveProjectSpace(s, 1)
	}); err != nil {
		t.Fatal(err)
	}
	// A stale writer cannot reinstate a removed reference.
	space.Version = 2
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error { return tx.SaveProjectSpace(space, 1) }); err == nil {
		t.Fatal("stale reference reappeared")
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		m, err := tx.LoadMaterial("m1")
		if err != nil {
			return err
		}
		s, err := tx.LoadProjectSpace("project")
		if err != nil {
			return err
		}
		domains, err := tx.ListMaterialDomains()
		if err != nil {
			return err
		}
		spaces, err := tx.ListProjectSpaces()
		if err != nil {
			return err
		}
		if len(m.Material.DomainIDs) != 1 || len(m.Revisions) != 1 || len(s.MaterialRefs) != 0 || len(domains) != 1 || len(spaces) != 1 {
			t.Fatal("removal lost source or classification")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.conn.QueryRow("SELECT count(*) FROM attention_material_revisions").Scan(&count); err != nil || count != 1 {
		t.Fatal("copied original", count, err)
	}
}

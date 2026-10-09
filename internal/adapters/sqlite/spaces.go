package sqlite

import "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"

// Human classifications and project references are heads in Attention. Their
// persistence does not grant access, move originals or write Workspace tables.
func (t *attentionTx) ListMaterialDomains() ([]app.MaterialDomain, error) {
	return listJSON[app.MaterialDomain](t, "SELECT data FROM attention_domains ORDER BY id")
}

func (t *attentionTx) LoadMaterialDomain(id string) (app.MaterialDomain, error) {
	var v app.MaterialDomain
	err := t.load("SELECT data FROM attention_domains WHERE id=?", "material domain", id, []any{id}, &v)
	return v, err
}

func (t *attentionTx) SaveMaterialDomain(v app.MaterialDomain, expected int) error {
	if err := requireVersion(v.Version, expected); err != nil {
		return err
	}
	data, err := encode(v)
	if err != nil {
		return err
	}
	return t.saveHead("attention_domains", v.ID, expected, data, nil, nil)
}

func (t *attentionTx) ListProjectSpaces() ([]app.ProjectSpace, error) {
	return listJSON[app.ProjectSpace](t, "SELECT data FROM attention_project_spaces ORDER BY id")
}

func (t *attentionTx) LoadProjectSpace(id string) (app.ProjectSpace, error) {
	var v app.ProjectSpace
	err := t.load("SELECT data FROM attention_project_spaces WHERE id=?", "project space", id, []any{id}, &v)
	return v, err
}

func (t *attentionTx) SaveProjectSpace(v app.ProjectSpace, expected int) error {
	if err := requireVersion(v.Version, expected); err != nil {
		return err
	}
	data, err := encode(v)
	if err != nil {
		return err
	}
	return t.saveHead("attention_project_spaces", v.ID, expected, data, nil, nil)
}

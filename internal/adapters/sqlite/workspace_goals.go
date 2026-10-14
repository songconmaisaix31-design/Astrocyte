package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (t *workspaceTransaction) LoadGoal(id string,revision int) (g domain.Goal,err error) {
	if revision==0 { err=t.read("SELECT data FROM workspace_goal_revisions WHERE id=? ORDER BY revision DESC LIMIT 1","goal",id,&g,id) } else { err=t.read("SELECT data FROM workspace_goal_revisions WHERE id=? AND revision=?","goal revision",id,&g,id,revision) }; return
}

func (t *workspaceTransaction) SaveGoal(g domain.Goal,expected int) error {
	if g.Revision!=expected+1 { return conflict("goal revision must advance once") }
	data,err:=json.Marshal(g);if err!=nil{return err}
	var result sql.Result
	if expected==0 {result,err=t.conn.ExecContext(t.ctx,"INSERT INTO workspace_goals(id,revision) VALUES(?,?) ON CONFLICT DO NOTHING",g.ID,g.Revision)}else{result,err=t.conn.ExecContext(t.ctx,"UPDATE workspace_goals SET revision=? WHERE id=? AND revision=?",g.Revision,g.ID,expected)}
	if err=workspaceCAS(result,err);err!=nil{return err}
	_,err=t.conn.ExecContext(t.ctx,"INSERT INTO workspace_goal_revisions(id,revision,data) VALUES(?,?,?)",g.ID,g.Revision,string(data));return err
}

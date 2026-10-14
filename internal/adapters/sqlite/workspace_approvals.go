package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (t *workspaceTransaction) LoadApproval(id string) (g domain.ApprovalGrant,err error) {
	err=t.read("SELECT data FROM workspace_approvals WHERE id=?","approval",id,&g,id);return
}

func (t *workspaceTransaction) SaveApproval(g domain.ApprovalGrant,expected int) error {
	if g.Epoch!=expected+1 {return conflict("approval epoch must advance once")}
	data,err:=json.Marshal(g);if err!=nil{return err}
	var result sql.Result
	if expected==0 {result,err=t.conn.ExecContext(t.ctx,"INSERT INTO workspace_approvals(id,epoch,proposal_id,proposal_revision,data) VALUES(?,?,?,?,?)",g.ID,g.Epoch,g.ProposalID,g.ProposalRevision,string(data))}else{result,err=t.conn.ExecContext(t.ctx,"UPDATE workspace_approvals SET epoch=?,data=? WHERE id=? AND epoch=?",g.Epoch,string(data),g.ID,expected)}
	return workspaceCAS(result,err)
}

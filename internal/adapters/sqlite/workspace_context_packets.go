package sqlite

import (
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (t *workspaceTransaction) LoadContextPacket(id string) (p domain.ImmutableContextPacket,err error) {
	err=t.read("SELECT data FROM workspace_context_packets WHERE id=?","context packet",id,&p,id);return
}

func (t *workspaceTransaction) SaveContextPacket(p domain.ImmutableContextPacket) error {
	data,err:=json.Marshal(p);if err!=nil{return err}
	_,err=t.conn.ExecContext(t.ctx,"INSERT INTO workspace_context_packets(id,project_id,predecessor_id,data,object_ref) VALUES(?,?,?,?,?)",p.ID,p.ProjectID,nullable(p.PredecessorID),string(data),p.ObjectRef);return storageError(err)
}

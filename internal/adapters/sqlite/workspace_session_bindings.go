package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type storedSessionBinding struct {
	domain.SessionBinding
	PrivateNativeSessionID string `json:"private_native_session_id"`
}

func (t *workspaceTransaction) LoadNativeSession(id string) (s domain.NativeSession,err error) {
	err=t.read("SELECT data FROM local_agent_sessions WHERE id=?","native session",id,&s,id);return
}

func (t *workspaceTransaction) LoadSessionBinding(id string) (b domain.SessionBinding,err error) {
	var stored storedSessionBinding
	err=t.read("SELECT data FROM workspace_session_bindings WHERE id=?","session binding",id,&stored,id)
	b=stored.SessionBinding;b.NativeSessionID=stored.PrivateNativeSessionID;return
}

func (t *workspaceTransaction) SaveSessionBinding(b domain.SessionBinding,expected int) error {
	if b.Version!=expected+1{return conflict("binding version must advance once")}
	data,err:=json.Marshal(storedSessionBinding{SessionBinding:b,PrivateNativeSessionID:b.NativeSessionID});if err!=nil{return err};var result sql.Result
	if expected==0{result,err=t.conn.ExecContext(t.ctx,"INSERT INTO workspace_session_bindings(id,version,project_id,session_id,grant_id,packet_id,data) VALUES(?,?,?,?,?,?,?)",b.ID,b.Version,b.ProjectID,b.SessionID,b.GrantID,b.PacketID,string(data))}else{result,err=t.conn.ExecContext(t.ctx,"UPDATE workspace_session_bindings SET version=?,data=? WHERE id=? AND version=?",b.Version,string(data),b.ID,expected)}
	return workspaceCAS(result,err)
}

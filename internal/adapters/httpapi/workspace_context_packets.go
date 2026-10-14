package httpapi

import (
	"net/http"

	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func (h *handler) registerWorkspaceContext(mux *http.ServeMux,s *workspace.S2Service){
	mux.HandleFunc("POST /api/v1/context-packets",workspaceCommand(h,func(c *workspace.BuildContextPacketCommand)*workspace.WorkspaceCommandMeta{return &c.WorkspaceCommandMeta},func(r *http.Request,c workspace.BuildContextPacketCommand)(any,error){p,err:=s.BuildContextPacket(r.Context(),localCaller(r),c);return localEnvelope("context_packet",p),err}))
	mux.HandleFunc("GET /api/v1/context-packets/{id}",func(w http.ResponseWriter,r *http.Request){p,err:=s.ReadContextPacket(r.Context(),localCaller(r),r.PathValue("id"));h.attentionResult(w,r,http.StatusOK,localEnvelope("context_packet",p),err)})
	mux.HandleFunc("POST /api/v1/session-bindings",workspaceCommand(h,func(c *workspace.BindSessionCommand)*workspace.WorkspaceCommandMeta{return &c.WorkspaceCommandMeta},func(r *http.Request,c workspace.BindSessionCommand)(any,error){b,err:=s.BindSession(r.Context(),localCaller(r),c);return localEnvelope("session_binding",b),err}))
}

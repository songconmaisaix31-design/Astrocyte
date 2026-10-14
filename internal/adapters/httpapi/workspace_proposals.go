package httpapi

import (
	"net/http"
	"strconv"

	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func workspaceCommand[T any](h *handler,meta func(*T)*workspace.WorkspaceCommandMeta,call func(*http.Request,T)(any,error)) http.HandlerFunc {
	return func(w http.ResponseWriter,r *http.Request){
		body,err:=decodeAttention[T](w,r)
		if err==nil {m:=meta(&body);m.IdempotencyKey=r.Header.Get("Idempotency-Key");if m.SchemaVersion!=1||m.RequestID==""||m.ExpectedVersion<0||m.IdempotencyKey==""{err=invalid("workspace command requires schema_version, request_id, expected_version and Idempotency-Key")}}
		var result any;if err==nil{result,err=call(r,body)};h.attentionResult(w,r,http.StatusOK,result,err)
	}
}

// H 在已认证的主router中注册，替换既有三个S0占位route。
func (h *handler) registerWorkspaceS2(mux *http.ServeMux,s *workspace.S2Service) {
	if s==nil {return}
	mux.HandleFunc("GET /api/v1/proposals",func(w http.ResponseWriter,r *http.Request){p,err:=s.ListProposalsFor(r.Context(),localCaller(r));h.attentionResult(w,r,http.StatusOK,localList(p),err)})
	mux.HandleFunc("POST /api/v1/proposals",workspaceCommand(h,func(c *workspace.CreateProposalCommand)*workspace.WorkspaceCommandMeta{return &c.WorkspaceCommandMeta},func(r *http.Request,c workspace.CreateProposalCommand)(any,error){p,err:=s.CreateProposal(r.Context(),localCaller(r),c);return localEnvelope("proposal",p),err}))
	mux.HandleFunc("PUT /api/v1/proposals/{id}",workspaceCommand(h,func(c *workspace.CreateProposalCommand)*workspace.WorkspaceCommandMeta{return &c.WorkspaceCommandMeta},func(r *http.Request,c workspace.CreateProposalCommand)(any,error){p,err:=s.ReviseProposal(r.Context(),localCaller(r),r.PathValue("id"),c);return localEnvelope("proposal",p),err}))
	mux.HandleFunc("GET /api/v1/proposals/{id}",func(w http.ResponseWriter,r *http.Request){revision:=0;var err error;if raw:=r.URL.Query().Get("revision");raw!=""{revision,err=strconv.Atoi(raw);if err!=nil||revision<1{err=invalid("revision must be positive")}};var p domain.Proposal;if err==nil{p,err=s.ReadProposal(r.Context(),localCaller(r),r.PathValue("id"),revision)};h.attentionResult(w,r,http.StatusOK,localEnvelope("proposal",p),err)})
	for _,route:=range []struct{path string;status domain.ProposalStatus}{{"submit",domain.ProposalInReview},{"decline",domain.ProposalDeclined}}{
		mux.HandleFunc("POST /api/v1/proposals/{id}/"+route.path,workspaceCommand(h,func(c *workspace.WorkspaceCommandMeta)*workspace.WorkspaceCommandMeta{return c},func(r *http.Request,c workspace.WorkspaceCommandMeta)(any,error){p,err:=s.TransitionProposal(r.Context(),localCaller(r),r.PathValue("id"),c,route.status);return localEnvelope("proposal",p),err}))
	}
	mux.HandleFunc("POST /api/v1/proposals/{id}/approvals",workspaceCommand(h,func(c *workspace.ApproveProposalCommand)*workspace.WorkspaceCommandMeta{return &c.WorkspaceCommandMeta},func(r *http.Request,c workspace.ApproveProposalCommand)(any,error){g,err:=s.ApproveProposal(r.Context(),localCaller(r),r.PathValue("id"),c);return localEnvelope("approval",g),err}))
	mux.HandleFunc("GET /api/v1/approvals/{id}",func(w http.ResponseWriter,r *http.Request){g,err:=s.ReadApproval(r.Context(),localCaller(r),r.PathValue("id"));h.attentionResult(w,r,http.StatusOK,localEnvelope("approval",g),err)})
	for _,route:=range []struct{path,status string}{{"revoke","revoked"},{"complete","completed"},{"expire","expired"}}{
		mux.HandleFunc("POST /api/v1/approvals/{id}/"+route.path,workspaceCommand(h,func(c *workspace.ChangeApprovalCommand)*workspace.WorkspaceCommandMeta{return &c.WorkspaceCommandMeta},func(r *http.Request,c workspace.ChangeApprovalCommand)(any,error){g,err:=s.ChangeApproval(r.Context(),localCaller(r),r.PathValue("id"),c,route.status);return localEnvelope("approval",g),err}))
	}
	h.registerWorkspaceGoals(mux,s)
	h.registerWorkspaceContext(mux,s)
}

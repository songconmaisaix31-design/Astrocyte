package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type S2Service struct {
	repo WorkspaceRepository
	files ProjectFiles
	references ProjectReferences
	snapshots ProjectSnapshotProvider
	objects ContextObjectStore
	owned OwnedSessionProvider
	admissions AdmittedOpportunityReader
}

func NewS2Service(repo WorkspaceRepository,files ProjectFiles,refs ProjectReferences) *S2Service { return &S2Service{repo:repo,files:files,references:refs} }

// 仅在主入口启动并发处理之前装配。
func (s *S2Service) ConfigureContext(snapshots ProjectSnapshotProvider,objects ContextObjectStore,owned OwnedSessionProvider) {
	s.snapshots,s.objects,s.owned=snapshots,objects,owned
}

func (s *S2Service) ConfigureAdmissions(reader AdmittedOpportunityReader) { s.admissions=reader }

func absent(err error) bool {var e *apierrors.ServiceError;return errors.As(err,&e)&&e.Code==apierrors.NotFound}

func NormalizeWorkspaceError(err error) error {
	var rule *domain.WorkspaceRuleError
	if errors.As(err,&rule){return &apierrors.ServiceError{Code:apierrors.Code(rule.Code),Message:rule.Message,RequiredAction:"review_workspace_version_and_permissions"}}
	return err
}

func semanticWorkspaceInput(input any) ([]byte,error) {
	data,err:=json.Marshal(input);if err!=nil{return nil,err}
	var fields map[string]json.RawMessage
	if err=json.Unmarshal(data,&fields);err!=nil{return nil,err}
	delete(fields,"request_id")
	for _,key:=range []string{"Command","Meta"}{
		if nested,ok:=fields[key];ok{var command map[string]json.RawMessage;if err=json.Unmarshal(nested,&command);err!=nil{return nil,err};delete(command,"request_id");fields[key],err=json.Marshal(command);if err!=nil{return nil,err}}
	}
	return json.Marshal(fields)
}

func (s *S2Service) command(ctx context.Context,c domain.Caller,meta WorkspaceCommandMeta,name string,input any,output any,fn func(WorkspaceTx)error) (err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	if c.ID=="" || (c.Kind!="human"&&c.Kind!="agent") || meta.SchemaVersion!=1 || meta.RequestID=="" || meta.IdempotencyKey=="" || meta.ExpectedVersion<0 {return domain.WorkspaceError(apierrors.ValidationFailed,"authenticated caller and complete command metadata are required")}
	data,err:=semanticWorkspaceInput(input);if err!=nil{return err};digest:=sha256.Sum256(data);requestDigest:=hex.EncodeToString(digest[:]);caller:=c.Kind+":"+c.ID
	return s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{
		old,err:=tx.LoadReceipt(caller,name,meta.IdempotencyKey)
		if err==nil {if old.Digest!=requestDigest{return domain.WorkspaceError(apierrors.VersionConflict,"idempotency key binds different command input")};return json.Unmarshal(old.Result,output)}
		if !absent(err){return err}
		if err=fn(tx);err!=nil{return err}
		result,err:=json.Marshal(output);if err!=nil{return err}
		return tx.SaveReceipt(caller,name,meta.IdempotencyKey,WorkspaceReceipt{Digest:requestDigest,Result:result})
	})
}

func workspaceEvent(tx WorkspaceTx,meta WorkspaceCommandMeta,kind,id string,version int,payload any) error {
	data,err:=json.Marshal(payload);if err!=nil{return err}
	return tx.AppendWorkspaceEvent(WorkspaceEvent{ID:uuid.NewString(),Type:"workspace."+kind,SchemaVersion:1,AggregateID:id,AggregateVersion:version,OccurredAt:time.Now().UTC(),CorrelationID:meta.RequestID,CausationID:meta.RequestID,Payload:data})
}

func projectAccess(tx WorkspaceTx,c domain.Caller,id string) (domain.LocalProject,error) {
	p,err:=tx.LoadProject(id);if err!=nil{return p,err}
	if c.Kind=="human"&&c.ID!="" {return p,nil}
	if c.Kind!="agent"||c.ID==""||c.ProjectID!=id{return p,domain.WorkspaceError(apierrors.ScopeDenied,"caller is outside project")}
	g,err:=tx.LoadProjectGrant(id,c.ID);if err!=nil{return p,err}
	if g.RevokedAt!=nil || g.ExpiresAt!=nil&&!g.ExpiresAt.After(time.Now()) {return p,domain.WorkspaceError(apierrors.ApprovalRevoked,"project grant expired or revoked")}
	if !has(g.Actions,"read_context"){return p,domain.WorkspaceError(apierrors.ScopeDenied,"project context permission is required")};return p,nil
}

func (s *S2Service) validateInputs(ctx context.Context,c domain.Caller,content domain.ProposalContent) (int,error) {
	var p domain.LocalProject
	if err:=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{var err error;p,err=projectAccess(tx,c,content.Scope.ProjectID);return err});err!=nil{return 0,NormalizeWorkspaceError(err)}
	for _,ref:=range content.Inputs{
		if s.references==nil{return 0,&apierrors.ServiceError{Code:apierrors.UnsupportedCapability,Message:"project reference reader is not connected"}}
		material,err:=s.references.ReadSelected(ctx,c,p.SpaceID,ref,false);if err!=nil{return 0,err}
		if material.Reference!=ref{return 0,&apierrors.ServiceError{Code:apierrors.ContextStale,Message:"reference reader changed the fixed revision"}}
	}
	return p.Settings.Revision,nil
}

func (s *S2Service) validateProjectRoot(p domain.LocalProject,root string) error {
	if s.files==nil{return domain.WorkspaceError(apierrors.UnsupportedCapability,"project path validation is unavailable")}
	canonical,err:=s.files.CanonicalRoot(root);if err!=nil{return err};if canonical!=root{return domain.WorkspaceError(apierrors.ScopeDenied,"protected root must use canonical spelling")}
	rel,err:=filepath.Rel(p.Root,root);if err!=nil{return err};if _,err=s.files.ValidateProjectPath(p.Root,rel);err!=nil{return err}
	if !p.Settings.AllowDirectory{return domain.WorkspaceError(apierrors.ScopeDenied,"project directory access is not approved")}
	for _,sub:=range p.Settings.AllowedSubdirs{dir,err:=s.files.ValidateProjectPath(p.Root,sub);if err!=nil{return err};relative,err:=filepath.Rel(dir,root);if err==nil&&relative!=".."&&!strings.HasPrefix(relative,".."+string(filepath.Separator))&&!filepath.IsAbs(relative){return nil}}
	return domain.WorkspaceError(apierrors.ScopeDenied,"protected root is outside current allowed subdirectories")
}

func (s *S2Service) validateContent(ctx context.Context,tx WorkspaceTx,c domain.Caller,content domain.ProposalContent,execute bool,inputSettings int) error {
	if err:=content.Validate(time.Now().UTC(),execute);err!=nil{return err}
	p,err:=projectAccess(tx,c,content.Scope.ProjectID);if err!=nil{return err}
	if p.Settings.Revision!=inputSettings{return domain.WorkspaceError(apierrors.ContextStale,"project permissions changed during input validation")}
	if content.Scope.SettingsRevision!=p.Settings.Revision{return domain.WorkspaceError(apierrors.ContextStale,"proposal fixes another project settings revision")}
	g,err:=tx.LoadGoal(content.Goal.ID,content.Goal.Revision);if err!=nil{return err}
	if g.ProjectID!=p.ID{return domain.WorkspaceError(apierrors.ScopeDenied,"goal is outside proposal project")}
	for _,root:=range content.Scope.Roots{
		if err=s.validateProjectRoot(p,root);err!=nil{return err}
	}
	if execute{for _,action:=range content.Scope.Actions{if err=ValidateProjectActionScope(p,ProjectActionScope{ProjectID:p.ID,SettingsRevision:content.Scope.SettingsRevision,Action:action,Roots:content.Scope.Roots,Tools:content.Scope.Tools,ExternalTargets:content.Scope.ExternalTargets});err!=nil{return err}}}
	return nil
}

func (s *S2Service) PutGoal(ctx context.Context,c domain.Caller,id string,cmd GoalCommand) (g domain.Goal,err error) {
	if err=human(c);err!=nil{return}
	err=s.command(ctx,c,cmd.WorkspaceCommandMeta,"workspace.goal.put",struct{ID string;Command GoalCommand}{id,cmd},&g,func(tx WorkspaceTx)error{
		g=cmd.Goal
		if g.Title==""||len(g.SuccessCriteria)==0||g.ProjectID==""||!has([]string{"active","paused","completed","archived"},g.Status){return domain.WorkspaceError(apierrors.ValidationFailed,"goal needs project, title, success criteria and valid status")}
		if _,err:=projectAccess(tx,c,g.ProjectID);err!=nil{return err}
		if id==""{if cmd.ExpectedVersion!=0{return domain.WorkspaceError(apierrors.VersionConflict,"new goal expects zero")};g.ID=uuid.NewString()}else{old,err:=tx.LoadGoal(id,0);if err!=nil{return err};if old.Revision!=cmd.ExpectedVersion{return domain.WorkspaceError(apierrors.VersionConflict,"goal changed")};if old.ProjectID!=g.ProjectID{return domain.WorkspaceError(apierrors.ScopeDenied,"goal cannot change project")};g.ID=id}
		g.Revision=cmd.ExpectedVersion+1;g.CreatedAt=time.Now().UTC();return tx.SaveGoal(g,cmd.ExpectedVersion)
	});return
}

func (s *S2Service) CreateProposal(ctx context.Context,c domain.Caller,cmd CreateProposalCommand) (p domain.Proposal,err error) {
	inputSettings,err:=s.validateInputs(ctx,c,cmd.Content);if err!=nil{return p,err}
	if cmd.Content.Scope.SettingsRevision==0{cmd.Content.Scope.SettingsRevision=inputSettings}
	err=s.command(ctx,c,cmd.WorkspaceCommandMeta,"workspace.proposal.create",cmd,&p,func(tx WorkspaceTx)error{
		if cmd.ExpectedVersion!=0{return domain.WorkspaceError(apierrors.VersionConflict,"new proposal expects zero")}
		if err:=s.validateContent(ctx,tx,c,cmd.Content,false,inputSettings);err!=nil{return err}
		now:=time.Now().UTC();p=domain.Proposal{ID:uuid.NewString(),Version:1,Revision:1,Status:domain.ProposalDraft,Content:cmd.Content,AssignmentStatus:"assigned",CreatedAt:now,UpdatedAt:now};return tx.SaveProposal(p,0)
	});return
}

func (s *S2Service) ReviseProposal(ctx context.Context,c domain.Caller,id string,cmd CreateProposalCommand) (p domain.Proposal,err error) {
	inputSettings,err:=s.validateInputs(ctx,c,cmd.Content);if err!=nil{return p,err}
	if cmd.Content.Scope.SettingsRevision==0{cmd.Content.Scope.SettingsRevision=inputSettings}
	err=s.command(ctx,c,cmd.WorkspaceCommandMeta,"workspace.proposal.revise",struct{ID string;Command CreateProposalCommand}{id,cmd},&p,func(tx WorkspaceTx)error{
		old,err:=tx.LoadProposal(id,0);if err!=nil{return err};if old.Version!=cmd.ExpectedVersion{return domain.WorkspaceError(apierrors.VersionConflict,"proposal changed")}
		if old.Content.Scope.ProjectID==""{if err=human(c);err!=nil{return err}}else{if _,err=projectAccess(tx,c,old.Content.Scope.ProjectID);err!=nil{return err};if old.Content.Scope.ProjectID!=cmd.Content.Scope.ProjectID{return domain.WorkspaceError(apierrors.ScopeDenied,"proposal cannot change project")}}
		if err=s.validateContent(ctx,tx,c,cmd.Content,false,inputSettings);err!=nil{return err}
		old.Status=domain.ProposalSuperseded;old.Version++;old.UpdatedAt=time.Now().UTC();if err=tx.SaveProposal(old,cmd.ExpectedVersion);err!=nil{return err}
		p=old;p.Revision++;p.Version++;p.Status=domain.ProposalDraft;p.Content=cmd.Content;p.GrantID="";p.AssignmentStatus="assigned";return tx.SaveProposal(p,old.Version)
	});return
}

func (s *S2Service) TransitionProposal(ctx context.Context,c domain.Caller,id string,meta WorkspaceCommandMeta,target domain.ProposalStatus) (p domain.Proposal,err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	if target!=domain.ProposalInReview&&target!=domain.ProposalDeclined{return p,domain.WorkspaceError(apierrors.ValidationFailed,"transition must submit or decline")}
	if target==domain.ProposalDeclined {if err=human(c);err!=nil{return}}
	current,err:=s.ReadProposal(ctx,c,id,0);if err!=nil{return p,err};inputSettings,err:=s.validateInputs(ctx,c,current.Content);if err!=nil{return p,err}
	err=s.command(ctx,c,meta,"workspace.proposal."+string(target),struct{ID string;Meta WorkspaceCommandMeta}{id,meta},&p,func(tx WorkspaceTx)error{
		var err error;p,err=tx.LoadProposal(id,0);if err!=nil{return err}
		if p.Revision!=current.Revision{return domain.WorkspaceError(apierrors.VersionConflict,"proposal inputs changed during validation")}
		if err=s.validateContent(ctx,tx,c,p.Content,target==domain.ProposalInReview,inputSettings);err!=nil{return err}
		if err=p.Transition(meta.ExpectedVersion,target,time.Now().UTC());err!=nil{return err};return tx.SaveProposal(p,meta.ExpectedVersion)
	});return
}

func (s *S2Service) ApproveProposal(ctx context.Context,c domain.Caller,id string,cmd ApproveProposalCommand) (g domain.ApprovalGrant,err error) {
	if err=human(c);err!=nil{return}
	current,err:=s.ReadProposal(ctx,c,id,0);if err!=nil{return g,err};inputSettings,err:=s.validateInputs(ctx,c,current.Content);if err!=nil{return g,err}
	err=s.command(ctx,c,cmd.WorkspaceCommandMeta,"workspace.proposal.approve",struct{ID string;Command ApproveProposalCommand}{id,cmd},&g,func(tx WorkspaceTx)error{
		p,err:=tx.LoadProposal(id,0);if err!=nil{return err}
		if p.Revision!=current.Revision{return domain.WorkspaceError(apierrors.VersionConflict,"proposal inputs changed during approval")}
		if p.Revision!=cmd.ExpectedRevision{return domain.WorkspaceError(apierrors.VersionConflict,"proposal revision changed")}
		if err=s.validateContent(ctx,tx,c,p.Content,true,inputSettings);err!=nil{return err}
		now:=time.Now().UTC();if !cmd.ExpiresAt.After(now)||cmd.ExpiresAt.After(p.Content.Stops.Deadline){return domain.WorkspaceError(apierrors.ValidationFailed,"approval expiration must fit stopping deadline")}
		if err=p.Transition(cmd.ExpectedVersion,domain.ProposalApproved,now);err!=nil{return err}
		g=domain.ApprovalGrant{ID:uuid.NewString(),ProposalID:p.ID,ProposalRevision:p.Revision,Goal:p.Content.Goal,Epoch:1,Status:"active",Scope:p.Content.Scope,Budget:p.Content.Budget,Stops:p.Content.Stops,ApproverID:c.ID,ApprovedAt:now,ChangedAt:now,ExpiresAt:cmd.ExpiresAt}
		p.GrantID=g.ID
		if err=tx.SaveProposal(p,cmd.ExpectedVersion);err!=nil{return err};if err=tx.SaveApproval(g,0);err!=nil{return err}
		return workspaceEvent(tx,cmd.WorkspaceCommandMeta,"proposal_approved",p.ID,p.Version,domain.ProposalApprovedEvent{ProposalID:p.ID,ProposalRevision:p.Revision,GrantID:g.ID,Epoch:g.Epoch,ActorID:c.ID,ActorKind:c.Kind})
	});return
}

func (s *S2Service) ChangeApproval(ctx context.Context,c domain.Caller,id string,cmd ChangeApprovalCommand,status string) (g domain.ApprovalGrant,err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	if err=human(c);err!=nil{return}
	if !has([]string{"revoked","completed","expired"},status){return g,domain.WorkspaceError(apierrors.ValidationFailed,"invalid approval state")}
	err=s.command(ctx,c,cmd.WorkspaceCommandMeta,"workspace.approval."+status,struct{ID string;Command ChangeApprovalCommand}{id,cmd},&g,func(tx WorkspaceTx)error{
		var err error;g,err=tx.LoadApproval(id);if err!=nil{return err}
		if g.Epoch!=cmd.ExpectedVersion{return domain.WorkspaceError(apierrors.VersionConflict,"approval epoch changed")}
		if g.Status!="active"||strings.TrimSpace(cmd.Reason)==""{return domain.WorkspaceError(apierrors.ValidationFailed,"active approval and reason required")}
		if status=="expired"&&g.ExpiresAt.After(time.Now()){return domain.WorkspaceError(apierrors.ValidationFailed,"approval has not expired")}
		g.Status=status;g.Epoch++;g.Reason=cmd.Reason;g.ChangedAt=time.Now().UTC();if err=tx.SaveApproval(g,cmd.ExpectedVersion);err!=nil{return err};return workspaceEvent(tx,cmd.WorkspaceCommandMeta,"approval_"+status,g.ID,g.Epoch,map[string]any{"grant_id":g.ID,"new_epoch":g.Epoch,"reason":g.Reason})
	});return
}

func (s *S2Service) QueryGrant(ctx context.Context,id string) (g domain.ApprovalGrant,err error) {
	err=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{var e error;g,e=tx.LoadApproval(id);if e==nil&&g.Status=="active"&&(!g.ExpiresAt.After(time.Now())||!g.Stops.Deadline.After(time.Now())){g.Status="expired"};return e});return
}

func (s *S2Service) WithAuthorizedGrant(ctx context.Context,check domain.GrantCheck,fn func(context.Context,domain.ApprovalGrant)error) (err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	return s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{
		g,err:=tx.LoadApproval(check.GrantID);if err!=nil{return err};if err=g.Check(check,time.Now().UTC());err!=nil{return err}
		p,err:=tx.LoadProject(g.Scope.ProjectID);if err!=nil{return err}
		if g.Scope.SettingsRevision!=p.Settings.Revision{return domain.WorkspaceError(apierrors.ApprovalRevoked,"project settings changed since approval")}
		var member *domain.ProjectGrant
		if check.Caller.Kind=="agent"{loaded,err:=tx.LoadProjectGrant(p.ID,check.Caller.ID);if err!=nil{return err};member=&loaded}
		if err=AuthorizeProjectActionCaller(p,check.Caller,member,check.Action,time.Now().UTC());err!=nil{return err}
		if err=ValidateProjectActionScope(p,ProjectActionScope{ProjectID:p.ID,SettingsRevision:g.Scope.SettingsRevision,Action:check.Action,Roots:[]string{check.Root},Tools:check.Tools,ExternalTargets:check.ExternalTargets});err!=nil{return err}
		if err=s.validateProjectRoot(p,check.Root);err!=nil{return err}
		proposal,err:=tx.LoadProposal(g.ProposalID,g.ProposalRevision);if err!=nil{return err}
		for _,ref:=range proposal.Content.Inputs{
			if s.references==nil{return domain.WorkspaceError(apierrors.UnsupportedCapability,"current source permission reader is required")}
			material,err:=s.references.ReadSelected(tx.Context(),check.Caller,p.SpaceID,ref,false);if err!=nil{return err};if material.Reference!=ref{return domain.WorkspaceError(apierrors.ContextStale,"current source reader changed approved revision")}
		}
		return fn(tx.Context(),g)
	})
}

func (s *S2Service) ListProposalsFor(ctx context.Context,c domain.Caller) ([]domain.Proposal,error) {
	if c.Kind=="human" {if err:=human(c);err!=nil{return nil,err};return s.repo.ListWorkspaceProposals(ctx,"")}
	if err:=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{_,err:=projectAccess(tx,c,c.ProjectID);return err});err!=nil{return nil,err};return s.repo.ListWorkspaceProposals(ctx,c.ProjectID)
}

func (s *S2Service) ListGoalsFor(ctx context.Context,c domain.Caller) ([]domain.Goal,error) {
	if c.Kind=="human"{if err:=human(c);err!=nil{return nil,err};return s.repo.ListWorkspaceGoals(ctx,"")}
	if err:=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{_,err:=projectAccess(tx,c,c.ProjectID);return err});err!=nil{return nil,NormalizeWorkspaceError(err)}
	return s.repo.ListWorkspaceGoals(ctx,c.ProjectID)
}

func (s *S2Service) ReadProposal(ctx context.Context,c domain.Caller,id string,revision int) (p domain.Proposal,err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	err=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{var err error;p,err=tx.LoadProposal(id,revision);if err!=nil{return err};if p.Content.Scope.ProjectID==""{return human(c)};_,err=projectAccess(tx,c,p.Content.Scope.ProjectID);return err});return
}

func (s *S2Service) ReadGoal(ctx context.Context,c domain.Caller,id string,revision int) (g domain.Goal,err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	err=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{var err error;g,err=tx.LoadGoal(id,revision);if err!=nil{return err};_,err=projectAccess(tx,c,g.ProjectID);return err});return
}

func (s *S2Service) ReadApproval(ctx context.Context,c domain.Caller,id string) (g domain.ApprovalGrant,err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	err=s.repo.WithWorkspaceTx(ctx,func(tx WorkspaceTx)error{var err error;g,err=tx.LoadApproval(id);if err!=nil{return err};_,err=projectAccess(tx,c,g.Scope.ProjectID);return err});if err==nil&&g.Status=="active"&&(!g.ExpiresAt.After(time.Now())||!g.Stops.Deadline.After(time.Now())){g.Status="expired"};return
}

package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (s *S2Service) BindSession(ctx context.Context,c domain.Caller,cmd BindSessionCommand) (binding domain.SessionBinding,err error) {
	defer func(){err=NormalizeWorkspaceError(err)}()
	if err=human(c);err!=nil{return}
	if s.owned==nil{return binding,domain.WorkspaceError(apierrors.UnsupportedCapability,"current owned native session provider is unavailable")}
	if cmd.ExpectedVersion!=0||cmd.MemberID==""{return binding,domain.WorkspaceError(apierrors.ValidationFailed,"new binding requires zero version and member identity")}
	facts,err:=s.owned.ReadOwnedSessionFacts(ctx,cmd.SessionID);if err!=nil{return binding,err}
	if !facts.Owned||facts.SessionID!=cmd.SessionID||facts.NativeID==""||facts.AdapterID==""||facts.WorktreeID==""{return binding,domain.WorkspaceError(apierrors.EvidenceMissing,"native provider did not confirm current owned session")}
	err=s.command(ctx,c,cmd.WorkspaceCommandMeta,"workspace.session.bind",cmd,&binding,func(tx WorkspaceTx)error{
		packet,err:=tx.LoadContextPacket(cmd.PacketID);if err!=nil{return err};grant,project,err:=packetGrant(tx,c,packet.GrantID,packet.GrantEpoch);if err!=nil{return err}
		session,err:=tx.LoadNativeSession(cmd.SessionID);if err!=nil{return err}
		if session.ProjectID!=project.ID||facts.ProjectID!=project.ID||session.NativeID!=facts.NativeID||session.CLI!=facts.AdapterID||facts.WorktreeID!=packet.Snapshot.WorktreeID||facts.Root!=packet.Snapshot.Root||session.Ownership!="owned"{return domain.WorkspaceError(apierrors.ScopeDenied,"owned native session does not match fixed packet project and worktree")}
		memberGrant,err:=tx.LoadProjectGrant(project.ID,cmd.MemberID);if err!=nil{return err};if memberGrant.RevokedAt!=nil||memberGrant.ExpiresAt!=nil&&!memberGrant.ExpiresAt.After(time.Now()){return domain.WorkspaceError(apierrors.ApprovalRevoked,"member project grant is inactive")}
		now:=time.Now().UTC();binding=domain.SessionBinding{ID:uuid.NewString(),Version:1,ProjectID:project.ID,SessionID:session.ID,AdapterID:facts.AdapterID,NativeSessionID:facts.NativeID,WorktreeID:facts.WorktreeID,GrantID:grant.ID,GrantEpoch:grant.Epoch,PacketID:packet.ID,MemberID:cmd.MemberID,Status:"bound",Capabilities:facts.Capabilities,CreatedAt:now,UpdatedAt:now}
		return tx.SaveSessionBinding(binding,0)
	});return
}

func (s *S2Service) WithAuthorizedSessionBindings(ctx context.Context,check domain.GrantCheck,ids []string,fn func(context.Context,domain.ApprovalGrant,[]domain.SessionBinding)error) error {
	return s.WithAuthorizedGrant(ctx,check,func(txctx context.Context,g domain.ApprovalGrant)error{
		return s.repo.WithWorkspaceTx(txctx,func(tx WorkspaceTx)error{
			bindings:=make([]domain.SessionBinding,0,len(ids))
			for _,id:=range ids{
				if s.owned==nil{return domain.WorkspaceError(apierrors.UnsupportedCapability,"current owned registry facts are required for protected member actions")}
				b,err:=tx.LoadSessionBinding(id);if err!=nil{return err};if b.Status!="bound"||b.GrantID!=g.ID||b.GrantEpoch!=g.Epoch||b.ProjectID!=g.Scope.ProjectID{return domain.WorkspaceError(apierrors.ApprovalRevoked,"member binding is outside current approval")}
				facts,err:=s.owned.ReadOwnedSessionFacts(txctx,b.SessionID);if err!=nil{return err};if !facts.Owned||facts.ProjectID!=b.ProjectID||facts.NativeID!=b.NativeSessionID||facts.WorktreeID!=b.WorktreeID||facts.AdapterID!=b.AdapterID{return domain.WorkspaceError(apierrors.EvidenceMissing,"current registry ownership does not match member binding")}
				member,err:=tx.LoadProjectGrant(b.ProjectID,b.MemberID);if err!=nil{return err};if member.RevokedAt!=nil||member.ExpiresAt!=nil&&!member.ExpiresAt.After(time.Now()){return domain.WorkspaceError(apierrors.ApprovalRevoked,"member project grant is inactive")}
				if !has(member.Actions,check.Action){return domain.WorkspaceError(apierrors.ScopeDenied,"member action is not authorized")}
				session,err:=tx.LoadNativeSession(b.SessionID);if err!=nil{return err};if session.NativeID!=b.NativeSessionID||session.ProjectID!=b.ProjectID||session.Ownership!="owned"||session.StopConfirmed{return domain.WorkspaceError(apierrors.EvidenceMissing,"bound native session is no longer current owned")}
				bindings=append(bindings,b)
			}
			return fn(txctx,g,bindings)
		})
	})
}

package app

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func TestAuthorizedBindingMemberParametersRejectChangedIdentityAndSnapshot(t *testing.T) {
	p := domain.Principal{Kind: "agent", ID: "member", ProjectID: "project"}
	binding := AuthorizedBinding{ID: "binding", Purpose: "claim_work_item", Version: 2, PacketID: "packet", ProjectID: p.ProjectID, MemberID: p.ID, SessionID: "session", NativeSessionID: "native", AdapterID: "adapter", WorktreeID: "worktree", Root: "fixed-canonical-worktree"}
	member := Member{Purpose: "claim_work_item", BindingID: binding.ID, Version: "2", HolderID: p.ID, ProjectID: p.ProjectID, SessionID: binding.SessionID, Worktree: binding.Root, OperationID: "original-input"}
	ctx := WithAuthorizedBinding(context.Background(), binding)
	for _, purpose := range []string{"claim_work_item", "submit_artifact"} {
		current := member
		current.Purpose = purpose
		authorized := binding
		authorized.Purpose = purpose
		if _, err := AuthorizedBindingForMember(WithAuthorizedBinding(context.Background(), authorized), p, current); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := AuthorizedBindingForMember(context.Background(), p, member); err == nil {
		t.Fatal("member without actual authorized binding context was accepted")
	}
	for _, change := range []func(*Member){
		func(m *Member) { m.BindingID = "other-binding" },
		func(m *Member) { m.Version = "1" },
		func(m *Member) { m.HolderID = "other-member" },
		func(m *Member) { m.ProjectID = "other-project" },
		func(m *Member) { m.SessionID = "other-session" },
		func(m *Member) { m.Worktree = "other-worktree" },
		func(m *Member) { m.OperationID = "" },
		func(m *Member) { m.Purpose = "submit_artifact" },
		func(m *Member) { m.Purpose = "" },
	} {
		current := member
		change(&current)
		if _, err := AuthorizedBindingForMember(ctx, p, current); err == nil {
			t.Fatal("changed member identity or fixed context was accepted")
		}
	}
	for _, change := range []func(*AuthorizedBinding){
		func(b *AuthorizedBinding) { b.NativeSessionID = "" },
		func(b *AuthorizedBinding) { b.AdapterID = "" },
		func(b *AuthorizedBinding) { b.WorktreeID = "" },
		func(b *AuthorizedBinding) { b.PacketID = "" },
	} {
		current := binding
		change(&current)
		if _, err := AuthorizedBindingForMember(WithAuthorizedBinding(context.Background(), current), p, member); err == nil {
			t.Fatal("incomplete actual binding was accepted")
		}
	}
}

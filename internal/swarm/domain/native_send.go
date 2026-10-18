package domain

import (
	"slices"
	"time"
)

func (l Lease) CheckNativeSend(m Mission, w WorkItem, holder string, fence int64, now time.Time) error {
	if err := l.Check(holder, fence, now); err != nil {
		return err
	}
	if l.Fence < 1 || l.WorkItemID == "" || l.WorkItemID != w.ID || l.MissionID != m.ID || w.MissionID != m.ID || l.ProjectID != m.Grant.ProjectID || w.Fence != l.Fence || w.State != "leased" {
		return Failure("lease_lost", "current work item no longer owns the original send lease")
	}
	if err := m.CanExecute(now); err != nil {
		return err
	}
	if !now.Before(m.Grant.ExpiresAt) {
		return Failure("approval_revoked", "original send approval has expired")
	}
	if l.BindingID == "" || l.ContextPacketID == "" || l.MemberVersion == "" || l.OperationID == "" || l.SessionID == "" || l.NativeSessionID == "" || l.AdapterID == "" || l.ControllerID == "" || l.StartupOperationID == "" || l.Worktree == "" || l.BaseCommit == "" || l.PolicyVersion != LocalPolicyVersion {
		return Failure("evidence_missing", "original send lease lacks its fixed execution identity")
	}
	if l.Reserved.Calls < 1 || !l.Reserved.Fits(m.Reserved) || !slices.Equal(l.WritePaths, w.WritePaths) || !Within(w.WritePaths, m.Grant.WritePaths) {
		return Failure("scope_denied", "original send reservation or write scope changed")
	}
	return nil
}

package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// InventoryProvider returns a cached CLI inventory. Snapshot must not execute
// commands, inspect credentials or scan project/session directories.
type InventoryProvider interface {
	Snapshot(ctx context.Context) ([]domain.LocalAgent, error)
}

// Returns cached actual registry observations; querying never starts a CLI.
// Installation remains supplied independently by the PATH inventory.
type NativeRegistryObservations interface {
	SnapshotNative(context.Context) ([]domain.LocalAgent, error)
}

// LocalAgentInventory is a human-readable installation/capability query, not
// permission to start or resume an Agent or access its private native state.
type LocalAgentInventory interface {
	ListLocalAgents(ctx context.Context) (apierrors.ListResult, error)
}

type LocalProjectRepository interface {
	ListProjects(context.Context) ([]domain.LocalProject, error)
	LoadProject(context.Context, string) (domain.LocalProject, error)
	SaveProject(context.Context, domain.LocalProject, int) error
	LoadGrant(context.Context, string, string) (domain.ProjectGrant, error)
	SaveGrant(context.Context, domain.ProjectGrant) error
	ListSessions(context.Context, string) ([]domain.NativeSession, error)
	LoadSession(context.Context, string) (domain.NativeSession, error)
	SaveSession(context.Context, domain.NativeSession) error
}

// References are consumed from Attention through an entrypoint bridge. Neither
// app imports the other. Access and active space membership are checked per call.
type ProjectReferences interface {
	ReadSelected(context.Context, domain.Caller, string, domain.FixedReference, bool) (domain.ContextMaterial, error)
	Linked(context.Context, domain.Caller, string, domain.FixedReference) ([]domain.FixedReference, error)
}
type NativeAdapter interface {
	ID() string
	Version() string
	Capabilities() map[string]domain.CapabilityObservation
	Discover(context.Context, domain.LocalProject) ([]domain.NativeSession, error)
	ReadContext(context.Context, domain.NativeSession) ([]domain.NativeEvent, error)
	Start(context.Context, domain.NativeRequest) (domain.NativeSession, error)
	Resume(context.Context, domain.NativeRequest) (domain.NativeSession, error)
	Send(context.Context, domain.NativeSession, string) (domain.NativeObservation, error)
	Stop(context.Context, domain.NativeSession) (domain.NativeObservation, error)
	Observe(context.Context, domain.NativeSession) (domain.NativeObservation, error)
}
type NativeRegistry interface {
	Adapter(string) (NativeAdapter, error)
	List() []string
}
type ProjectFiles interface {
	CanonicalRoot(string) (string, error)
	ValidateProjectPath(string, string) (string, error)
	ReadProjectFiles(context.Context, domain.LocalProject, []string) ([]domain.ContextFile, error)
	DiscoverProjects(context.Context, string) ([]domain.ProjectCandidate, error)
}

// RegisteredProjectSource observes only registered roots through fixed programs.
// It never grants project access or starts an Agent. HTTP GET uses the repository.
type RegisteredProjectSource interface {
	DiscoverRegistered(context.Context) (domain.ProjectDiscoverySnapshot, error)
}

type RegisteredProjectRepository interface {
	LoadRegisteredProjectDiscovery(context.Context) (domain.ProjectDiscoverySnapshot, error)
	SaveRegisteredProjectDiscovery(context.Context, domain.ProjectDiscoverySnapshot) error
}

// Separate from project control so unsupported discovery remains explicit.
// Refresh is human-only; neither operation registers a space or changes grants.
type RegisteredProjects interface {
	ListRegisteredProjects(context.Context, domain.Caller) (domain.ProjectDiscoverySnapshot, error)
	RefreshRegisteredProjects(context.Context, domain.Caller) (domain.ProjectDiscoverySnapshot, error)
}
type TextProcessor interface {
	ConfigurationID(context.Context, string) (string, error)
	ProcessSelectedText(context.Context, domain.TextRequest) (domain.TextResult, error)
}

// LocalProjects is separate from the S0 project projection so old consumers do
// not mistake a native project root for an execution/environment identity.
type LocalProjects interface {
	ProbeNativeCLI(context.Context, domain.Caller, string, string) (domain.NativeSession, error)
	CheckProjectModel(context.Context, domain.Caller, string, string) (domain.LocalProject, error)
	Shutdown(context.Context) error
	ListProjects(context.Context, domain.Caller) ([]domain.LocalProject, error)
	DiscoverProjects(context.Context, domain.Caller, string) ([]domain.ProjectCandidate, error)
	RegisterProject(context.Context, domain.Caller, domain.RegisterProjectCommand) (domain.LocalProject, error)
	SetProjectSettings(context.Context, domain.Caller, string, domain.SettingsCommand) (domain.LocalProject, error)
	GrantProjectAgent(context.Context, domain.Caller, string, domain.ProjectGrant) (domain.ProjectGrant, error)
	RevokeProjectAgent(context.Context, domain.Caller, string, string) (domain.ProjectGrant, error)
	IssueProjectAgentToken(context.Context, domain.Caller, string, string) (domain.AgentToken, error)
	ReadProjectContext(context.Context, domain.Caller, string, domain.ContextRequest) (domain.ContextPacket, error)
	ListNativeSessions(context.Context, domain.Caller, string) ([]domain.NativeSession, error)
	DiscoverNativeSessions(context.Context, domain.Caller, string, string) ([]domain.NativeSession, error)
	ReadNativeContext(context.Context, domain.Caller, string, string) ([]domain.NativeEvent, error)
	StartNativeSession(context.Context, domain.Caller, string, domain.NativeCommand) (domain.NativeSession, error)
	ResumeNativeSession(context.Context, domain.Caller, string, string, domain.NativeCommand) (domain.NativeSession, error)
	SendNativeMessage(context.Context, domain.Caller, string, string, string) (domain.NativeObservation, error)
	StopNativeSession(context.Context, domain.Caller, string, string) (domain.NativeObservation, error)
	ObserveNativeSession(context.Context, domain.Caller, string, string) (domain.NativeObservation, error)
}

// ScopedAgentTokens authenticate only already human-granted project identities.
// Validation must reload the current grant; a token cannot approve its own scope.
type ScopedAgentTokens interface {
	AuthenticateAgentToken(context.Context, string) (domain.Caller, error)
}

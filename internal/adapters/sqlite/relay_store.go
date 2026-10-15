package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// RelayStore is isolated from the business DB and every published cognition
// snapshot. Its constructor never invokes the shared migration runner.
type RelayStore struct{ db *DB }

var _ app.RelayStore = (*RelayStore)(nil)

func RelayStateRoot() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" || !filepath.IsAbs(base) {
		return "", errors.New("absolute LOCALAPPDATA required for relay state")
	}
	return filepath.Join(base, "astrocyte-workbench", "v0.2-access"), nil
}

// Only the exact independent access directory is accepted. Tests use an
// isolated LOCALAPPDATA; HTTP cannot select a state path or DB filename.
func OpenRelayStore(root string) (*RelayStore, error) {
	expected, err := RelayStateRoot()
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(root)
	if err != nil || !strings.EqualFold(filepath.Clean(abs), filepath.Clean(expected)) {
		return nil, errors.New("relay state must use the independent v0.2-access directory")
	}
	// Reject redirecting an otherwise approved pathname into a formal DB or copy.
	for p := abs; ; p = filepath.Dir(p) {
		info, e := os.Lstat(p)
		if e == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("relay state symlink is not allowed")
		}
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return nil, e
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	if err = os.MkdirAll(abs, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(abs, "relay.sqlite")
	if info, e := os.Lstat(path); e == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("relay DB must be a regular independent file")
		}
		if err = verifyRelayFileReadOnly(path); err != nil {
			return nil, err
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	db, err := Open(path)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (*RelayStore, error) { _ = db.Close(); return nil, e }
	rows, err := db.conn.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return fail(err)
	}
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			_ = rows.Close()
			return fail(err)
		}
		if name != "relay_state" && name != "relay_receipts" {
			_ = rows.Close()
			return fail(errors.New("existing DB is not an exclusive relay store"))
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return fail(err)
	}
	_, err = db.conn.Exec(`CREATE TABLE IF NOT EXISTS relay_state (id INTEGER PRIMARY KEY CHECK(id=1), data TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS relay_receipts (id TEXT PRIMARY KEY, data TEXT NOT NULL, input BLOB NOT NULL);`)
	if err != nil {
		return fail(err)
	}
	if _, err = db.conn.Exec("PRAGMA wal_checkpoint(FULL)"); err != nil {
		return fail(err)
	}
	return &RelayStore{db: db}, nil
}

// Inspect before sqlite.Open can set WAL or any write PRAGMA. Immutable RO
// avoids WAL/SHM creation even for a foreign database placed as a hard link.
func verifyRelayFileReadOnly(path string) error {
	uriPath := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" && !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("immutable", "1")
	q.Set("_pragma", "query_only(1)")
	u.RawQuery = q.Encode()
	conn, err := sql.Open("sqlite", u.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return err
		}
		if name != "relay_state" && name != "relay_receipts" {
			return errors.New("existing file is not an exclusive relay store")
		}
		count++
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if count != 2 {
		return errors.New("existing file lacks the committed relay store identity")
	}
	return nil
}

func (s *RelayStore) Close() error { return s.db.Close() }

func emptyRelayState() app.RelayState {
	return app.RelayState{Tasks: map[string]domain.RelayTask{}, Workspaces: map[string]domain.RelayWorkspace{},
		Terminals: map[string]domain.RelayTerminal{}, Sessions: map[string]domain.NativeSession{},
		OwnedSessions: map[string]domain.RelayOwnedSession{},
		Receipts:      map[string]domain.RelayReceipt{}, Transitions: []domain.RelayTransition{},
		Events: []domain.RelayEvent{}, Published: map[string]bool{}, PendingActions: map[string]string{}, PublicationErrors: map[string]bool{}}
}

type relayQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type relayWorkspacePrivate struct {
	SourceRoot     string
	GitObservation string
}
type relayStateWire struct {
	State                 app.RelayState
	WorkspaceInternals    map[string]relayWorkspacePrivate
	SessionHistoryPaths   map[string]string
	OwnedSessionInternals map[string]relayOwnedSessionPrivate
	ControllerEvidence    map[string]relayControllerEvidencePrivate
	ControllerHistory     map[string][]relayControllerEvidencePrivate
}

// 私有字段仅在本机隔离库中编码，公开类型的 JSON 保持隐藏。
type relayControllerEvidencePrivate struct {
	ControllerID, PreviousControllerID, StartupOperationID string
	SessionID, NativeID, Provider, WorkspaceRoot, Relation string
	Facts                                                  []domain.RelayDriverFact
	Turns                                                  []relayTurnBindingPrivate
	ExitConfirmed                                          bool
}

type relayTurnBindingPrivate struct {
	OperationID, SessionID, NativeID, ControllerID, StartupOperationID, TurnID string
	AcceptedAt                                                                 time.Time
	StartedAt                                                                  *time.Time
}

func encodeRelayController(p domain.RelayControllerEvidence) relayControllerEvidencePrivate {
	result := relayControllerEvidencePrivate{ControllerID: p.ControllerID, PreviousControllerID: p.PreviousControllerID, StartupOperationID: p.StartupOperationID, SessionID: p.SessionID, NativeID: p.NativeID, Provider: p.Provider, WorkspaceRoot: p.WorkspaceRoot, Relation: p.Relation, Facts: p.Facts}
	result.ExitConfirmed = p.ExitConfirmed
	for _, t := range p.Turns {
		result.Turns = append(result.Turns, relayTurnBindingPrivate{OperationID: t.OperationID, SessionID: t.SessionID, NativeID: t.NativeID, ControllerID: t.ControllerID, StartupOperationID: t.StartupOperationID, TurnID: t.TurnID, AcceptedAt: t.AcceptedAt, StartedAt: t.StartedAt})
	}
	return result
}

func decodeRelayController(p relayControllerEvidencePrivate) domain.RelayControllerEvidence {
	result := domain.RelayControllerEvidence{ControllerID: p.ControllerID, PreviousControllerID: p.PreviousControllerID, StartupOperationID: p.StartupOperationID, SessionID: p.SessionID, NativeID: p.NativeID, Provider: p.Provider, WorkspaceRoot: p.WorkspaceRoot, Relation: p.Relation, Facts: p.Facts}
	result.ExitConfirmed = p.ExitConfirmed
	for _, t := range p.Turns {
		result.Turns = append(result.Turns, domain.RelayTurnBinding{OperationID: t.OperationID, SessionID: t.SessionID, NativeID: t.NativeID, ControllerID: t.ControllerID, StartupOperationID: t.StartupOperationID, TurnID: t.TurnID, AcceptedAt: t.AcceptedAt, StartedAt: t.StartedAt})
	}
	return result
}

// Only the isolated relay state encodes these references. DTO serialization
// keeps every path/native identifier hidden; no archive contents are copied.
type relayOwnedSessionPrivate struct {
	WorkspaceRoot, SourceRoot, SessionID, NativeID, Ownership string
	ArchivePath, ArchiveRoot, ArchiveFormat                   string
	NativeCapture                                             *domain.RelayNativeSessionCapture
	ControllerID, StartupOperationID                          string
	CaptureControllerID, CaptureOperationID                   string
}

func loadRelayState(ctx context.Context, q relayQuerier) (app.RelayState, error) {
	state := emptyRelayState()
	var data []byte
	err := q.QueryRowContext(ctx, "SELECT data FROM relay_state WHERE id=1").Scan(&data)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return state, err
	}
	if err == nil {
		var wire relayStateWire
		if err = json.Unmarshal(data, &wire); err != nil {
			return state, err
		}
		state = wire.State
		state.ControllerEvidence = map[string]domain.RelayControllerEvidence{}
		for id, p := range wire.ControllerEvidence {
			state.ControllerEvidence[id] = decodeRelayController(p)
		}
		state.ControllerHistory = map[string][]domain.RelayControllerEvidence{}
		for id, history := range wire.ControllerHistory {
			for _, p := range history {
				state.ControllerHistory[id] = append(state.ControllerHistory[id], decodeRelayController(p))
			}
		}
		for id, private := range wire.WorkspaceInternals {
			w := state.Workspaces[id]
			w.SourceRoot = private.SourceRoot
			w.GitObservation = private.GitObservation
			state.Workspaces[id] = w
		}
		for id, path := range wire.SessionHistoryPaths {
			session := state.Sessions[id]
			session.HistoryPath = path
			state.Sessions[id] = session
		}
		for id, private := range wire.OwnedSessionInternals {
			bound := state.OwnedSessions[id]
			bound.WorkspaceRoot, bound.SourceRoot, bound.SessionID, bound.NativeID, bound.Ownership = private.WorkspaceRoot, private.SourceRoot, private.SessionID, private.NativeID, private.Ownership
			bound.ArchivePath, bound.ArchiveRoot, bound.ArchiveFormat = private.ArchivePath, private.ArchiveRoot, private.ArchiveFormat
			bound.NativeCapture = private.NativeCapture
			bound.ControllerID, bound.StartupOperationID = private.ControllerID, private.StartupOperationID
			if bound.NativeCapture != nil {
				bound.NativeCapture.ControllerID, bound.NativeCapture.StartupOperationID = private.CaptureControllerID, private.CaptureOperationID
			}
			state.OwnedSessions[id] = bound
		}
	}
	if state.OwnedSessions == nil {
		state.OwnedSessions = map[string]domain.RelayOwnedSession{}
	}
	if state.ControllerEvidence == nil {
		state.ControllerEvidence = map[string]domain.RelayControllerEvidence{}
	}
	if state.HandoffDeclarations == nil {
		state.HandoffDeclarations = map[string]string{}
	}
	if state.ControllerHistory == nil {
		state.ControllerHistory = map[string][]domain.RelayControllerEvidence{}
	}
	if state.Published == nil {
		state.Published = map[string]bool{}
	}
	if state.PendingActions == nil {
		state.PendingActions = map[string]string{}
	}
	if state.PublicationErrors == nil {
		state.PublicationErrors = map[string]bool{}
	}
	state.Receipts = map[string]domain.RelayReceipt{}
	rows, err := q.QueryContext(ctx, "SELECT data,input FROM relay_receipts")
	if err != nil {
		return state, err
	}
	defer rows.Close()
	for rows.Next() {
		var receipt domain.RelayReceipt
		var input []byte
		if err = rows.Scan(&data, &input); err != nil {
			return state, err
		}
		if err = json.Unmarshal(data, &receipt); err != nil {
			return state, err
		}
		receipt.Input = input
		state.Receipts[receipt.ID] = receipt
	}
	return state, rows.Err()
}

func (s *RelayStore) Snapshot(ctx context.Context) (app.RelayState, error) {
	tx, err := s.db.conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return app.RelayState{}, err
	}
	defer tx.Rollback()
	state, err := loadRelayState(ctx, tx)
	if err != nil {
		return state, err
	}
	return state, tx.Commit()
}

// BEGIN IMMEDIATE arbitrates writers across service instances, not merely a
// Go mutex. The workspace reservation and receipt are committed atomically.
func (s *RelayStore) Update(ctx context.Context, update func(*app.RelayState) error) error {
	conn, err := s.db.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	defer conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
	state, err := loadRelayState(ctx, conn)
	if err != nil {
		return err
	}
	old := make(map[string]domain.RelayReceipt, len(state.Receipts))
	oldFacts := append([]domain.RelayRuntimeFact{}, state.RuntimeFacts...)
	oldRelations := append([]domain.RelaySessionRelation{}, state.SessionRelations...)
	oldHandoffs := make(map[string]string, len(state.HandoffDeclarations))
	oldHistory := make(map[string][]domain.RelayControllerEvidence, len(state.ControllerHistory))
	for id, history := range state.ControllerHistory {
		for _, evidence := range history {
			evidence.Facts = append([]domain.RelayDriverFact{}, evidence.Facts...)
			evidence.Turns = append([]domain.RelayTurnBinding{}, evidence.Turns...)
			for i, turn := range evidence.Turns {
				if turn.StartedAt != nil {
					when := *turn.StartedAt
					evidence.Turns[i].StartedAt = &when
				}
			}
			oldHistory[id] = append(oldHistory[id], evidence)
		}
	}
	for id, previous := range state.HandoffDeclarations {
		oldHandoffs[id] = previous
	}
	for id, receipt := range state.Receipts {
		old[id] = receipt
	}
	if err = update(&state); err != nil {
		return err
	}
	for id, previous := range oldHandoffs {
		if state.HandoffDeclarations[id] != previous {
			return errors.New("relay handoff declaration is immutable")
		}
	}
	for id, previous := range oldHistory {
		current := state.ControllerHistory[id]
		if len(current) < len(previous) {
			return errors.New("relay controller history cannot be removed")
		}
		for i, evidence := range previous {
			if !reflect.DeepEqual(evidence, current[i]) {
				return errors.New("relay controller history is append-only")
			}
		}
	}
	if len(state.RuntimeFacts) < len(oldFacts) || len(state.SessionRelations) < len(oldRelations) {
		return errors.New("relay facts and relations cannot be removed")
	}
	for i, previous := range oldFacts {
		if !reflect.DeepEqual(previous, state.RuntimeFacts[i]) {
			return errors.New("relay runtime facts are append-only")
		}
	}
	for i, previous := range oldRelations {
		if !reflect.DeepEqual(previous, state.SessionRelations[i]) {
			return errors.New("relay session relations are append-only")
		}
	}
	for id, previous := range old {
		current, ok := state.Receipts[id]
		if !ok {
			return errors.New("relay receipts cannot be removed")
		}
		if string(previous.Input) != string(current.Input) {
			return errors.New("relay receipt input is immutable")
		}
		if previous.Status != "pending" {
			a, _ := json.Marshal(previous)
			b, _ := json.Marshal(current)
			if string(a) != string(b) || string(previous.Input) != string(current.Input) {
				return errors.New("final relay receipt is immutable")
			}
		}
	}
	for id, receipt := range state.Receipts {
		if id != receipt.ID {
			return errors.New("relay receipt key mismatch")
		}
		data, e := json.Marshal(receipt)
		if e != nil {
			return e
		}
		if _, e = conn.ExecContext(ctx, `INSERT INTO relay_receipts(id,data,input) VALUES(?,?,?)
ON CONFLICT(id) DO UPDATE SET data=excluded.data,input=excluded.input`, id, string(data), receipt.Input); e != nil {
			return e
		}
	}
	state.Receipts = nil
	wire := relayStateWire{State: state, WorkspaceInternals: map[string]relayWorkspacePrivate{}}
	wire.ControllerEvidence = map[string]relayControllerEvidencePrivate{}
	wire.ControllerHistory = map[string][]relayControllerEvidencePrivate{}
	for id, history := range state.ControllerHistory {
		for _, p := range history {
			wire.ControllerHistory[id] = append(wire.ControllerHistory[id], encodeRelayController(p))
		}
	}
	for id, p := range state.ControllerEvidence {
		wire.ControllerEvidence[id] = encodeRelayController(p)
	}
	wire.SessionHistoryPaths = map[string]string{}
	for id, session := range state.Sessions {
		if session.HistoryPath != "" {
			wire.SessionHistoryPaths[id] = session.HistoryPath
		}
	}
	wire.OwnedSessionInternals = map[string]relayOwnedSessionPrivate{}
	for id, bound := range state.OwnedSessions {
		wire.OwnedSessionInternals[id] = relayOwnedSessionPrivate{WorkspaceRoot: bound.WorkspaceRoot, SourceRoot: bound.SourceRoot, SessionID: bound.SessionID, NativeID: bound.NativeID, Ownership: bound.Ownership, ArchivePath: bound.ArchivePath, ArchiveRoot: bound.ArchiveRoot, ArchiveFormat: bound.ArchiveFormat, NativeCapture: bound.NativeCapture}
		private := wire.OwnedSessionInternals[id]
		private.ControllerID, private.StartupOperationID = bound.ControllerID, bound.StartupOperationID
		if bound.NativeCapture != nil {
			private.CaptureControllerID, private.CaptureOperationID = bound.NativeCapture.ControllerID, bound.NativeCapture.StartupOperationID
		}
		wire.OwnedSessionInternals[id] = private
	}
	for id, w := range state.Workspaces {
		wire.WorkspaceInternals[id] = relayWorkspacePrivate{SourceRoot: w.SourceRoot, GitObservation: w.GitObservation}
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return err
	}
	if _, err = conn.ExecContext(ctx, "INSERT INTO relay_state(id,data) VALUES(1,?) ON CONFLICT(id) DO UPDATE SET data=excluded.data", string(data)); err != nil {
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("commit relay state: %w", err)
	}
	return nil
}

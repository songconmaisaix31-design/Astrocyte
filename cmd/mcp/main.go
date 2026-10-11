// Command mcp serves the five cognitive tools over stdin/stdout. Startup binds
// identity and project authority; tool arguments cannot change that binding.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/songconmaisaix31-design/Astrocyte/cmd/internal/attentionruntime"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/mcp"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func optionalFloat(text, name string) (*float64, error) {
	if text == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return nil, fmt.Errorf("%s must be an explicit finite nonnegative value", name)
	}
	return &v, nil
}

func run(ctx context.Context, args []string, in io.Reader, out, stderr io.Writer) error {
	flags := flag.NewFlagSet("astrocyte-mcp", flag.ContinueOnError)
	flags.SetOutput(stderr)
	data := flags.String("data-dir", "", "directory containing state.sqlite")
	project := flags.String("project", "", "trusted startup project scope")
	session := flags.String("session", "", "trusted startup session identity")
	id := flags.String("principal-id", "", "trusted caller identity")
	kind := flags.String("principal-kind", "", "agent or human")
	veto := flags.Bool("human-veto", false, "enable user veto for the human startup channel")
	strengthText := flags.String("signal-strength", "", "explicit signal strength; no product default")
	decayText := flags.String("decay", "", "explicit topic decay per day")
	confidenceText := flags.String("initial-confidence", "", "explicit initial confidence")
	successText := flags.String("success-confidence-delta", "", "explicit successful-task confidence increment")
	dropText := flags.String("confidence-drop-threshold", "", "explicit conflict threshold")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *project == "" || *session == "" || *id == "" || (*kind != "human" && *kind != "agent") || (*veto && *kind != "human") {
		return fmt.Errorf("explicit project, session, principal-id and principal-kind are required; human-veto requires human")
	}
	strength, err := optionalFloat(*strengthText, "signal-strength")
	if err != nil {
		return err
	}
	if strength == nil {
		return fmt.Errorf("signal-strength remains undecided; supply an explicit startup value")
	}
	options := app.DynamicsOptions{}
	for _, p := range []struct {
		text, name string
		dst        **float64
	}{{*decayText, "decay", &options.Decay}, {*confidenceText, "initial-confidence", &options.InitialConfidence}, {*successText, "success-confidence-delta", &options.SuccessConfidenceDelta}, {*dropText, "confidence-drop-threshold", &options.ConfidenceDropThreshold}} {
		*p.dst, err = optionalFloat(p.text, p.name)
		if err != nil {
			return err
		}
	}
	if *data == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		*data = filepath.Join(base, "astrocyte")
	}
	if err = os.MkdirAll(*data, 0700); err != nil {
		return err
	}
	db, err := sqlite.Open(filepath.Join(*data, "state.sqlite"))
	if err != nil {
		return err
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		return err
	}
	repo := sqlite.NewEventRepository(db)
	experiences := app.NewExperienceService(sqlite.NewExperienceRepository(db), repo, app.ExperienceOptions{ProjectID: *project})
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), options)
	events := app.NewEventService(repo, *strength, experiences.CaptureCommittedFailure)
	recorder := &attentionruntime.EventRecorder{Events: events, Repository: repo, Engine: engine}
	binding := app.MCPBinding{Principal: app.Principal{ID: *id, Kind: *kind}, Scope: *project, Session: *session, CanVeto: *veto}
	backend, err := app.NewEngineMCPBackend(binding, engine, experiences, recorder, app.MCPBackendOptions{})
	if err != nil {
		return err
	}
	server, err := mcp.NewAstrocyteServer(backend, binding)
	if err != nil {
		return err
	}
	return server.Serve(ctx, in, out)
}

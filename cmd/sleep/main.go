// Command sleep runs the original daily consolidation once for an explicitly
// selected calendar day. It does not schedule tasks or call a model.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
	_ "time/tzdata"

	"github.com/songconmaisaix31-design/Astrocyte/cmd/attentionruntime"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func policy(value, name string, required bool) (*float64, error) {
	if value == "" {
		if required {
			return nil, fmt.Errorf("%s remains undecided; supply an explicit value", name)
		}
		return nil, nil
	}
	n, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
		return nil, fmt.Errorf("%s requires a finite nonnegative value", name)
	}
	return &n, nil
}
func run(ctx context.Context, args []string, out, stderr io.Writer) error {
	flags := flag.NewFlagSet("astrocyte-sleep", flag.ContinueOnError)
	flags.SetOutput(stderr)
	data := flags.String("data-dir", "", "directory containing state.sqlite")
	date := flags.String("day", "", "explicit calendar day YYYY-MM-DD")
	zone := flags.String("timezone", "", "explicit IANA timezone, for example Asia/Shanghai or UTC")
	decayText := flags.String("decay", "", "explicit per-day topic decay; no default")
	confidenceText := flags.String("initial-confidence", "", "explicit new topic confidence; no default")
	successText := flags.String("success-confidence-delta", "", "optional explicit successful-task confidence increment")
	dropText := flags.String("confidence-drop-threshold", "", "optional explicit conflict threshold")
	facts := flags.String("fact-projection", "", "explicit none or facts-map-v1 for known literal raw.facts sources")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *date == "" || *zone == "" || (*facts != "none" && *facts != "facts-map-v1") {
		return fmt.Errorf("day, timezone and explicit fact-projection=none|facts-map-v1 are required")
	}
	location, err := time.LoadLocation(*zone)
	if err != nil {
		return fmt.Errorf("timezone: %w", err)
	}
	day, err := time.ParseInLocation(time.DateOnly, *date, location)
	if err != nil {
		return fmt.Errorf("day: %w", err)
	}
	options := app.DynamicsOptions{}
	for _, item := range []struct {
		value, name string
		required    bool
		dst         **float64
	}{{*decayText, "decay", true, &options.Decay}, {*confidenceText, "initial-confidence", true, &options.InitialConfidence}, {*successText, "success-confidence-delta", false, &options.SuccessConfidenceDelta}, {*dropText, "confidence-drop-threshold", false, &options.ConfidenceDropThreshold}} {
		*item.dst, err = policy(item.value, item.name, item.required)
		if err != nil {
			return err
		}
	}
	if *options.InitialConfidence > 1 {
		return fmt.Errorf("initial-confidence must not exceed one")
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
	sleepOptions := app.SleepOptions{}
	projection, limitation := "unconfigured", "fact retention is unmeasurable without an explicitly configured source projection; no model was called"
	if *facts == "facts-map-v1" {
		sleepOptions.ProjectFacts = app.LiteralSleepFactsMap
		projection = "facts-map-v1"
		limitation = "literal string raw.facts from an explicitly selected known source format; no inferred facts or model"
	}
	result, err := attentionruntime.NewSleepService(db, options, sleepOptions).Run(ctx, day)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		app.SleepResult
		FactProjection string `json:"fact_projection"`
		Limitation     string `json:"limitation"`
	}{result, projection, limitation})
}

package agents

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// The caller owns public-source/project consent checks. This port accepts only
// text already selected by that owner and never chooses another installed CLI.
func (r *Registry) textSession(ctx context.Context, cli, prompt string, seconds int) (*Native, domain.NativeSession, error) {
	a, err := r.Adapter(cli)
	if err != nil {
		return nil, domain.NativeSession{}, err
	}
	n := a.(*Native)
	root, err := os.MkdirTemp("", "astrocyte-selected-text-")
	if err != nil {
		return nil, domain.NativeSession{}, err
	}
	// Empty nonprivate runtime directory retained; removing runtime state is a
	// separate cleanup policy, never a reason to change native session identity.
	id := uuid.NewString()
	p := domain.LocalProject{ID: id, Root: root, Settings: domain.ProjectSettings{Revision: 1}}
	s, err := n.Start(ctx, domain.NativeRequest{SessionID: id, Project: p, Session: domain.NativeSession{ID: id, ProjectID: id}, Command: domain.NativeCommand{Message: prompt, DeadlineSeconds: seconds}, Packet: domain.ContextPacket{SchemaVersion: 1, ProjectID: id, Mode: "selected_text"}})
	return n, s, err
}

func (r *Registry) ConfigurationID(ctx context.Context, cli string) (string, error) {
	n, s, err := r.textSession(ctx, cli, "", 30)
	if err != nil {
		return "", err
	}
	p, err := n.get(s)
	if err != nil {
		return "", err
	}
	obs, stopErr := n.Stop(context.WithoutCancel(ctx), s)
	if stopErr != nil || !obs.StopConfirmed {
		return "", nativeError(apierrors.DeliveryUnknown, "configuration native process exit is unconfirmed")
	}
	if p.model == "" || p.provider == "" {
		return "", nativeError(apierrors.EvidenceMissing, "native current model/provider could not be observed without inference")
	}
	return fmt.Sprintf("cli:%s;version:%s;model:%s;provider:%s;policy:selected-text-tools-disabled-v1", cli, n.Version(), p.model, p.provider), nil
}

func (r *Registry) ProcessSelectedText(ctx context.Context, input domain.TextRequest) (domain.TextResult, error) {
	var result domain.TextResult
	if strings.TrimSpace(input.Prompt) == "" || len(input.Prompt) > 128*1024 {
		return result, nativeError(apierrors.BudgetExhausted, "selected text input is empty or exceeds 128 KiB")
	}
	seconds := input.DeadlineSeconds
	if seconds == 0 {
		seconds = 180
	}
	if seconds < 1 || seconds > 1800 {
		return result, nativeError(apierrors.BudgetExhausted, "selected text deadline is outside 1 to 1800 seconds")
	}
	limit := input.OutputLimit
	if limit == 0 {
		limit = 64 * 1024
	}
	if limit < 1 || limit > 128*1024 {
		return result, nativeError(apierrors.BudgetExhausted, "selected text output limit is invalid")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()
	n, s, err := r.textSession(ctx, input.CLI, input.Prompt, seconds)
	if err != nil {
		return result, err
	}
	defer func() { _, _ = n.Stop(context.Background(), s) }()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		obs, err := n.Observe(ctx, s)
		if err != nil {
			return result, err
		}
		if obs.OutputTruncated || obs.Status == "blocked" {
			return result, nativeError(apierrors.BudgetExhausted, "native output exceeded bounds")
		}
		if obs.Status == "failed" || (obs.StopConfirmed && obs.Status != "completed") {
			return result, nativeError(apierrors.ProviderUnavailable, "native selected-text processing failed")
		}
		if obs.Status == "completed" {
			var text strings.Builder
			for _, event := range obs.Events {
				if event.Kind == "text" {
					text.WriteString(event.Text)
				}
			}
			if text.Len() == 0 || text.Len() > limit {
				return result, nativeError(apierrors.EvidenceMissing, "native output is empty or exceeds selected-text bound")
			}
			result.Text = text.String()
			result.NativeID = s.NativeID
			result.Version = s.Version
			p, _ := n.get(s)
			if p != nil && p.model != "" {
				model := p.model
				result.Model = &model
			}
			// Native usage/cost not implemented in this projection: null remains
			// explicit; neither zero usage nor an invented charge is returned.
			return result, nil
		}
		select {
		case <-ctx.Done():
			return result, nativeError(apierrors.DeliveryUnknown, "native selected-text result is unknown after deadline")
		case <-ticker.C:
		}
	}
}

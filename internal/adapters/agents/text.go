package agents

import (
	"context"
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
	if err := ctx.Err(); err != nil {
		return "", err
	}
	a, err := r.Adapter(cli)
	if err != nil {
		return "", err
	}
	n := a.(*Native)
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.configID == "" || time.Since(n.configAt) > 5*time.Minute {
		return "", nativeError(apierrors.EvidenceMissing, "native configuration is unobserved or stale; explicitly probe the approved CLI")
	}
	return n.configID, nil
}

func (r *Registry) ProcessSelectedText(ctx context.Context, input domain.TextRequest) (result domain.TextResult, resultErr error) {
	if strings.TrimSpace(input.Prompt) == "" || len(input.Prompt) > 128*1024 {
		return result, nativeError(apierrors.ValidationFailed, "selected text input is empty or exceeds 128 KiB")
	}
	seconds := input.DeadlineSeconds
	if seconds == 0 {
		seconds = 180
	}
	if seconds < 1 || seconds > 1800 {
		return result, nativeError(apierrors.ValidationFailed, "selected text deadline is outside 1 to 1800 seconds")
	}
	limit := input.OutputLimit
	if limit == 0 {
		limit = 64 * 1024
	}
	if limit < 1 || limit > 128*1024 {
		return result, nativeError(apierrors.ValidationFailed, "selected text output limit is invalid")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()
	expectedConfig, err := r.ConfigurationID(ctx, input.CLI)
	if err != nil {
		return result, err
	}
	n, s, err := r.textSession(ctx, input.CLI, "", seconds)
	if err != nil {
		return result, err
	}
	defer func() {
		obs, err := n.Stop(context.Background(), s)
		if err != nil || !obs.StopConfirmed {
			resultErr = nativeError(apierrors.DeliveryUnknown, "selected-text owned process stop remains unconfirmed")
		}
	}()
	p, err := n.get(s)
	if err != nil {
		return result, err
	}
	actualConfig := n.processConfiguration(p)
	if actualConfig == "" {
		return result, nativeError(apierrors.EvidenceMissing, "native protocol does not report the current model before a paid turn; selected-text processing is unavailable")
	}
	if actualConfig != expectedConfig {
		return result, nativeError(apierrors.ContextStale, "native model configuration changed; refresh before processing")
	}
	if _, err := n.Send(ctx, s, packetPrompt(s.ContextPacket, input.Prompt)); err != nil {
		return result, err
	}
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
			p.mu.Lock()
			if p.model != "" {
				model := p.model
				result.Model = &model
			}
			p.mu.Unlock()
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

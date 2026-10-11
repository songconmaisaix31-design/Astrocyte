package worksscan

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

const maxDeliveryEntries = 256

// Delivery parses "AI 投递 JSON" envelopes. It is the adapter value wired into
// the works service; the free function below keeps the parser itself testable.
type Delivery struct{}

// ParseDelivery implements the application's WorksDeliverySource port.
func (Delivery) ParseDelivery(ctx context.Context, content []byte) ([]domain.Work, error) {
	return ParseDelivery(ctx, content)
}

// ParseDelivery parses a bounded "AI 投递 JSON" envelope into work
// observations. It performs no filesystem access: every Path is stored as the
// work root for later verification, never read here. Invalid entries are
// rejected with an error so a partial delivery never silently imports half a
// list.
func ParseDelivery(ctx context.Context, content []byte) ([]domain.Work, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var delivery domain.WorkDelivery
	if err := json.Unmarshal(content, &delivery); err != nil {
		return nil, err
	}
	if len(delivery.Works) > maxDeliveryEntries {
		return nil, errors.New("delivery has too many entries")
	}
	works := make([]domain.Work, 0, len(delivery.Works))
	for _, entry := range delivery.Works {
		if err := validateEntry(entry); err != nil {
			return nil, err
		}
		platform := strings.TrimSpace(entry.Platform)
		path := strings.TrimSpace(entry.Path)
		kind := domain.NormalizeWorkKind(entry.Kind)
		works = append(works, domain.Work{
			ID:       domain.WorkIdentity(platform, path),
			Platform: platform,
			Name:     strings.TrimSpace(entry.Name),
			Root:     path,
			Kind:     kind,
			Source:   domain.WorkSourceJSON,
			Attitude: entry.Attitude,
			Group:    entry.Group,
			Handoff:  domain.WorkHandoff{Content: entry.Handoff, Source: domain.HandoffSourceGenerated},
			Files:    domain.WorkFileFact{SourcePath: path, Verified: false},
		})
	}
	return works, nil
}

func validateEntry(entry domain.WorkDeliveryEntry) error {
	if !domain.ValidWorkName(entry.Name) {
		return errors.New("work name is empty or over 200 characters")
	}
	if !domain.ValidWorkPlatform(entry.Platform) {
		return errors.New("platform label is empty or over 64 characters")
	}
	if strings.TrimSpace(entry.Path) == "" {
		return errors.New("work path is required")
	}
	if entry.Attitude != "" && !domain.ValidAttitude(entry.Attitude) {
		return errors.New("attitude must be continue, park or done")
	}
	if !domain.ValidWorkGroup(entry.Group) {
		return errors.New("group label is over 200 characters")
	}
	if len(entry.Handoff) > 64*1024 {
		return errors.New("handoff manual is over 64 KiB")
	}
	return nil
}

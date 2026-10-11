package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type KnowledgeRepository interface {
	WithKnowledgeTx(context.Context, func(KnowledgeTx) error) error
}

// KnowledgeTx is also bound to the shared sleep transaction by the SQLite adapter.
type KnowledgeTx interface {
	LoadKnowledge(string) (StoredKnowledge, error)
	ListKnowledge(string, bool) ([]StoredKnowledge, error)
	SaveKnowledge(StoredKnowledge, int) error
	AddKnowledgeShare(string, ShareAudit) error
	AddKnowledgeAction(string, ActionAudit) error
	KnowledgeAttention(string) (domain.AttentionNode, error)
	KnowledgeSource(string) (StoredEvent, error)
	LoadKnowledgeReceipt(string, string, string) (Receipt, error)
	SaveKnowledgeReceipt(string, string, string, Receipt) error
}

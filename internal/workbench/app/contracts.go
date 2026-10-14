// Package app defines the read-only cognitive workbench projection contract.
// It has no persistence, transport or engine dependency.
package app

import (
	"context"
	"errors"
)

var ErrInvalidQuery = errors.New("invalid workbench query")
var ErrUnavailable = errors.New("workbench snapshot unavailable")

type Event struct {
	ID     string `json:"event_id"`
	Kind   string `json:"kind"`
	Topic  string `json:"topic"`
	Tokens *int64 `json:"tokens"`
	Scope  string `json:"scope"`
	TS     string `json:"ts"`
}
type Topic struct {
	Topic      string  `json:"topic"`
	Attention  float64 `json:"attention"`
	Momentum   float64 `json:"momentum"`
	Confidence float64 `json:"confidence"`
}
type Card struct {
	Topic
	Delta *float64 `json:"attention_delta"`
}
type Point struct {
	TS        string  `json:"ts"`
	Attention float64 `json:"attention"`
}
type Series struct {
	Topic  string  `json:"topic"`
	Points []Point `json:"points"`
}
type Health struct {
	Status     string `json:"status"`
	SnapshotAt string `json:"snapshot_at"`
	Timezone   string `json:"timezone"`
}
type Sleep struct {
	Day         string  `json:"day"`
	Status      string  `json:"status"`
	CompletedAt *string `json:"completed_at"`
}
type DayCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}
type SourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}
type Today struct {
	Day         string        `json:"day"`
	Focus       []Card        `json:"focus"`
	Activity    []Event       `json:"activity"`
	Sleep       Sleep         `json:"sleep"`
	DailyCounts []DayCount    `json:"daily_counts"`
	Sources     []SourceCount `json:"sources"`
}
type Focus struct {
	Topics []Topic  `json:"topics"`
	Series []Series `json:"series"`
}
type AgentGroup struct {
	Source string  `json:"source"`
	Today  int     `json:"today"`
	Days7  int     `json:"days7"`
	Days30 int     `json:"days30"`
	Recent []Event `json:"recent"`
}
type Agents struct {
	Groups []AgentGroup `json:"groups"`
}
type Events struct {
	Events   []Event `json:"events"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}
type Aggregate struct {
	Day   string `json:"day"`
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}
type Evidence struct {
	Topic      string      `json:"topic"`
	Aggregates []Aggregate `json:"aggregates"`
	Events
}
type Filters struct {
	Topics  []string `json:"topics"`
	Kinds   []string `json:"kinds"`
	Sources []string `json:"sources"`
}

type Query struct {
	Topic, Kind, Source, Start, End string
	Page, PageSize                  int
}

// View is a closed, typed projection union. Exactly one pointer is set for a
// successful read. Raw payloads cannot enter the shared serialization boundary.
type View struct {
	SnapshotAt string
	Health     *Health
	Today      *Today
	Focus      *Focus
	Agents     *Agents
	Events     *Events
	Evidence   *Evidence
	Filters    *Filters
}
type Reader interface {
	Read(context.Context, string, Query) (View, error)
}

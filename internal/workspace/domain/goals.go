package domain

import "time"

type Goal struct {
	ID              string    `json:"id"`
	Revision        int       `json:"revision"`
	ProjectID       string    `json:"project_id"`
	Title           string    `json:"title"`
	SuccessCriteria []string  `json:"success_criteria"`
	Priority        int       `json:"priority"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type CurrentGoalSnapshot struct {
	Goal                    Goal      `json:"goal"`
	ProjectSettingsRevision int       `json:"project_settings_revision"`
	ObservedAt              time.Time `json:"observed_at"`
}

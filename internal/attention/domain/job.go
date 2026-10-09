package domain

import "time"

// JobRules has no clock or storage dependency. Attempts and deadline are
// persisted by the application before external I/O begins.
type JobRules struct {
	State           string
	Attempts        int
	MaxAttempts     int
	Deadline        time.Time
	ExternalStarted bool
	DeliveryUnknown bool
}

func (j JobRules) Start(now time.Time) (JobRules, error) {
	if j.State != "queued" {
		return j, ErrVersion
	}
	if j.DeliveryUnknown {
		return j, ErrUnknown
	}
	if j.MaxAttempts < 1 || j.Attempts >= j.MaxAttempts || !now.Before(j.Deadline) {
		return j, ErrInvalid
	}
	j.State, j.Attempts = "running", j.Attempts+1
	return j, nil
}

func (j JobRules) Retry(now time.Time) (JobRules, error) {
	if j.DeliveryUnknown {
		return j, ErrUnknown
	}
	if j.State != "failed" && j.State != "cancelled" {
		return j, ErrVersion
	}
	if j.MaxAttempts < 1 || j.Attempts >= j.MaxAttempts || !now.Before(j.Deadline) {
		return j, ErrInvalid
	}
	j.State, j.ExternalStarted = "queued", false
	return j, nil
}

func (j JobRules) Recover() JobRules {
	if j.State == "running" {
		j.State = "failed"
		j.DeliveryUnknown = j.ExternalStarted
	}
	return j
}

func (j JobRules) Cancel() (JobRules, error) {
	if j.State != "queued" && j.State != "running" {
		return j, ErrVersion
	}
	j.DeliveryUnknown = j.DeliveryUnknown || (j.State == "running" && j.ExternalStarted)
	j.State = "cancelled"
	return j, nil
}

package app

import (
	"context"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestRecoveryReadOnlyVsUnknownModelAndDeadline(t *testing.T) {
	s, r, _ := fixture(t)
	now := s.options.Clock()
	for _, kind := range []string{"import", "distillation"} {
		r.state.Jobs[kind] = Job{SchemaVersion: 1, JobID: kind, Status: "running", Kind: kind, Version: 2, Attempts: 1, MaxAttempts: 3, DeadlineAt: now.Add(time.Hour), ExternalStarted: true, OperationID: "original-" + kind, Payload: []byte(`{"durable":true}`), Caller: human}
	}
	if err := s.RecoverJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	read, _ := s.GetJob(context.Background(), human, "import")
	model, _ := s.GetJob(context.Background(), human, "distillation")
	if read.Status != "failed" || read.DeliveryUnknown || model.Status != "failed" || !model.DeliveryUnknown || model.Error.Code != apierrors.DeliveryUnknown {
		t.Fatal("recovery collapsed different external effects")
	}
	if string(model.Payload) != `{"durable":true}` || model.Caller.ID != human.ID {
		t.Fatal("internal payload or caller lost on persistence")
	}
	_, err := s.RetryJob(context.Background(), human, model.JobID, meta("blind-resend", model.Version))
	errorCode(t, err, apierrors.DeliveryUnknown)
	if r.state.Jobs[model.JobID].OperationID != "original-distillation" {
		t.Fatal("unknown operation replaced")
	}
	s.options.Clock = func() time.Time { return now.Add(time.Hour) }
	_, err = s.RetryJob(context.Background(), human, read.JobID, meta("expired", read.Version))
	errorCode(t, err, apierrors.ValidationFailed)
}

func TestExplicitUnknownReadResultIsPreservedEvenForImport(t *testing.T) {
	s, _, _ := fixture(t)
	s.sources = sourceFunc(func(context.Context, ImportMaterialCommand) (ImportedSource, error) {
		return ImportedSource{}, serviceError(apierrors.DeliveryUnknown, "Unknown external adapter outcome", "reconcile")
	})
	response, err := s.ImportMaterial(context.Background(), human, importCommand("unknown", "source"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	job, _ := s.GetJob(context.Background(), human, response.JobID)
	if !job.DeliveryUnknown {
		t.Fatal("explicit unknown was replaced by ordinary read failure")
	}
	_, err = s.RetryJob(context.Background(), human, job.JobID, meta("retry", job.Version))
	errorCode(t, err, apierrors.DeliveryUnknown)
}

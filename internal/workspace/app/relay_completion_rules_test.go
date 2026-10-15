package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 缺失验证器和非法引用只产生拒绝，不构造控制器、证据或完成成功。
func TestRelayCompletionMissingGoalEvidenceCannotVerify(t *testing.T) {
	service := &relayService{options: RelayServiceOptions{Clock: time.Now}}
	for _, from := range []string{domain.RelayAccepted, domain.RelayWorking} {
		for _, reference := range []string{"", "private/path", "valid-reference-without-reader"} {
			result, verification, err := service.prepareCompletion(context.Background(), domain.Caller{Kind: "human"}, domain.RelayTask{State: from}, RelayTransitionCommand{From: from, To: domain.RelayDone, CompletionConfirmed: true, CompletionChannel: domain.RelayVerifiedComplete, CompletionEvidenceID: reference}, domain.RelayReceipt{})
			var failure *apierrors.ServiceError
			if !errors.As(err, &failure) || failure.Code != apierrors.EvidenceMissing || result.ID != "" || verification != nil {
				t.Fatal("absent owned goal evidence produced a verified completion")
			}
		}
	}
	failure := relayCompletionEvidenceError(&apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "PRIVATE_OUTPUT_CANARY", RequiredAction: "PRIVATE_PATH_CANARY"})
	if strings.Contains(failure.Error(), "PRIVATE_") {
		t.Fatal("verification failure disclosed original private output or paths")
	}
}

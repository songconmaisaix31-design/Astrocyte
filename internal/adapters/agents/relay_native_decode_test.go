package agents

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// 纯 JSON 输入检查，不创建控制器、进程或成功协议响应。
func TestRelayNativeDecodeRejectsMalformedValuesWithoutRawLeak(t *testing.T) {
	for index, input := range []string{"", "null", "  null  ", `{"private":"RAW_DECODE_CANARY",`, `{"thread":"RAW_DECODE_CANARY"}`, `{"thread":{"id":12}}`, `[]`, `true`, `{"thread":{"id":"public"}} trailing`} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			var value struct {
				Thread struct {
					ID string `json:"id"`
				} `json:"thread"`
			}
			err := decodeNativeValue(json.RawMessage(input), &value)
			var known *apierrors.ServiceError
			if !errors.As(err, &known) || known.Code != apierrors.DeliveryUnknown || known.Retryable || strings.Contains(err.Error(), "RAW_DECODE_CANARY") {
				t.Fatal("malformed value was accepted, retryable or leaked its raw content")
			}
		})
	}
	for _, input := range []string{"null", `"true"`, "{}", "0"} {
		var value bool
		if err := decodeNativeValue(json.RawMessage(input), &value); err == nil {
			t.Fatal("malformed boolean was coerced to a native success value")
		}
	}
}

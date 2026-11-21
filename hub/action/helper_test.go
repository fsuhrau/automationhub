package action

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device/appproto"
	"testing"
)

func TestPayloadValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got, err := payloadValue(ptr("x"), "value")
		noErr(t, err)
		mustEqual(t, "x", got)
	})
	t.Run("bool", func(t *testing.T) {
		got, err := payloadValue(ptr(true), "visible")
		noErr(t, err)
		mustEqual(t, true, got)
	})
	t.Run("missing field", func(t *testing.T) {
		got, err := payloadValue[appproto.Tests](nil, "tests")
		if err == nil || err.Error() != "response payload has no tests" {
			t.Fatalf("unexpected error: %v", err)
		}
		mustEqual(t, appproto.Tests{}, got)
	})
}

func TestSerializeKeepsTypeAndPayload(t *testing.T) {
	got, err := serialize(appproto.ActionType_Move, appproto.RequestData{})
	noErr(t, err)
	mustEqual(t, `{"actionID":"","actionType":5,"payload":{}}`, string(got))
}

func TestTouchPayload(t *testing.T) {
	mustEqual(t, appproto.RequestData{Touch: &appproto.Touch{Xoffset: 1, Yoffset: 2}}, touchPayload(1, 2))
}

func TestProcessResponseRejectsMalformedResponse(t *testing.T) {
	for _, a := range []Interface{&Custom{}, &GetValue{}, &IsDisplayed{}, &TestsGet{}} {
		t.Run(fmt.Sprintf("%T", a), func(t *testing.T) {
			if err := a.ProcessResponse(&appproto.Response{Success: true}); err == nil {
				t.Fatal("expected an error for a response without payload")
			}
		})
	}
}

package action

import (
	"flag"
	"github.com/fsuhrau/automationhub/device/appproto"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata") //nolint:gochecknoglobals // test flag

func noErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func ptr[T any](v T) *T { return &v }

type actionCase struct {
	name     string
	action   Interface
	typ      appproto.ActionType
	response appproto.Response
	check    func(t *testing.T, a Interface)
}

func actionCases() []actionCase {
	ok := func(p appproto.ResponseData) appproto.Response { return appproto.Response{Success: true, Payload: p} }
	return []actionCase{
		{"custom", &Custom{RequestData: `{"cmd":"x"}`}, appproto.ActionType_Custom,
			ok(appproto.ResponseData{Data: ptr([]byte("abc"))}),
			func(t *testing.T, a Interface) {
				mustEqual(t, []byte("abc"), a.(*Custom).Content)
				mustEqual(t, true, a.(*Custom).Success)
			}},
		{"drag_and_drop", &DragAndDrop{From: "a", To: "b"}, appproto.ActionType_DragAndDrop, ok(appproto.ResponseData{}), nil},
		{"get_value", &GetValue{ElementID: "7", Attr: "text"}, appproto.ActionType_ElementGetValue,
			ok(appproto.ResponseData{Value: ptr("hello")}),
			func(t *testing.T, a Interface) {
				mustEqual(t, "hello", a.(*GetValue).Value)
				mustEqual(t, true, a.(*GetValue).Success)
			}},
		{"is_displayed", &IsDisplayed{ElementID: "7"}, appproto.ActionType_ElementIsDisplayed,
			ok(appproto.ResponseData{Visible: ptr(true)}),
			func(t *testing.T, a Interface) {
				mustEqual(t, true, a.(*IsDisplayed).IsDisplayed)
				mustEqual(t, true, a.(*IsDisplayed).Success)
			}},
		{"set_value", &SetValue{ElementID: "7", Attr: "text", Value: "v"}, appproto.ActionType_ElementSetValue, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*SetValue).Success) }},
		{"touch_element", &TouchElement{ElementID: "7"}, appproto.ActionType_ElementTouch, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchElement).Success) }},
		{"longtouch_element", &LongTouchElement{ElementID: "7"}, appproto.ActionType_LongTouch, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*LongTouchElement).Success) }},
		{"get_screenshot", &GetScreenshot{}, appproto.ActionType_GetScreenshot,
			ok(appproto.ResponseData{Screenshot: &appproto.Screenshot{Sceengraph: []byte("<a/>"), ContentType: appproto.ContentType_Xml}}),
			func(t *testing.T, a Interface) {
				mustEqual(t, true, a.(*GetScreenshot).Success)
				mustEqual(t, []byte("<a/>"), a.(*GetScreenshot).SceengraphXML())
			}},
		{"scroll", &Scroll{OffsetX: 1, OffsetY: -2}, appproto.ActionType_Move, ok(appproto.ResponseData{}), nil},
		{"tests_get", &TestsGet{}, appproto.ActionType_GetTests,
			ok(appproto.ResponseData{Tests: &appproto.Tests{Tests: []appproto.Test{{Class: "C", Method: "M"}}}}),
			func(t *testing.T, a Interface) {
				mustEqual(t, []appproto.Test{{Class: "C", Method: "M"}}, a.(*TestsGet).Tests)
				mustEqual(t, true, a.(*TestsGet).Success)
			}},
		{"test_start", &TestStart{Assembly: "As", Class: "C", Method: "M", Env: map[string]string{"k": "v"}}, appproto.ActionType_ExecuteTest, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TestStart).Success) }},
		{"touch_down", &TouchDownPosition{PosX: 3, PosY: 4}, appproto.ActionType_TouchDown, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchDownPosition).Success) }},
		{"touch_up", &TouchUpPosition{PosX: 3, PosY: 4}, appproto.ActionType_TouchUp, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchUpPosition).Success) }},
		{"touch_move", &TouchMovePosition{PosX: 3, PosY: 4}, appproto.ActionType_TouchMove, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchMovePosition).Success) }},
		{"unity_reset", &UnityReset{}, appproto.ActionType_UnityReset, ok(appproto.ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*UnityReset).Success) }},
		{"get_sceengraph", &GetSceenGraph{}, appproto.ActionType_GetSceneGraph,
			ok(appproto.ResponseData{Screenshot: &appproto.Screenshot{Sceengraph: []byte("<a/>"), ContentType: appproto.ContentType_Xml}}),
			func(t *testing.T, a Interface) { mustEqual(t, "<a/>", a.(*GetSceenGraph).Content()) }},
	}
}

func TestActionsSerializeGolden(t *testing.T) {
	for _, tc := range actionCases() {
		t.Run(tc.name, func(t *testing.T) {
			mustEqual(t, tc.typ, tc.action.GetActionType())
			got, err := tc.action.Serialize()
			noErr(t, err)
			path := filepath.Join("testdata", tc.name+".golden")
			if *update {
				noErr(t, os.MkdirAll("testdata", 0o755))
				noErr(t, os.WriteFile(path, got, 0o644))
			}
			want, err := os.ReadFile(path)
			noErr(t, err)
			mustEqual(t, string(want), string(got))
		})
	}
}

func TestActionsProcessResponse(t *testing.T) {
	for _, tc := range actionCases() {
		t.Run(tc.name, func(t *testing.T) {
			resp := tc.response
			noErr(t, tc.action.ProcessResponse(&resp))
			if tc.check != nil {
				tc.check(t, tc.action)
			}
		})
	}
}

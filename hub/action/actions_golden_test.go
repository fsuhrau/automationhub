package action

import (
	"flag"
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
	typ      ActionType
	response Response
	check    func(t *testing.T, a Interface)
}

func actionCases() []actionCase {
	ok := func(p ResponseData) Response { return Response{Success: true, Payload: p} }
	return []actionCase{
		{"custom", &Custom{RequestData: `{"cmd":"x"}`}, ActionType_Custom,
			ok(ResponseData{Data: ptr([]byte("abc"))}),
			func(t *testing.T, a Interface) {
				mustEqual(t, []byte("abc"), a.(*Custom).Content)
				mustEqual(t, true, a.(*Custom).Success)
			}},
		{"drag_and_drop", &DragAndDrop{From: "a", To: "b"}, ActionType_DragAndDrop, ok(ResponseData{}), nil},
		{"get_value", &GetValue{ElementID: "7", Attr: "text"}, ActionType_ElementGetValue,
			ok(ResponseData{Value: ptr("hello")}),
			func(t *testing.T, a Interface) {
				mustEqual(t, "hello", a.(*GetValue).Value)
				mustEqual(t, true, a.(*GetValue).Success)
			}},
		{"is_displayed", &IsDisplayed{ElementID: "7"}, ActionType_ElementIsDisplayed,
			ok(ResponseData{Visible: ptr(true)}),
			func(t *testing.T, a Interface) {
				mustEqual(t, true, a.(*IsDisplayed).IsDisplayed)
				mustEqual(t, true, a.(*IsDisplayed).Success)
			}},
		{"set_value", &SetValue{ElementID: "7", Attr: "text", Value: "v"}, ActionType_ElementSetValue, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*SetValue).Success) }},
		{"touch_element", &TouchElement{ElementID: "7"}, ActionType_ElementTouch, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchElement).Success) }},
		{"longtouch_element", &LongTouchElement{ElementID: "7"}, ActionType_LongTouch, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*LongTouchElement).Success) }},
		{"get_screenshot", &GetScreenshot{}, ActionType_GetScreenshot,
			ok(ResponseData{Screenshot: &Screenshot{Sceengraph: []byte("<a/>"), ContentType: ContentType_Xml}}),
			func(t *testing.T, a Interface) {
				mustEqual(t, true, a.(*GetScreenshot).Success)
				mustEqual(t, []byte("<a/>"), a.(*GetScreenshot).SceengraphXML())
			}},
		{"scroll", &Scroll{OffsetX: 1, OffsetY: -2}, ActionType_Move, ok(ResponseData{}), nil},
		{"tests_get", &TestsGet{}, ActionType_GetTests,
			ok(ResponseData{Tests: &Tests{Tests: []Test{{Class: "C", Method: "M"}}}}),
			func(t *testing.T, a Interface) {
				mustEqual(t, []Test{{Class: "C", Method: "M"}}, a.(*TestsGet).Tests)
				mustEqual(t, true, a.(*TestsGet).Success)
			}},
		{"test_start", &TestStart{Assembly: "As", Class: "C", Method: "M", Env: map[string]string{"k": "v"}}, ActionType_ExecuteTest, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TestStart).Success) }},
		{"touch_down", &TouchDownPosition{PosX: 3, PosY: 4}, ActionType_TouchDown, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchDownPosition).Success) }},
		{"touch_up", &TouchUpPosition{PosX: 3, PosY: 4}, ActionType_TouchUp, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchUpPosition).Success) }},
		{"touch_move", &TouchMovePosition{PosX: 3, PosY: 4}, ActionType_TouchMove, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*TouchMovePosition).Success) }},
		{"unity_reset", &UnityReset{}, ActionType_UnityReset, ok(ResponseData{}),
			func(t *testing.T, a Interface) { mustEqual(t, true, a.(*UnityReset).Success) }},
		{"get_sceengraph", &GetSceenGraph{}, ActionType_GetSceneGraph,
			ok(ResponseData{Screenshot: &Screenshot{Sceengraph: []byte("<a/>"), ContentType: ContentType_Xml}}),
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

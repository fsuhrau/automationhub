package androiddevice

import (
	"strings"
	"testing"
)

func TestNativeScriptParser(t *testing.T) {
	script := `
	wait_for xpath:"//node[@class='android.widget.Button'][@text='Agree']" timeout:20;
	click xpath:"//node[@class='android.widget.Button'][@text='Agree']";
	wait_for xpath:"//node[@class='android.widget.Button'][@text='1-tap buy']" timeout:20;
	click xpath:"//node[@class='android.widget.Button'][@text='1-tap buy']";`

	actions, err := ParseNativeScript([]byte(script))
	if err != nil {
		t.Errorf("parse native script failed: %v", err)
	}

	if len(actions) != 4 {
		t.Errorf("actions mismatch expected 4 parsed %d", len(actions))
	}
}

func TestNativeScriptParserErrors(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{"wait_for format", `wait_for xpath:"//a";`, "invalid format for wait_for"},
		{"click format", `click //a`, "invalid format for click"},
		{"unknown action", `swipe up;`, "unsupported native script"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseNativeScript([]byte(tt.script)); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestNativeScriptActions(t *testing.T) {
	actions, err := ParseNativeScript([]byte("\nwait_for xpath:\"//a\" timeout:7;\n\nclick xpath:\"//b\";\n"))
	if err != nil || len(actions) != 2 {
		t.Fatalf("actions %v err %v", actions, err)
	}
	wait, click := actions[0].(*WaitForAction), actions[1].(*ClickAction)
	if wait.XPath != "//a" || wait.Timeout != 7 || wait.GetAction() != "wait_for" || click.XPath != "//b" || click.GetAction() != "click" {
		t.Fatalf("wait %+v click %+v", wait, click)
	}
}

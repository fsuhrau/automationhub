package androiddevice

import "testing"

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
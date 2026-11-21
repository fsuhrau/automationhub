package device

import "testing"

func TestPropertiesAreValid(t *testing.T) {
	if (&Properties{}).AreValid() {
		t.Error("properties without an app are valid")
	}
	if !(&Properties{App: "game"}).AreValid() {
		t.Error("properties with an app are invalid")
	}
}

package iosdevice

import "testing"

func TestNativeScriptParser(t *testing.T) {
	script := `
			hub.waitFor("xpath", "//*[@name='Kaufen']", 20)
			hub.click("xpath", "//*[@name='Kaufen']")
			if hub.exists("xpath", "//*[@name='Anmelden']", 10)
			then
				hub.sendKeys("xxx")
				hub.click("xpath", "//*[@name='Anmelden']")
			end
			hub.waitForAlertWith("[Environment: Sandbox]", 60)
			hub.waitFor("xpath", "//*[@name='OK']", 20)
			hub.click("xpath", "//*[@name='OK']")
		`
	scriptHandler := New(nil, nil)
	if err := scriptHandler.Execute(script); err != nil {
		t.Errorf("luascript error %v", err)
	}
}

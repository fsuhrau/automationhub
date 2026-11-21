package webdriver

import "net/http"

type AppInfoValue struct {
	ProcessArguments struct {
		Env struct {
		} `json:"env"`
		Args []interface{} `json:"args"`
	} `json:"processArguments"`
	Name     string `json:"name"`
	Pid      int    `json:"pid"`
	BundleId string `json:"bundleId"`
}

type AppInfo struct {
	Value     AppInfoValue `json:"value"`
	SessionId string       `json:"sessionId"`
}

// Launch starts the app; arguments are passed to it as launch arguments.
func (c *Client) Launch(bundleIdentifier string, shouldWaitForQuiescence bool, arguments []string) error {
	shouldWait := 0
	if shouldWaitForQuiescence {
		shouldWait = 1
	}
	body := map[string]any{
		"bundleId":                bundleIdentifier,
		"shouldWaitForQuiescence": shouldWait,
		"arguments":               arguments,
	}
	return c.call(call{"launch", http.MethodPost, c.getSessionUrl("/wda/apps/launch"), body}, nil)
}

// Terminate stops the app.
func (c *Client) Terminate(bundleIdentifier string) error {
	body := map[string]string{"bundleId": bundleIdentifier}
	return c.call(call{"terminate", http.MethodPost, c.getSessionUrl("/wda/apps/terminate"), body}, nil)
}

// ActiveAppInfo describes the app in the foreground.
func (c *Client) ActiveAppInfo() (*AppInfo, error) {
	var res AppInfo
	if err := c.call(call{"get app info", http.MethodGet, c.getSessionUrl("/wda/activeAppInfo"), nil}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

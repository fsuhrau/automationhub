package webdriver

// SendText types the text into the focused element, one key per character.
func (c *Client) SendText(value string) error {
	keys := make([]string, 0, len(value))
	for _, r := range value {
		keys = append(keys, string(r))
	}
	return c.post("send keys", "/wda/keys", map[string][]string{"value": keys})
}

// PressButton presses a hardware button: "home", "volumeUp" or "volumeDown".
func (c *Client) PressButton(name string) error {
	return c.post("press button", "/wda/pressButton", map[string]string{"name": name})
}

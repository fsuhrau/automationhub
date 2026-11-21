package webdriver

import "net/http"

type SettingValue struct {
	ShouldUseCompactResponses    bool   `json:"shouldUseCompactResponses"`
	ElementResponseAttributes    string `json:"elementResponseAttributes"`
	MjpegServerScreenshotQuality int    `json:"mjpegServerScreenshotQuality"`
	MjpegServerFramerate         int    `json:"mjpegServerFramerate"`
	MjpegScalingFactor           int    `json:"mjpegScalingFactor"`
	ScreenshotQuality            int    `json:"screenshotQuality"`
	KeyboardAutocorrection       int    `json:"keyboardAutocorrection"`
	KeyboardPrediction           bool   `json:"keyboardPrediction"`
	SnapshotTimeout              int    `json:"snapshotTimeout"`
	CustomSnapshotTimeout        int    `json:"customSnapshotTimeout"`
	SnapshotMaxDepth             int    `json:"snapshotMaxDepth"`
	UseFirstMatch                bool   `json:"useFirstMatch"`
	BoundElementsByIndex         bool   `json:"boundElementsByIndex"`
	ReduceMotion                 bool   `json:"reduceMotion"`
	DefaultActiveApplication     string `json:"defaultActiveApplication"`
	ActiveAppDetectionPoint      string `json:"activeAppDetectionPoint"`
	IncludeNonModalElements      bool   `json:"includeNonModalElements"`
	DefaultAlertAction           string `json:"defaultAlertAction"`
	AcceptAlertButtonSelector    string `json:"acceptAlertButtonSelector"`
	DismissAlertButtonSelector   string `json:"dismissAlertButtonSelector"`
	ScreenshotOrientation        string `json:"screenshotOrientation"`
	WaitForIdleTimeout           int    `json:"waitForIdleTimeout"`
	AnimationCoolOffTimeout      int    `json:"animationCoolOffTimeout"`
}

type Settings struct {
	SessionID string                 `json:"sessionId"`
	Value     map[string]interface{} `json:"value"`
}

// GetSettings reads the Appium settings of the session.
func (c *Client) GetSettings() (*Settings, error) {
	var settings Settings
	if err := c.call(call{"get settings", http.MethodGet, c.getSessionUrl("/appium/settings"), nil}, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// SetSettings writes the Appium settings of the session.
func (c *Client) SetSettings(settings Settings) error {
	return c.post("set settings", "/appium/settings", settings)
}

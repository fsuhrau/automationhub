package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

func (c *Client) GetSettings() (*Settings, error) {
	req, err := http.NewRequest("GET", c.getSessionUrl("/appium/settings"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var settings Settings
	if err := json.Unmarshal(body, &settings); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(body, &wdaError)
		return nil, fmt.Errorf("unable to get settings: %s", wdaError.Value.Error)
	}
	return &settings, err
}

func (c *Client) SetSettings(settings Settings) error {
	jsonRequest, _ := json.Marshal(settings)
	body := bytes.NewBuffer(jsonRequest)
	req, err := http.NewRequest("POST", c.getSessionUrl("/appium/settings"), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}

	return err
}

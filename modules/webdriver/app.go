package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func (c *Client) Launch(bundleIdentifier string, shouldWaitForQuiescence bool, arguments []string) error {
	type request struct {
		ShouldWaitForQuiescence int      `json:"shouldWaitForQuiescence"`
		BundleID                string   `json:"bundleId"`
		Arguments               []string `json:"arguments"`
	}

	type response struct {
		SessionID string       `json:"sessionId"`
		Value     SessionValue `json:"value"`
	}

	shouldWait := 0
	if shouldWaitForQuiescence {
		shouldWait = 1
	}

	jsonBody, _ := json.Marshal(request{
		BundleID:                bundleIdentifier,
		ShouldWaitForQuiescence: shouldWait,
		Arguments:               arguments,
	})

	requestBody := bytes.NewBuffer(jsonBody)
	req, err := http.NewRequest("POST", c.getSessionUrl("/wda/apps/launch"), requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	var res response
	if err := json.Unmarshal(respBody, &res); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		return fmt.Errorf("unable to launch: %s", wdaError.Value.Error)
	}
	return nil
}

func (c *Client) Terminate(bundleIdentifier string) error {
	type request struct {
		BundleID string `json:"bundleId"`
	}

	type response struct {
		SessionID string      `json:"sessionId"`
		Value     interface{} `json:"value"`
	}

	jsonBody, _ := json.Marshal(request{
		BundleID: bundleIdentifier,
	})

	requestBody := bytes.NewBuffer(jsonBody)
	req, err := http.NewRequest("POST", c.getSessionUrl("/wda/apps/terminate"), requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	var r response
	if err := json.Unmarshal(respBody, &r); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		return fmt.Errorf("unable to terminate: %s", wdaError.Value.Error)
	}
	return nil
}

func (c *Client) ActiveAppInfo() (*AppInfo, error) {

	req, err := http.NewRequest("GET", c.getSessionUrl("/wda/activeAppInfo"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	var res AppInfo
	if err := json.Unmarshal(respBody, &res); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		return nil, fmt.Errorf("unable to get app info: %s", wdaError.Value.Error)
	}
	return &res, nil
}

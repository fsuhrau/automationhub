package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/session/%s/wda/apps/launch", c.address, c.sessionId), requestBody)
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
	respBody, _ := ioutil.ReadAll(resp.Body)

	var res response
	return json.Unmarshal(respBody, &res)
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/session/%s/wda/apps/terminate", c.address, c.sessionId), requestBody)
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
	respBody, _ := ioutil.ReadAll(resp.Body)

	var res response
	return json.Unmarshal(respBody, &res)
}

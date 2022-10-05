package webdriver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (c *Client) GetAlertText() (string, error) {

	type response struct {
		SessionId string `json:"sessionId"`
		Value     string `json:"value"`
	}

	req, err := http.NewRequest("GET", c.getSessionUrl("/alert/text"), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)

	var r response
	if err := json.Unmarshal(responseBody, &r); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(responseBody, &wdaError)
		return "", fmt.Errorf("unable to get alert text: %s", wdaError.Value.Error)
	}
	fmt.Println(string(responseBody))
	if len(r.Value) == 0 {
		return "", fmt.Errorf("there is no alert text")
	}
	return r.Value, err
}

func (c *Client) AcceptAlert(accept bool) (string, error) {

	type response struct {
		SessionId string `json:"sessionId"`
		Value     string `json:"value"`
	}

	action := "accept"
	if !accept {
		action = "dismiss"
	}

	req, err := http.NewRequest("POST", c.getSessionUrl("/alert/%s", action), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)

	var r response
	if err := json.Unmarshal(responseBody, &r); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(responseBody, &wdaError)
		return "", fmt.Errorf("unable to accept/dissmiss alert: %s", wdaError.Value.Error)
	}
	return r.Value, err
}

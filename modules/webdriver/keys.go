package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) SendText(value string) error {
	type request struct {
		Value []string `json:"value"`
	}

	var v []string
	for i := range value {
		v = append(v, string(value[i]))
	}
	requestJson, _ := json.Marshal(request{
		Value: v,
	})
	requestBody := bytes.NewBuffer(requestJson)

	req, err := http.NewRequest("POST", c.getSessionUrl("/wda/keys"), requestBody)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result Element
	if err := json.Unmarshal(respBody, &result); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		return fmt.Errorf("unable to send keys: %s", wdaError.Value.Error)
	}
	return nil
}

func (c *Client) PressButton(name string) error {
	type request struct {
		Name []byte `json:"name"`
	}
	requestJson, _ := json.Marshal(request{
		Name: []byte(name),
	})
	requestBody := bytes.NewBuffer(requestJson)

	req, err := http.NewRequest("POST", c.getSessionUrl("/wda/pressButton"), requestBody)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result Element
	if err := json.Unmarshal(respBody, &result); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		return fmt.Errorf("unable to press button: %s", wdaError.Value.Error)
	}

	return nil
}

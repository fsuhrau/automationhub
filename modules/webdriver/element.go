package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ElementValue struct {
	Element string `json:"ELEMENT"`
}

type Element struct {
	SessionId string        `json:"sessionId"`
	Value     *ElementValue `json:"value"`
}

type ElementButtons struct {
	SessionId string   `json:"sessionId"`
	Value     []string `json:"value"`
}

func (c *Client) FindElement(by, value string) (*Element, error) {
	type request struct {
		Using string `json:"using"`
		Value string `json:"value"`
	}
	requestJson, _ := json.Marshal(request{
		Using: by,
		Value: value,
	})
	requestBody := bytes.NewBuffer(requestJson)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/session/%s/element", c.address, c.sessionId), requestBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	var element Element
	if err := json.Unmarshal(body, &element); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(body, &wdaError)
		return nil, fmt.Errorf("unable to find element: %s", wdaError.Value.Error)
	}
	if element.Value == nil || len(element.Value.Element) == 0 {
		return nil, fmt.Errorf("unable to find element")
	}

	return &element, err
}

func (c *Client) TapElement(elementId string) error {
	// Create request
	req, err := http.NewRequest("POST", c.getSessionUrl("/element/%s/click", elementId), nil)

	// Headers
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return err
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	var result Element
	if err := json.Unmarshal(respBody, &result); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		return fmt.Errorf("unable to tap element: %s", wdaError.Value.Error)
	}
	return nil
}

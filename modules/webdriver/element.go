package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ElementValue struct {
	Element string `json:"ELEMENT"`
}

type Element struct {
	SessionId string        `json:"sessionId"`
	Error     string        `json:"error"`
	Value     *ElementValue `json:"value"`
}

type ElementText struct {
	SessionId string `json:"sessionId"`
	Value     string `json:"value"`
}

type ElementButtons struct {
	SessionId string   `json:"sessionId"`
	Value     []string `json:"value"`
}

// find element name: PKCompactNavigationWrapperView
// find element name: Kaufen
// tap element OK
// find element class name: XCUIElementTypeSecureTextField // Password Field
// find element name: Anmelden
// tap element Anmelden
// /alert/text "Du bist jetzt startklar...."
// find element name: OK
// tap element ok
{
"value": "Du bist jetzt startklar\nDein Kauf war erfolgreich.\n\n[Environment: Sandbox]",
"sessionId": "ED5E22E0-31AD-40E8-96DD-325EB08FE49C"
}

// /wda/alert/buttons
{
"value": [
"OK"
],
"sessionId": "ED5E22E0-31AD-40E8-96DD-325EB08FE49C"
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

	body, _ := ioutil.ReadAll(resp.Body)

	var settings Element
	err = json.Unmarshal(body, &settings)
	return &settings, err
}

func sendFindElement() {
	// Find Element (POST http://169.254.208.38:8100/session/ED5E22E0-31AD-40E8-96DD-325EB08FE49C/element)

	json := []byte(`{"using": "name","value": "PKCompactNavigationWrapperView"}`)
	body := bytes.NewBuffer(json)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", "http://169.254.208.38:8100/session/ED5E22E0-31AD-40E8-96DD-325EB08FE49C/element", body)

	// Headers
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
}
{
"value": null,
"sessionId": "ED5E22E0-31AD-40E8-96DD-325EB08FE49C"
}
func sendTapElement() {
	// Tap Element (POST http://169.254.208.38:8100/session/ED5E22E0-31AD-40E8-96DD-325EB08FE49C/element/04010000-0000-0000-6812-000000000000/click)

	json := []byte(`{"using": "name","value": "Kaufen"}`)
	body := bytes.NewBuffer(json)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", "http://169.254.208.38:8100/session/ED5E22E0-31AD-40E8-96DD-325EB08FE49C/element/04010000-0000-0000-6812-000000000000/click", body)

	// Headers
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
}

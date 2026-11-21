package webdriver

import (
	"errors"
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

// FindElement looks an element up, by is the locator strategy ("name",
// "class name", …).
func (c *Client) FindElement(by, value string) (*Element, error) {
	body := map[string]string{"using": by, "value": value}
	var element Element
	if err := c.call(call{"find element", http.MethodPost, c.getSessionUrl("/element"), body}, &element); err != nil {
		return nil, err
	}
	if element.Value == nil || len(element.Value.Element) == 0 {
		return nil, errors.New("unable to find element")
	}
	return &element, nil
}

// TapElement clicks the element.
func (c *Client) TapElement(elementId string) error {
	return c.call(call{"tap element", http.MethodPost, c.getSessionUrl("/element/%s/click", elementId), nil}, nil)
}

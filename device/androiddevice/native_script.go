package androiddevice

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	waitForEx = regexp.MustCompile(`wait_for xpath:"(.+)" timeout:([0-9]+);`)
	clickEx   = regexp.MustCompile(`click xpath:"(.+)";`)
)

type NativeAction interface {
	GetAction() string
}

type WaitForAction struct {
	XPath   string
	Timeout int64
}

func (a *WaitForAction) GetAction() string {
	return "wait_for"
}

type ClickAction struct {
	XPath string
}

func (a *ClickAction) GetAction() string {
	return "click"
}

// ParseNativeScript parses the lines of a native script into actions.
func ParseNativeScript(data []byte) ([]NativeAction, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var actions []NativeAction
	for scanner.Scan() {
		content := strings.TrimSpace(scanner.Text())
		if content == "" {
			continue
		}
		action, err := parseNativeLine(content)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func parseNativeLine(content string) (NativeAction, error) {
	switch {
	case strings.HasPrefix(content, "wait_for"):
		return parseWaitFor(content)
	case strings.HasPrefix(content, "click"):
		return parseClick(content)
	}
	return nil, fmt.Errorf("unsupported native script:  %s", content)
}

func parseWaitFor(content string) (NativeAction, error) {
	m := waitForEx.FindStringSubmatch(content)
	if m == nil {
		return nil, fmt.Errorf("invalid format for wait_for")
	}
	timeout, _ := strconv.ParseInt(m[2], 10, 64)
	return &WaitForAction{XPath: m[1], Timeout: timeout}, nil
}

func parseClick(content string) (NativeAction, error) {
	m := clickEx.FindStringSubmatch(content)
	if m == nil {
		return nil, fmt.Errorf("invalid format for click")
	}
	return &ClickAction{XPath: m[1]}, nil
}

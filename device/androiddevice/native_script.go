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
	clickEx = regexp.MustCompile(`click xpath:"(.+)";`)
)

type NativeAction interface {
	GetAction() string
}

type WaitForAction struct {
	XPath string
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

func ParseNativeScript(data []byte) ([]NativeAction, error) {
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	var actions []NativeAction
	_ = actions
	for scanner.Scan() {
		content := strings.TrimSpace(scanner.Text())
		if len(content) == 0 {
			continue
		}
		if strings.HasPrefix(content, "wait_for") {
			actionContent := waitForEx.FindAllStringSubmatch(content, -1)
			if len(actionContent) == 0 {
				return nil, fmt.Errorf("invalid format for wait_for")
			}
			timeout, _ := strconv.ParseInt(actionContent[0][2], 10, 64)
			actions = append(actions, &WaitForAction{
				XPath:   actionContent[0][1],
				Timeout: timeout,
			})
		}  else if strings.HasPrefix(content, "click") {
			actionContent := clickEx.FindAllStringSubmatch(content, -1)
			if len(actionContent) == 0 {
				return nil, fmt.Errorf("invalid format for click")
			}
			actions = append(actions, &ClickAction{
				XPath:   actionContent[0][1],
			})
		} else {
			return nil, fmt.Errorf("unsupported native script:  %s", content)
		}
	}
	return actions, nil
}
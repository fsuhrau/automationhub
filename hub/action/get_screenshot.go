package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

type GetScreenshot struct {
	Success    bool
	screenshot *Screenshot
	width      int
	height     int
}

func (a *GetScreenshot) GetActionType() ActionType {
	return ActionType_GetScreenshot
}

func (a *GetScreenshot) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetScreenshot,
	}
	return json.Marshal(req)
}

func (a *GetScreenshot) ProcessResponse(response *Response) error {
	a.Success = response.Success
	a.screenshot = response.Payload.Screenshot
	if a.screenshot != nil && len(a.screenshot.Screenshot) > 0 {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(a.screenshot.Screenshot))
		if err != nil {
			return fmt.Errorf("decode screenshot: %w", err)
		}
		a.width = cfg.Width
		a.height = cfg.Height
	}
	return nil
}

func (a *GetScreenshot) SceengraphXML() []byte {
	if a.screenshot == nil {
		return nil
	}
	var content []byte
	switch a.screenshot.ContentType {
	case ContentType_Flatbuffer:
		content = convertFlatToXml(a.screenshot.Sceengraph, true)
		break
	case ContentType_Json:
		content = a.screenshot.Sceengraph
		break
	case ContentType_Xml:
		content = a.screenshot.Sceengraph
		break
	}
	return content
}

func (a *GetScreenshot) ScreenshotData() []byte {
	if a.screenshot == nil {
		return nil
	}
	return a.screenshot.Screenshot
}

func (a *GetScreenshot) Width() int {
	return a.width
}

func (a *GetScreenshot) Height() int {
	return a.height
}

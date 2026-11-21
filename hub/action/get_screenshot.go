package action

import (
	"bytes"
	"fmt"
	"github.com/fsuhrau/automationhub/device/appproto"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

type GetScreenshot struct {
	Success    bool
	screenshot *appproto.Screenshot
	width      int
	height     int
}

func (a *GetScreenshot) GetActionType() appproto.ActionType {
	return appproto.ActionType_GetScreenshot
}

func (a *GetScreenshot) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_GetScreenshot, appproto.RequestData{})
}

func (a *GetScreenshot) ProcessResponse(response *appproto.Response) error {
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
	case appproto.ContentType_Flatbuffer:
		content = convertFlatToXml(a.screenshot.Sceengraph, true)
	case appproto.ContentType_Json:
		content = a.screenshot.Sceengraph
	case appproto.ContentType_Xml:
		content = a.screenshot.Sceengraph
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

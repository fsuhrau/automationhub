package action

import (
	"bytes"
	"encoding/json"
	"image"
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
	if a.screenshot != nil {
		reader := bytes.NewReader(a.screenshot.Screenshot)
		srcImage, _, _ := image.Decode(reader)
		a.width = srcImage.Bounds().Dx()
		a.height = srcImage.Bounds().Dy()
	}
	return nil
}

func (a *GetScreenshot) SceengraphXML() []byte {
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
	return a.screenshot.Screenshot
}

func (a *GetScreenshot) Width() int {
	return a.width
}

func (a *GetScreenshot) Height() int {
	return a.height
}

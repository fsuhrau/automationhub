package action

import (
	"bytes"
	"google.golang.org/protobuf/proto"
	"image"
)

type GetScreenshot struct {
	Success    bool
	screenshot *Screenshot
	width      int
	height     int
}

func (a *GetScreenshot) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetScreenshot,
	}
	return proto.Marshal(req)
}

func (a *GetScreenshot) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	a.screenshot = resp.GetScreenshot()
	if a.screenshot != nil {
		reader := bytes.NewReader(a.screenshot.Screenshot)
		srcImage, _, _ := image.Decode(reader)
		a.width = srcImage.Bounds().Dx()
		a.height = srcImage.Bounds().Dy()
	}
	return nil
}

func (a *GetScreenshot) SceengraphXML() []byte {
	return convertFlatToXml(a.screenshot.Sceengraph, true)
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

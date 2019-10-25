package action

import (
	"google.golang.org/protobuf/proto"
)

type GetScreenshot struct {
	Success bool
	screenshot *Screenshot
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
	return nil
}

func (a *GetScreenshot) SceengraphXML() []byte {
	return convertFlatToXml(a.screenshot.Sceengraph, true)
}

func (a *GetScreenshot) ScreenshotData() []byte {
	return a.screenshot.Screenshot
}

package action

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/hub/networking"
	"google.golang.org/protobuf/proto"
)

type GetSceenGraph struct {
	content []byte
	contentType ContentType
}

type Node struct {
	XMLName    xml.Name `xml:""`
	Class      string   `xml:"Class,attr"`
	Id         int32    `xml:"ID,attr"`
	Name       string   `xml:"Name,attr"`
	CSS        string   `xml:"CSS,attr"`
	X          int32    `xml:"X,attr"`
	Y          int32    `xml:"Y,attr"`
	RectangleX int32    `xml:"RectangleX,attr"`
	RectangleY int32    `xml:"RectangleY,attr"`
	IsVisible  byte     `xml:"isVisible,attr"`
	LabelText  string   `xml:"LabelText,attr"`
	Children   []Node   `xml:""`
}


func nodeIteratorFunc(flatNode *networking.Node) Node {
	node := Node{
		XMLName:    xml.Name{Local: flatNode.Class()},
		Id:         flatNode.ID(),
		Class:      flatNode.Class(),
		Name:       flatNode.Name(),
		CSS:        flatNode.CSS(),
		X:          flatNode.X(),
		Y:          flatNode.Y(),
		RectangleX: flatNode.RectangleX(),
		RectangleY: flatNode.RectangleY(),
		IsVisible:  flatNode.IsVisible(),
		LabelText:  flatNode.LabelText(),
	}

	var child networking.Node
	for i := 0; i < flatNode.ChildrenLength(); i++ {
		if flatNode.Children(&child, i) {
			node.Children = append(node.Children, nodeIteratorFunc(&child))
		}
	}
	return node
}

func convertFlatToXml(content []byte, pretty bool) []byte {
	node := networking.GetRootAsNode(content, 0)
	xmlNode := nodeIteratorFunc(node)
	var output []byte
	if pretty {
		output, _ = xml.MarshalIndent(xmlNode, "", " ")
	} else {
		output, _ = xml.Marshal(xmlNode)
	}
	//fmt.Printf("\n\n%v\n\n", string(output))
	return output
}

func (a *GetSceenGraph) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetSceneGraph,
	}
	return proto.Marshal(req)
}

func (a *GetSceenGraph) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	if resp.GetScreenshot() != nil {
		a.content = resp.GetScreenshot().Sceengraph
		a.contentType = resp.GetScreenshot().ContentType
	}
	return nil
}

func (a *GetSceenGraph) Content() string {
	result := ""

	if len(a.content) == 0 {
		return result
	}

	switch a.contentType {
	case ContentType_Flatbuffer:
		result = string(convertFlatToXml(a.content, true))
		break
	case ContentType_Json:
		result = string(a.content)
		break
	case ContentType_Xml:
		result = string(a.content)
		break
	}

	return result
}

func (a *GetSceenGraph) XML() (*xmlquery.Node, error) {
	if len(a.content) == 0 {
		return nil, fmt.Errorf("can't convert sceengraph because no content available")
	}
	xmlDocString := convertFlatToXml(a.content, false)
	return xmlquery.Parse(bytes.NewReader(xmlDocString))
}

package action

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/hub/networking"
)

type GetSceenGraph struct {
	content     []byte
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

func (a *GetSceenGraph) GetActionType() ActionType {
	return ActionType_GetSceneGraph
}

func (a *GetSceenGraph) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetSceneGraph,
	}
	return json.Marshal(req)
}

func (a *GetSceenGraph) ProcessResponse(response *Response) error {
	if response.Payload.Screenshot != nil {
		a.content = response.Payload.Screenshot.Sceengraph
		a.contentType = response.Payload.Screenshot.ContentType
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

// XMLBytes returns the scene graph as XML. JSON scene graphs can't be
// converted and return an error.
func (a *GetSceenGraph) XMLBytes(pretty bool) ([]byte, error) {
	if len(a.content) == 0 {
		return nil, fmt.Errorf("can't convert sceengraph because no content available")
	}
	switch a.contentType {
	case ContentType_Flatbuffer:
		return convertFlatToXml(a.content, pretty), nil
	case ContentType_Xml:
		return a.content, nil
	}
	return nil, fmt.Errorf("scene graph content type %v is not xml", a.contentType)
}

func (a *GetSceenGraph) XML() (*xmlquery.Node, error) {
	content, err := a.XMLBytes(false)
	if err != nil {
		return nil, err
	}
	return xmlquery.Parse(bytes.NewReader(content))
}

// FindElementID returns the ID attribute of the first element matching the
// XPath expression, preferring visible elements.
func (a *GetSceenGraph) FindElementID(xpath string) (string, error) {
	doc, err := a.XML()
	if err != nil {
		return "", err
	}
	nodes, err := xmlquery.QueryAll(doc, xpath)
	if err != nil {
		return "", fmt.Errorf("invalid xpath %q: %w", xpath, err)
	}
	var fallback string
	for _, n := range nodes {
		id := n.SelectAttr("ID")
		if id == "" {
			continue
		}
		if n.SelectAttr("isVisible") != "0" {
			return id, nil
		}
		if fallback == "" {
			fallback = id
		}
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("no element matches %q", xpath)
}

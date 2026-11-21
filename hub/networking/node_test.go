package networking

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
)

// buildNode encodes a node with a child through the generated builders.
func buildNode() []byte {
	b := flatbuffers.NewBuilder(0)
	child := func() flatbuffers.UOffsetT {
		name := b.CreateString("child")
		NodeStart(b)
		NodeAddName(b, name)
		NodeAddID(b, 2)
		return NodeEnd(b)
	}()
	NodeStartChildrenVector(b, 1)
	b.PrependUOffsetT(child)
	children := b.EndVector(1)

	class, name, css, label := b.CreateString("Button"), b.CreateString("play"), b.CreateString("btn"), b.CreateString("Play")
	NodeStart(b)
	NodeAddClass(b, class)
	NodeAddID(b, 1)
	NodeAddName(b, name)
	NodeAddCSS(b, css)
	NodeAddX(b, 10)
	NodeAddY(b, 20)
	NodeAddRectangleX(b, 30)
	NodeAddRectangleY(b, 40)
	NodeAddIsVisible(b, 1)
	NodeAddLabelText(b, label)
	NodeAddChildren(b, children)
	b.Finish(NodeEnd(b))
	return b.FinishedBytes()
}

func TestNodeRoundTrip(t *testing.T) {
	node := GetRootAsNode(buildNode(), 0)

	if string(node.Class()) != "Button" || node.ID() != 1 || string(node.Name()) != "play" || string(node.CSS()) != "btn" {
		t.Errorf("identity %q %d %q %q", node.Class(), node.ID(), node.Name(), node.CSS())
	}
	if node.X() != 10 || node.Y() != 20 || node.RectangleX() != 30 || node.RectangleY() != 40 || node.IsVisible() != 1 {
		t.Errorf("geometry %d %d %d %d %d", node.X(), node.Y(), node.RectangleX(), node.RectangleY(), node.IsVisible())
	}
	if string(node.LabelText()) != "Play" || node.ChildrenLength() != 1 {
		t.Errorf("label %q, %d children", node.LabelText(), node.ChildrenLength())
	}
	var child Node
	if !node.Children(&child, 0) || child.ID() != 2 || string(child.Name()) != "child" {
		t.Errorf("child %d %q", child.ID(), child.Name())
	}
	if !node.Children(nil, 0) {
		t.Error("a child can be read without an object to fill")
	}
}

func TestNodeWithoutChildren(t *testing.T) {
	b := flatbuffers.NewBuilder(0)
	NodeStart(b)
	b.Finish(NodeEnd(b))
	node := GetRootAsNode(b.FinishedBytes(), 0)

	if node.ChildrenLength() != 0 || node.Children(nil, 0) {
		t.Fatal("a node without children has none")
	}
}

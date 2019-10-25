// automatically generated, do not modify

package networking

import (
	flatbuffers "github.com/google/flatbuffers/go"
)
type Node struct {
	_tab flatbuffers.Table
}

func GetRootAsNode(buf []byte, offset flatbuffers.UOffsetT) *Node {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Node{}
	x.Init(buf, n + offset)
	return x
}

func (rcv *Node) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Node) Class() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *Node) ID() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Node) Name() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *Node) CSS() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *Node) X() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Node) Y() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Node) RectangleX() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Node) RectangleY() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Node) IsVisible() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Node) LabelText() string {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		return rcv._tab.String(o + rcv._tab.Pos)
	}
	return ""
}

func (rcv *Node) Children(obj *Node, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
	if obj == nil {
		obj = new(Node)
	}
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Node) ChildrenLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func NodeStart(builder *flatbuffers.Builder) { builder.StartObject(11) }
func NodeAddClass(builder *flatbuffers.Builder, Class flatbuffers.UOffsetT) { builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(Class), 0) }
func NodeAddID(builder *flatbuffers.Builder, ID int32) { builder.PrependInt32Slot(1, ID, 0) }
func NodeAddName(builder *flatbuffers.Builder, Name flatbuffers.UOffsetT) { builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(Name), 0) }
func NodeAddCSS(builder *flatbuffers.Builder, CSS flatbuffers.UOffsetT) { builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(CSS), 0) }
func NodeAddX(builder *flatbuffers.Builder, X int32) { builder.PrependInt32Slot(4, X, 0) }
func NodeAddY(builder *flatbuffers.Builder, Y int32) { builder.PrependInt32Slot(5, Y, 0) }
func NodeAddRectangleX(builder *flatbuffers.Builder, RectangleX int32) { builder.PrependInt32Slot(6, RectangleX, 0) }
func NodeAddRectangleY(builder *flatbuffers.Builder, RectangleY int32) { builder.PrependInt32Slot(7, RectangleY, 0) }
func NodeAddIsVisible(builder *flatbuffers.Builder, IsVisible byte) { builder.PrependByteSlot(8, IsVisible, 0) }
func NodeAddLabelText(builder *flatbuffers.Builder, LabelText flatbuffers.UOffsetT) { builder.PrependUOffsetTSlot(9, flatbuffers.UOffsetT(LabelText), 0) }
func NodeAddChildren(builder *flatbuffers.Builder, Children flatbuffers.UOffsetT) { builder.PrependUOffsetTSlot(10, flatbuffers.UOffsetT(Children), 0) }
func NodeStartChildrenVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT { return builder.StartVector(4, numElems, 4)
}
func NodeEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }

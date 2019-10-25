// automatically generated by the FlatBuffers compiler, do not modify

#ifndef FLATBUFFERS_GENERATED_SCENE_INNIUM_NETWORKING_H_
#define FLATBUFFERS_GENERATED_SCENE_INNIUM_NETWORKING_H_

#include "flatbuffers/flatbuffers.h"


namespace innium {
namespace networking {

struct Node;

struct Node : private flatbuffers::Table {
  const flatbuffers::String *Class() const { return GetPointer<const flatbuffers::String *>(4); }
  int32_t ID() const { return GetField<int32_t>(6, 0); }
  const flatbuffers::String *Name() const { return GetPointer<const flatbuffers::String *>(8); }
  const flatbuffers::String *CSS() const { return GetPointer<const flatbuffers::String *>(10); }
  int32_t X() const { return GetField<int32_t>(12, 0); }
  int32_t Y() const { return GetField<int32_t>(14, 0); }
  int32_t RectangleX() const { return GetField<int32_t>(16, 0); }
  int32_t RectangleY() const { return GetField<int32_t>(18, 0); }
  uint8_t IsVisible() const { return GetField<uint8_t>(20, 0); }
  const flatbuffers::String *LabelText() const { return GetPointer<const flatbuffers::String *>(22); }
  const flatbuffers::Vector<flatbuffers::Offset<Node>> *Children() const { return GetPointer<const flatbuffers::Vector<flatbuffers::Offset<Node>> *>(24); }
  bool Verify(flatbuffers::Verifier &verifier) const {
    return VerifyTableStart(verifier) &&
           VerifyField<flatbuffers::uoffset_t>(verifier, 4 /* Class */) &&
           verifier.Verify(Class()) &&
           VerifyField<int32_t>(verifier, 6 /* ID */) &&
           VerifyField<flatbuffers::uoffset_t>(verifier, 8 /* Name */) &&
           verifier.Verify(Name()) &&
           VerifyField<flatbuffers::uoffset_t>(verifier, 10 /* CSS */) &&
           verifier.Verify(CSS()) &&
           VerifyField<int32_t>(verifier, 12 /* X */) &&
           VerifyField<int32_t>(verifier, 14 /* Y */) &&
           VerifyField<int32_t>(verifier, 16 /* RectangleX */) &&
           VerifyField<int32_t>(verifier, 18 /* RectangleY */) &&
           VerifyField<uint8_t>(verifier, 20 /* IsVisible */) &&
           VerifyField<flatbuffers::uoffset_t>(verifier, 22 /* LabelText */) &&
           verifier.Verify(LabelText()) &&
           VerifyField<flatbuffers::uoffset_t>(verifier, 24 /* Children */) &&
           verifier.Verify(Children()) &&
           verifier.VerifyVectorOfTables(Children()) &&
           verifier.EndTable();
  }
};

struct NodeBuilder {
  flatbuffers::FlatBufferBuilder &fbb_;
  flatbuffers::uoffset_t start_;
  void add_Class(flatbuffers::Offset<flatbuffers::String> Class) { fbb_.AddOffset(4, Class); }
  void add_ID(int32_t ID) { fbb_.AddElement<int32_t>(6, ID, 0); }
  void add_Name(flatbuffers::Offset<flatbuffers::String> Name) { fbb_.AddOffset(8, Name); }
  void add_CSS(flatbuffers::Offset<flatbuffers::String> CSS) { fbb_.AddOffset(10, CSS); }
  void add_X(int32_t X) { fbb_.AddElement<int32_t>(12, X, 0); }
  void add_Y(int32_t Y) { fbb_.AddElement<int32_t>(14, Y, 0); }
  void add_RectangleX(int32_t RectangleX) { fbb_.AddElement<int32_t>(16, RectangleX, 0); }
  void add_RectangleY(int32_t RectangleY) { fbb_.AddElement<int32_t>(18, RectangleY, 0); }
  void add_IsVisible(uint8_t IsVisible) { fbb_.AddElement<uint8_t>(20, IsVisible, 0); }
  void add_LabelText(flatbuffers::Offset<flatbuffers::String> LabelText) { fbb_.AddOffset(22, LabelText); }
  void add_Children(flatbuffers::Offset<flatbuffers::Vector<flatbuffers::Offset<Node>>> Children) { fbb_.AddOffset(24, Children); }
  NodeBuilder(flatbuffers::FlatBufferBuilder &_fbb) : fbb_(_fbb) { start_ = fbb_.StartTable(); }
  NodeBuilder &operator=(const NodeBuilder &);
  flatbuffers::Offset<Node> Finish() {
    auto o = flatbuffers::Offset<Node>(fbb_.EndTable(start_, 11));
    return o;
  }
};

inline flatbuffers::Offset<Node> CreateNode(flatbuffers::FlatBufferBuilder &_fbb,
   flatbuffers::Offset<flatbuffers::String> Class = 0,
   int32_t ID = 0,
   flatbuffers::Offset<flatbuffers::String> Name = 0,
   flatbuffers::Offset<flatbuffers::String> CSS = 0,
   int32_t X = 0,
   int32_t Y = 0,
   int32_t RectangleX = 0,
   int32_t RectangleY = 0,
   uint8_t IsVisible = 0,
   flatbuffers::Offset<flatbuffers::String> LabelText = 0,
   flatbuffers::Offset<flatbuffers::Vector<flatbuffers::Offset<Node>>> Children = 0) {
  NodeBuilder builder_(_fbb);
  builder_.add_Children(Children);
  builder_.add_LabelText(LabelText);
  builder_.add_RectangleY(RectangleY);
  builder_.add_RectangleX(RectangleX);
  builder_.add_Y(Y);
  builder_.add_X(X);
  builder_.add_CSS(CSS);
  builder_.add_Name(Name);
  builder_.add_ID(ID);
  builder_.add_Class(Class);
  builder_.add_IsVisible(IsVisible);
  return builder_.Finish();
}

inline const Node *GetNode(const void *buf) { return flatbuffers::GetRoot<Node>(buf); }

inline bool VerifyNodeBuffer(flatbuffers::Verifier &verifier) { return verifier.VerifyBuffer<Node>(); }

inline void FinishNodeBuffer(flatbuffers::FlatBufferBuilder &fbb, flatbuffers::Offset<Node> root) { fbb.Finish(root); }

}  // namespace networking
}  // namespace innium

#endif  // FLATBUFFERS_GENERATED_SCENE_INNIUM_NETWORKING_H_

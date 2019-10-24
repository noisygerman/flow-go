// Code generated by capnpc-go. DO NOT EDIT.

package captain

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type GuaranteedCollection struct{ capnp.Struct }

// GuaranteedCollection_TypeID is the unique identifier for the type GuaranteedCollection.
const GuaranteedCollection_TypeID = 0xc45eaaad82049c1b

func NewGuaranteedCollection(s *capnp.Segment) (GuaranteedCollection, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return GuaranteedCollection{st}, err
}

func NewRootGuaranteedCollection(s *capnp.Segment) (GuaranteedCollection, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return GuaranteedCollection{st}, err
}

func ReadRootGuaranteedCollection(msg *capnp.Message) (GuaranteedCollection, error) {
	root, err := msg.RootPtr()
	return GuaranteedCollection{root.Struct()}, err
}

func (s GuaranteedCollection) String() string {
	str, _ := text.Marshal(0xc45eaaad82049c1b, s.Struct)
	return str
}

func (s GuaranteedCollection) Hash() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s GuaranteedCollection) HasHash() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s GuaranteedCollection) SetHash(v []byte) error {
	return s.Struct.SetData(0, v)
}

func (s GuaranteedCollection) Signature() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return []byte(p.Data()), err
}

func (s GuaranteedCollection) HasSignature() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s GuaranteedCollection) SetSignature(v []byte) error {
	return s.Struct.SetData(1, v)
}

// GuaranteedCollection_List is a list of GuaranteedCollection.
type GuaranteedCollection_List struct{ capnp.List }

// NewGuaranteedCollection creates a new list of GuaranteedCollection.
func NewGuaranteedCollection_List(s *capnp.Segment, sz int32) (GuaranteedCollection_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return GuaranteedCollection_List{l}, err
}

func (s GuaranteedCollection_List) At(i int) GuaranteedCollection {
	return GuaranteedCollection{s.List.Struct(i)}
}

func (s GuaranteedCollection_List) Set(i int, v GuaranteedCollection) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s GuaranteedCollection_List) String() string {
	str, _ := text.MarshalList(0xc45eaaad82049c1b, s.List)
	return str
}

// GuaranteedCollection_Promise is a wrapper for a GuaranteedCollection promised by a client call.
type GuaranteedCollection_Promise struct{ *capnp.Pipeline }

func (p GuaranteedCollection_Promise) Struct() (GuaranteedCollection, error) {
	s, err := p.Pipeline.Struct()
	return GuaranteedCollection{s}, err
}

const schema_ca5023959d73b9b5 = "x\xda\x12\xf8\xed\xc0d\xc8\xda\xcf\xc6\xc0\x10\xe8\xc3\xca" +
	"\xf6_z\x0eK\xd3\xdaUqG\x18\x04\x05\x19\xffo" +
	"\xddY<w\xaar\xc0)\x06V&v\x06\x06c]" +
	"f#Fa[fv\x06\x06aK\xe6r\x86\xe3\xff" +
	"\x93\xf3srR\x93K2\x99\xf3\xf3\xf4\x92\x13\x0b\xf2" +
	"\x0a\xac\xdcK\x13\x8b\x12\xf3JRSS\x9cAr\xf2" +
	"\xc9%\x99\xf9y\x01\x8c\x8c\x81\x1c\xcc,\x0c\x0c,\x8c" +
	"\x0c\x0c\x82\x9aZ\x0c\x0c\x81*\xcc\x8c\x81\x06L\x8c\x82" +
	"\x8c\x8c\"\x8c A\xdd \x06\x86@\x1df\xc6@\x0b" +
	"&F\xfe\x8c\xc4\xe2\x0cF^\x06&F^\x06\xc6\xff" +
	"\xc5\x99\xe9y\x89%\xa5E\x0c\x8c\xa901@\x00\x00" +
	"\x00\xff\xff\xdd\xbe)\xe7"

func init() {
	schemas.Register(schema_ca5023959d73b9b5,
		0xc45eaaad82049c1b)
}

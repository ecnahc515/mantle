// Code generated by protoc-gen-gogo.
// source: log_entry.proto
// DO NOT EDIT!

/*
	Package protobuf is a generated protocol buffer package.

	It is generated from these files:
		log_entry.proto

	It has these top-level messages:
		LogEntry
*/
package protobuf

import proto "github.com/coreos/mantle/Godeps/_workspace/src/github.com/gogo/protobuf/proto"
import math "math"

// discarding unused import gogoproto "github.com/gogo/protobuf/gogoproto/gogo.pb"

import io "io"
import fmt "fmt"
import github_com_gogo_protobuf_proto "github.com/coreos/mantle/Godeps/_workspace/src/github.com/gogo/protobuf/proto"

import strings "strings"
import reflect "reflect"

import sort "sort"
import strconv "strconv"

import bytes "bytes"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type LogEntry struct {
	Index            *uint64 `protobuf:"varint,1,req" json:"Index,omitempty"`
	Term             *uint64 `protobuf:"varint,2,req" json:"Term,omitempty"`
	CommandName      *string `protobuf:"bytes,3,req" json:"CommandName,omitempty"`
	Command          []byte  `protobuf:"bytes,4,opt" json:"Command,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *LogEntry) Reset()      { *m = LogEntry{} }
func (*LogEntry) ProtoMessage() {}

func (m *LogEntry) GetIndex() uint64 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return 0
}

func (m *LogEntry) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *LogEntry) GetCommandName() string {
	if m != nil && m.CommandName != nil {
		return *m.CommandName
	}
	return ""
}

func (m *LogEntry) GetCommand() []byte {
	if m != nil {
		return m.Command
	}
	return nil
}

func init() {
}
func (m *LogEntry) Unmarshal(data []byte) error {
	var hasFields [1]uint64
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				v |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Index = &v
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				v |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Term = &v
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommandName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + int(stringLen)
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(data[iNdEx:postIndex])
			m.CommandName = &s
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000004)
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Command", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Command = append([]byte{}, data[iNdEx:postIndex]...)
			iNdEx = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			iNdEx -= sizeOfWire
			skippy, err := skipLogEntry(data[iNdEx:])
			if err != nil {
				return err
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Index")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("Term")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("CommandName")
	}

	return nil
}
func skipLogEntry(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for {
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipLogEntry(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}
func (this *LogEntry) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&LogEntry{`,
		`Index:` + valueToStringLogEntry(this.Index) + `,`,
		`Term:` + valueToStringLogEntry(this.Term) + `,`,
		`CommandName:` + valueToStringLogEntry(this.CommandName) + `,`,
		`Command:` + valueToStringLogEntry(this.Command) + `,`,
		`XXX_unrecognized:` + fmt.Sprintf("%v", this.XXX_unrecognized) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringLogEntry(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *LogEntry) Size() (n int) {
	var l int
	_ = l
	if m.Index != nil {
		n += 1 + sovLogEntry(uint64(*m.Index))
	}
	if m.Term != nil {
		n += 1 + sovLogEntry(uint64(*m.Term))
	}
	if m.CommandName != nil {
		l = len(*m.CommandName)
		n += 1 + l + sovLogEntry(uint64(l))
	}
	if m.Command != nil {
		l = len(m.Command)
		n += 1 + l + sovLogEntry(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovLogEntry(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozLogEntry(x uint64) (n int) {
	return sovLogEntry(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func NewPopulatedLogEntry(r randyLogEntry, easy bool) *LogEntry {
	this := &LogEntry{}
	v1 := uint64(uint64(r.Uint32()))
	this.Index = &v1
	v2 := uint64(uint64(r.Uint32()))
	this.Term = &v2
	v3 := randStringLogEntry(r)
	this.CommandName = &v3
	if r.Intn(10) != 0 {
		v4 := r.Intn(100)
		this.Command = make([]byte, v4)
		for i := 0; i < v4; i++ {
			this.Command[i] = byte(r.Intn(256))
		}
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedLogEntry(r, 5)
	}
	return this
}

type randyLogEntry interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneLogEntry(r randyLogEntry) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringLogEntry(r randyLogEntry) string {
	v5 := r.Intn(100)
	tmps := make([]rune, v5)
	for i := 0; i < v5; i++ {
		tmps[i] = randUTF8RuneLogEntry(r)
	}
	return string(tmps)
}
func randUnrecognizedLogEntry(r randyLogEntry, maxFieldNumber int) (data []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		data = randFieldLogEntry(data, r, fieldNumber, wire)
	}
	return data
}
func randFieldLogEntry(data []byte, r randyLogEntry, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		data = encodeVarintPopulateLogEntry(data, uint64(key))
		v6 := r.Int63()
		if r.Intn(2) == 0 {
			v6 *= -1
		}
		data = encodeVarintPopulateLogEntry(data, uint64(v6))
	case 1:
		data = encodeVarintPopulateLogEntry(data, uint64(key))
		data = append(data, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		data = encodeVarintPopulateLogEntry(data, uint64(key))
		ll := r.Intn(100)
		data = encodeVarintPopulateLogEntry(data, uint64(ll))
		for j := 0; j < ll; j++ {
			data = append(data, byte(r.Intn(256)))
		}
	default:
		data = encodeVarintPopulateLogEntry(data, uint64(key))
		data = append(data, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return data
}
func encodeVarintPopulateLogEntry(data []byte, v uint64) []byte {
	for v >= 1<<7 {
		data = append(data, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	data = append(data, uint8(v))
	return data
}
func (m *LogEntry) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *LogEntry) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Index == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("Index")
	} else {
		data[i] = 0x8
		i++
		i = encodeVarintLogEntry(data, i, uint64(*m.Index))
	}
	if m.Term == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("Term")
	} else {
		data[i] = 0x10
		i++
		i = encodeVarintLogEntry(data, i, uint64(*m.Term))
	}
	if m.CommandName == nil {
		return 0, github_com_gogo_protobuf_proto.NewRequiredNotSetError("CommandName")
	} else {
		data[i] = 0x1a
		i++
		i = encodeVarintLogEntry(data, i, uint64(len(*m.CommandName)))
		i += copy(data[i:], *m.CommandName)
	}
	if m.Command != nil {
		data[i] = 0x22
		i++
		i = encodeVarintLogEntry(data, i, uint64(len(m.Command)))
		i += copy(data[i:], m.Command)
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeFixed64LogEntry(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32LogEntry(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintLogEntry(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (this *LogEntry) GoString() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&protobuf.LogEntry{` +
		`Index:` + valueToGoStringLogEntry(this.Index, "uint64"),
		`Term:` + valueToGoStringLogEntry(this.Term, "uint64"),
		`CommandName:` + valueToGoStringLogEntry(this.CommandName, "string"),
		`Command:` + valueToGoStringLogEntry(this.Command, "byte"),
		`XXX_unrecognized:` + fmt.Sprintf("%#v", this.XXX_unrecognized) + `}`}, ", ")
	return s
}
func valueToGoStringLogEntry(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringLogEntry(e map[int32]github_com_gogo_protobuf_proto.Extension) string {
	if e == nil {
		return "nil"
	}
	s := "map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings.Join(ss, ",") + "}"
	return s
}
func (this *LogEntry) VerboseEqual(that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*LogEntry)
	if !ok {
		return fmt.Errorf("that is not of type *LogEntry")
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *LogEntry but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *LogEntrybut is not nil && this == nil")
	}
	if this.Index != nil && that1.Index != nil {
		if *this.Index != *that1.Index {
			return fmt.Errorf("Index this(%v) Not Equal that(%v)", *this.Index, *that1.Index)
		}
	} else if this.Index != nil {
		return fmt.Errorf("this.Index == nil && that.Index != nil")
	} else if that1.Index != nil {
		return fmt.Errorf("Index this(%v) Not Equal that(%v)", this.Index, that1.Index)
	}
	if this.Term != nil && that1.Term != nil {
		if *this.Term != *that1.Term {
			return fmt.Errorf("Term this(%v) Not Equal that(%v)", *this.Term, *that1.Term)
		}
	} else if this.Term != nil {
		return fmt.Errorf("this.Term == nil && that.Term != nil")
	} else if that1.Term != nil {
		return fmt.Errorf("Term this(%v) Not Equal that(%v)", this.Term, that1.Term)
	}
	if this.CommandName != nil && that1.CommandName != nil {
		if *this.CommandName != *that1.CommandName {
			return fmt.Errorf("CommandName this(%v) Not Equal that(%v)", *this.CommandName, *that1.CommandName)
		}
	} else if this.CommandName != nil {
		return fmt.Errorf("this.CommandName == nil && that.CommandName != nil")
	} else if that1.CommandName != nil {
		return fmt.Errorf("CommandName this(%v) Not Equal that(%v)", this.CommandName, that1.CommandName)
	}
	if !bytes.Equal(this.Command, that1.Command) {
		return fmt.Errorf("Command this(%v) Not Equal that(%v)", this.Command, that1.Command)
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}
func (this *LogEntry) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*LogEntry)
	if !ok {
		return false
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Index != nil && that1.Index != nil {
		if *this.Index != *that1.Index {
			return false
		}
	} else if this.Index != nil {
		return false
	} else if that1.Index != nil {
		return false
	}
	if this.Term != nil && that1.Term != nil {
		if *this.Term != *that1.Term {
			return false
		}
	} else if this.Term != nil {
		return false
	} else if that1.Term != nil {
		return false
	}
	if this.CommandName != nil && that1.CommandName != nil {
		if *this.CommandName != *that1.CommandName {
			return false
		}
	} else if this.CommandName != nil {
		return false
	} else if that1.CommandName != nil {
		return false
	}
	if !bytes.Equal(this.Command, that1.Command) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

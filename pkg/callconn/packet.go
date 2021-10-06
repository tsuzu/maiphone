package callconn

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// {
	uuidLen = 36
	// }
)

var (
	ConnectPayload   = []byte{0x10, 0x10}
	ConnectedPayload = []byte{0x20, 0x10}
	PingPayload      = []byte{0x90, 0x20}
	PongPayload      = []byte{0x91, 0x20}
)

type Direction int

const (
	ToInterphone   Direction = 0x01
	FromInterphone Direction = 0x02
)

type Packet struct {
	Direction Direction
	UUID      string
	Flags     []byte
	ID        int

	Trailing bytes.Buffer
}

func NewPacket(dir Direction) *Packet {
	return &Packet{
		Direction: dir,
		UUID:      "",
		Flags:     make([]byte, 8),
	}
}

func NewPacketToInterphone() *Packet {
	return NewPacket(ToInterphone)
}

func ParsePacket(b []byte) (*Packet, error) {
	if len(b) < 56 {
		return nil, fmt.Errorf("too short")
	}
	pkt := &Packet{}

	if b[0] != 0x40 {
		return nil, fmt.Errorf("invalid magic")
	}
	b = b[1:]

	pkt.Direction = Direction(b[0])
	b = b[1:]

	if b[0] != '{' {
		return nil, fmt.Errorf("invalid '{'")
	}
	b = b[1:]

	pkt.UUID = string(b[:uuidLen])
	b = b[uuidLen:]

	if b[0] != '}' {
		return nil, fmt.Errorf("invalid '}'")
	}

	b = b[1:]

	pkt.Flags = b[:8]
	b = b[8:]

	pkt.ID = int(binary.LittleEndian.Uint32(b[:4]))
	b = b[4:]

	if int(binary.LittleEndian.Uint32(b[:4])) != len(b) {
		return nil, fmt.Errorf("invalid length")
	}
	b = b[4:]

	pkt.Trailing.Write(b)

	return pkt, nil
}

func (p *Packet) Compose() []byte {
	if len([]byte(p.UUID)) != uuidLen {
		panic("Invalid UUID length")
	}

	buf := bytes.Buffer{}

	buf.WriteByte(0x40)
	buf.WriteByte(byte(p.Direction))
	buf.WriteString("{" + p.UUID + "}")
	buf.Write(p.Flags)
	binary.Write(&buf, binary.LittleEndian, int32(p.ID))
	binary.Write(&buf, binary.LittleEndian, int32(buf.Len()+4+p.Trailing.Len()))
	buf.Write(p.Trailing.Bytes())

	return buf.Bytes()
}

func (p *Packet) SetPayloadType(pt []byte) {
	p.Flags[4] = pt[0]
	p.Flags[5] = pt[1]
}

func (p *Packet) GetPayloadType() []byte {
	return []byte{p.Flags[4], p.Flags[5]}
}

func (p *Packet) Write(b []byte) {
	p.Trailing.Write(b)
}

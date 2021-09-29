package interphoneconn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var (
	protocolHeader = []byte{0x40, 0x00, 0x00, 0x00}
)

const (
	protocolHeaderLen = 4
	sequentialIDLen   = 4
	allLen            = 4
	trailingLen       = 4
	// {
	uuidLen = 36
	// }
	flagLen = 14
)

type Packet struct {
	ID    int
	UUID  string
	Flags []byte

	Trailing bytes.Buffer
}

func NewPacket() *Packet {
	p := Packet{}

	p.Flags = make([]byte, 14)

	return &p
}

func ParsePacket(r io.Reader) (*Packet, error) {
	p := Packet{}

	header, err := io.ReadAll(io.LimitReader(r, 4))

	if err != nil {
		return nil, fmt.Errorf("failed to read protocol header: %w", err)
	}

	if !bytes.Equal(header, protocolHeader) {
		return nil, fmt.Errorf("invalid protocol header")
	}

	var id int32
	err = binary.Read(r, binary.LittleEndian, &id)

	if err != nil {
		return nil, fmt.Errorf("failed to read packet id: %w", err)
	}

	p.ID = int(id)

	var length int32
	err = binary.Read(r, binary.LittleEndian, &length)

	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}

	data, err := io.ReadAll(io.LimitReader(r, int64(length-protocolHeaderLen-sequentialIDLen-allLen)))

	if err != nil {
		return nil, fmt.Errorf("failed to read packet data: %w", err)
	}
	data = data[trailingLen:]

	p.UUID = string(data[1 : 1+uuidLen])
	data = data[2+uuidLen:]

	p.Flags = data[:flagLen]
	data = data[flagLen:]

	p.Trailing.Write(data)

	return &p, nil
}

func (p *Packet) Write(b []byte) {
	p.Trailing.Write(b)
}

func (p *Packet) UpdateFlags(b []byte) {
	if len(b) != len(p.Flags) {
		panic("Invalid length")
	}

	copy(p.Flags, b)
}

func (p *Packet) Compose() []byte {
	if len([]byte(p.UUID)) != uuidLen {
		panic("Invalid UUID length")
	}

	data := bytes.Buffer{}
	data.Write(protocolHeader)
	binary.Write(&data, binary.LittleEndian, int32(p.ID))
	lengthIndex := data.Len()
	data.Write([]byte{0, 0, 0, 0}) // Fake length
	binary.Write(&data, binary.LittleEndian, int32(p.Trailing.Len()))

	data.WriteString("{" + p.UUID + "}")
	data.Write(p.Flags)
	data.Write(p.Trailing.Bytes())

	binary.LittleEndian.PutUint32(data.Bytes()[lengthIndex:lengthIndex+4], uint32(data.Len()))

	return data.Bytes()
}

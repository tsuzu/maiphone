package callconn

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Conn struct {
	seqID int32
	uuid  string

	conn net.Conn

	Logger *log.Logger
}

func NewConn(address, uuid string) (*Conn, error) {
	conn, err := net.Dial("udp", address)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to the interphone: %w", err)
	}

	c := &Conn{
		seqID:  0,
		conn:   conn,
		uuid:   uuid,
		Logger: log.Default(),
	}

	return c, nil

}

func (i *Conn) Close() error {
	return i.conn.Close()
}

func (i *Conn) Send(pkt *Packet) (int, error) {
	seq := pkt.ID
	if seq == 0 {
		seq = int(atomic.AddInt32(&i.seqID, 1))
		pkt.ID = seq
	}

	pkt.UUID = i.uuid

	_, err := i.conn.Write(pkt.Compose())

	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	return seq, nil
}

func (i *Conn) recv() (*Packet, error) {
	raw := make([]byte, 8192)

	n, err := i.conn.Read(raw)

	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	pkt, err := ParsePacket(raw[:n])

	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return pkt, nil
}

func (i *Conn) Recv() (*Packet, error) {
	for {
		pkt, err := i.recv()

		if err != nil {
			return nil, err
		}

		if bytes.Equal(pkt.GetPayloadType(), PongPayload) {
			continue
		}

		return pkt, nil
	}
}

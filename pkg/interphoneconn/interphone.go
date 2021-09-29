package interphoneconn

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
)

type Conn struct {
	seqID int32
	uuid  string

	conn   net.Conn
	closed chan<- struct{}
}

func NewConn(address, uuid string) (*Conn, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to the interphone: %w", err)
	}

	closed := make(chan struct{})
	go func() {
		for {

			conn.Read(nil)
		}
		<-closed
	}()

	return &Conn{
		conn:   conn,
		uuid:   uuid,
		closed: closed,
	}, nil
}

func (i *Conn) Send(pkt *Packet) error {
	seq := atomic.AddInt32(&i.seqID, 1)
	pkt.ID = int(seq)
	pkt.UUID = i.uuid

	_, err := i.conn.Write(pkt.Compose())

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (i *Conn) SendRecv(ctx context.Context, pkt *Packet, handler func(*Packet) error) error {
	return nil
}

func (i *Conn) startUp() error {
	return nil
}

package mngconn

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type handlerType = func(pkt *Packet)

type Conn struct {
	seqID int32
	uuid  string

	conn   net.Conn
	closed chan struct{}
	recved chan *Packet

	Logger *log.Logger
}

func NewConn(address, uuid string) (*Conn, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to the interphone: %w", err)
	}

	c := &Conn{
		seqID:  0,
		conn:   conn,
		uuid:   uuid,
		closed: make(chan struct{}, 1),
		recved: make(chan *Packet, 1),
		Logger: log.Default(),
	}

	go c.recvWorker()

	return c, nil
}

func (i *Conn) recvWorker() {
	reader := bufio.NewReader(i.conn)
	defer i.conn.Close()

	checkClosed := func() bool {
		select {
		case <-i.closed:
			return true
		default:
			return false
		}
	}

	for {
		if checkClosed() {
			return
		}

		pkt, err := ParsePacket(reader)

		if checkClosed() {
			return
		}

		if err != nil {
			i.Logger.Printf("failed to read packet: %+v", err)

			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			i.Close()

			return
		}

		i.defaultHandler(pkt)
	}
}

func (i *Conn) defaultHandler(pkt *Packet) {
	if pkt.GetPayloadType() == PongPayload {
		return
	}

	if _, err := i.Send(NewPong(pkt.ID)); err != nil {
		i.Logger.Printf("failed to send pong: %+v", err)
	}

	atomic.StoreInt32(&i.seqID, int32(pkt.ID))

	i.recved <- pkt
}

func (i *Conn) Close() error {
	defer func() {
		recover()
	}()

	close(i.closed)

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

func (i *Conn) Recv() (*Packet, error) {
	select {
	case pkt := <-i.recved:
		return pkt, nil
	case <-i.closed:
		return nil, net.ErrClosed
	}
}

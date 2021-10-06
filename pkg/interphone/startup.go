package interphone

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/tsuzu/maiphone/pkg/callconn"
	"github.com/tsuzu/maiphone/pkg/mngconn"
)

func (i *Interphone) mngConnWaitFor(filter func(pkt *mngconn.Packet) bool) error {
	for {
		pkt, err := i.mngconn.Recv()

		if err != nil {
			return fmt.Errorf("failed to receive packet: %v", err)
		}

		if filter(pkt) {
			return nil
		}
	}
}

func (i *Interphone) callConnwaitFor(filter func(pkt *callconn.Packet) bool) error {
	for {
		pkt, err := i.callconn.Recv()

		if err != nil {
			return fmt.Errorf("failed to receive packet: %v", err)
		}

		if filter(pkt) {
			return nil
		}
	}
}

func (i *Interphone) initialize() error {
	{
		pkt := mngconn.NewPacket()
		pkt.SetPayloadType(mngconn.ConnectPayload)

		_, err := i.mngconn.Send(pkt)

		if err != nil {
			return fmt.Errorf("failed to send connect message: %w", err)
		}
	}

	i.mngConnWaitFor(func(pkt *mngconn.Packet) bool {
		return pkt.GetPayloadType() == mngconn.ConnectedPayload
	})

	{
		pkt := callconn.NewPacketToInterphone()
		pkt.SetPayloadType(callconn.ConnectPayload)

		i.callconn.Send(pkt)
	}

	i.callConnwaitFor(func(pkt *callconn.Packet) bool {
		return bytes.Equal(pkt.GetPayloadType(), callconn.ConnectedPayload)
	})

	{
		pkt := mngconn.NewPacket()
		pkt.SetPayloadType(mngconn.InitSend)
		pkt.Flags[4] = 0x22

		_, err := i.mngconn.Send(pkt)

		if err != nil {
			return fmt.Errorf("failed to send initialization message: %w", err)
		}
	}

	i.mngConnWaitFor(func(pkt *mngconn.Packet) bool {
		return pkt.GetPayloadType() == mngconn.InitRecv
	})

	{
		pkt := mngconn.NewPacket()
		pkt.SetPayloadType(mngconn.InitSend)
		copy(pkt.Flags[3:], []byte{0x04, 0x01, 0x80})

		_, err := i.mngconn.Send(pkt)

		if err != nil {
			return fmt.Errorf("failed to send initialization message: %w", err)
		}
	}

	i.mngConnWaitFor(func(pkt *mngconn.Packet) bool {
		return pkt.GetPayloadType() == mngconn.StatusNotification
	})

	{
		pkt := mngconn.NewPacket()
		pkt.SetPayloadType(mngconn.InitSend)
		pkt.Flags[4] = 0x21

		raw := `9fef6607d208caa65b2e0f2d4ce7301652dab28def08be406602945637b69afee70dda21aa9f6e314fb4c84dd0bc`
		decoded, _ := hex.DecodeString(raw)
		pkt.Trailing.Write(decoded)

		_, err := i.mngconn.Send(pkt)

		if err != nil {
			return fmt.Errorf("failed to send initialization message: %w", err)
		}
	}

	i.mngConnWaitFor(func(pkt *mngconn.Packet) bool {
		return pkt.GetPayloadType() == mngconn.InitRecv
	})

	return nil
}

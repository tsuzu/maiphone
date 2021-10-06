package interphone

import (
	"fmt"
	"log"
	"time"

	"github.com/tsuzu/maiphone/pkg/callconn"
	"github.com/tsuzu/maiphone/pkg/mngconn"
)

type Interphone struct {
	mngconn  *mngconn.Conn
	callconn *callconn.Conn

	Logger *log.Logger
}

func NewInterphone(mngaddr, calladdr, uuid string) (*Interphone, error) {
	i := &Interphone{
		Logger: log.Default(),
	}

	mc, err := mngconn.NewConn(mngaddr, uuid)

	if err != nil {
		return nil, fmt.Errorf("failed to create mngconn: %v", err)
	}

	i.mngconn = mc

	cc, err := callconn.NewConn(calladdr, uuid)

	if err != nil {
		return nil, fmt.Errorf("failed to create callconn: %v", err)
	}

	i.callconn = cc

	if err := i.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize connections: %w", err)
	}

	go i.ping()

	return i, nil
}

func (i *Interphone) ping() {
	ticker := time.NewTicker(10 * time.Second)

	for {
		<-ticker.C

		{
			pkt := mngconn.NewPacket()
			pkt.SetPayloadType(mngconn.PingPayload)

			if _, err := i.mngconn.Send(pkt); err != nil {
				i.Logger.Printf("failed to ping via mngconn: %+v", err)

				return
			}
		}

		{
			pkt := callconn.NewPacketToInterphone()
			pkt.SetPayloadType(callconn.PingPayload)

			if _, err := i.callconn.Send(pkt); err != nil {
				i.Logger.Printf("failed to ping via callconn: %+v", err)

				return
			}
		}
	}
}

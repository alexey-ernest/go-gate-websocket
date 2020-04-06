package gatewebsocket

import (
	"github.com/json-iterator/go"
	"github.com/alexey-ernest/go-gate-websocket/ws"
	"log"
	"fmt"
	"time"
	"math/rand"
	"sync"
)

const (
	depthLimit    = 30
	depthInterval = "0.00000001"
	depthSubscribeMethod   = "depth.subscribe"
	depthUpdateMethod = "depth.update"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type gateWs struct {
	baseURL string
	Conn *ws.WsConn
	pool *sync.Pool // pool for updateMessage structs to re-use old objects
}

func NewGateWs() *gateWs {
	gWs := &gateWs{
		baseURL: "wss://ws.gate.io/v3",
		pool: &sync.Pool {
			New: func()interface{} {
				return &UpdateMessage{}
			},
		},
	}
	
	return gWs
}

func (this *gateWs) subscribe(endpoint string, string pair, handle func(msg []byte) error) error {
	wsBuilder := ws.NewWsBuilder().
		WsUrl(endpoint).
		AutoReconnect().
		MessageHandleFunc(handle)
	this.Conn = wsBuilder.Build()

	sub := SubscribeMessage {
		id: rand.Int(),
		method: depthSubscribeMethod,
		params: []string{pair, depthLimit, depthInterval},
	}
	msg, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	this.Conn.Subscribe(msg)
	return nil
}

func (this *gateWs) SubscribeDepth(pair string, callback func (*Depth)) (error, chan<- struct{}) {
	endpoint := this.baseURL
	close := make(chan struct{})

	handle := func(msg []byte) error {
		updmsg := this.pool.Get().(*UpdateMessage)
		defer this.pool.Put(updmsg) // return to the pool once done

		if err := json.Unmarshal(msg, updmsg); err != nil {
			log.Printf("json unmarshal error: %s", string(msg))
			return err
		}

		if updmsg.Method != depthUpdateMethod {
			return nil
		}

		// parsing message to the depth format
		rawDepth := AcquireDepth()
		if err := json.Unmarshal(updmsg.Params[0], &rawDepth.Clean); err != nil {
			log.Printf("json unmarshal error: %s", string(msg))
			return err
		}
		if err := json.Unmarshal(updmsg.Params[1], &rawDepth); err != nil {
			log.Printf("json unmarshal error: %s", string(msg))
			return err
		}
		if err := json.Unmarshal(updmsg.Params[2], &rawDepth.Market); err != nil {
			log.Printf("json unmarshal error: %s", string(msg))
			return err
		}

		callback(rawDepth)
		return nil
	}
	this.subscribe(endpoint, pair, handle)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <- close:
				this.Conn.Close()
				return
			case t := <-ticker.C:
				this.Conn.SendPingMessage([]byte(t.String()))
			}
		}
	}()

	return nil, close
}

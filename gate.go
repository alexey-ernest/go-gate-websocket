package gatewebsocket

import (
	"github.com/json-iterator/go"
	"github.com/alexey-ernest/go-gate-websocket/ws"
	"log"
	"time"
	"math/rand"
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
}

func NewGateWs() *gateWs {
	gWs := &gateWs{
		baseURL: "wss://ws.gate.io/v3",
	}
	
	return gWs
}

func (this *gateWs) subscribe(endpoint string, pair string, handle func(msg []byte) error) error {
	wsBuilder := ws.NewWsBuilder().
		WsUrl(endpoint).
		AutoReconnect().
		MessageHandleFunc(handle)
	this.Conn = wsBuilder.Build()

	sub := &subscribeMessage{
		Id: rand.Int(),
		Method: depthSubscribeMethod,
		Params: []interface{}{pair, depthLimit, depthInterval},
	}
	this.Conn.Subscribe(sub)
	return nil
}

func (this *gateWs) SubscribeDepth(pair string, callback func (*Depth)) (error, chan<- struct{}) {
	endpoint := this.baseURL
	close := make(chan struct{})

	handle := func(msg []byte) error {
		method := string(msg[12:24]) // manually parsing byte array is much faster
		//method := jsoniter.Get(msg, "method").ToString()
		if method != depthUpdateMethod {
			return nil
		}

		// parsing message to the depth format
		rawDepth := AcquireDepth()

		rawDepth.Clean = string(msg[38:42]) == "true" // manually parsing byte array is much faster
		//rawDepth.Clean = jsoniter.Get(msg, "params", 0).ToBool()
		
		// manually cutting params related bytes, it is much faster than parsing the whole message
		paramsStarts := 44
		if !rawDepth.Clean {
			paramsStarts++
		}
		paramsbytes := msg[paramsStarts:len(msg)-14]
		for i := len(paramsbytes) - 1; i >= 0; i -= 1 {
			if paramsbytes[i] == '}' {
				paramsbytes = paramsbytes[:i+1]
				break
			}
		}
		//paramsbytes := []byte(jsoniter.Get(msg, "params", 1).ToString())

		// parsing params
		if err := json.Unmarshal(paramsbytes, &rawDepth); err != nil {
			log.Printf("json unmarshal error: %s, %s", err, string(msg))
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

package gatewebsocket

import (
	"testing"
	"time"
)

func TestGateDepthConnection(t *testing.T) {
	ws := NewGateWs()
	err, close := ws.SubscribeDepth("BTC_USDT", func (d *Depth) {})
	if err != nil {
		t.Fatalf("failed to connect to gate websocket")
	}
	close <- struct{}{}
}

func TestGateDepthCloseConnectionDirectly(t *testing.T) {
	ws := NewGateWs()
	err, _ := ws.SubscribeDepth("BTC_USDT", func (d *Depth) {})
	if err != nil {
		t.Fatalf("failed to connect to gate websocket")
	}
	
	ws.Conn.Close()
	time.Sleep(time.Duration(1) * time.Second)
}

func TestGateDepthMessage(t *testing.T) {
	ws := NewGateWs()
	messages := make(chan *Depth, 10)
	err, close := ws.SubscribeDepth("BTC_USDT", func (d *Depth) {
		messages <- d
	})

	if err != nil {
		t.Fatalf("failed to connect to binance @depth websocket")
	}

	msg := <- messages
	if msg.Market == nil {
		t.Errorf("Market should be defined")
	}
	if len(msg.Bids) == 0 && len(msg.Asks) == 0 {
		t.Errorf("Depth update should contain asks or bids")
	}

	close <- struct{}{}
}

func BenchmarkGateMessageHandling(b *testing.B) {
	ws := NewBinanceWs()
	//messages := make(chan *Depth, 10)
	err, close := ws.SubscribeDepth("BTC_USDT2", func (d *Depth) {
		d.DecrementReferenceCount()
	})
	if err != nil {
		b.Fatalf("failed to connect to gate websocket")
	}

	b.ResetTimer()
	msg := []byte("{\"method\": \"depth.update\", \"params\": [true, {\"asks\": [[\"8000.00\",\"9.6250\"]],\"bids\": [[\"8000.00\",\"9.6250\"]]}, \"EOS_USDT\"],\"id\": null}")
	for i := 0; i < b.N; i += 1 {
		ws.Conn.ReceiveMessage(msg)
	}

	close <- struct{}{}
}

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

	time.Sleep(time.Duration(1) * time.Second)
	close <- struct{}{}
}

func TestGateDepthCloseConnectionDirectly(t *testing.T) {
	ws := NewGateWs()
	err, _ := ws.SubscribeDepth("BTC_USDT", func (d *Depth) {})
	if err != nil {
		t.Fatalf("failed to connect to gate websocket")
	}
	
	time.Sleep(time.Duration(1) * time.Second)
	ws.Conn.Close()
}

func TestGateDepthMessage(t *testing.T) {
	ws := NewGateWs()
	messages := make(chan *Depth, 10)
	err, close := ws.SubscribeDepth("BTC_USDT", func (d *Depth) {
		messages <- d
	})

	if err != nil {
		t.Fatalf("failed to connect to gate websocket")
	}

	msg := <- messages
	if !msg.Clean {
		t.Errorf("First depth update message should be clean")
	}
	if len(msg.Bids) == 0 && len(msg.Asks) == 0 {
		t.Errorf("Depth update should contain asks or bids")
	}

	close <- struct{}{}
}

func BenchmarkGateMessageHandling(b *testing.B) {
	ws := NewGateWs()
	err, close := ws.SubscribeDepth("BTC_USDT2", func (d *Depth) {
		d.DecrementReferenceCount() // release the object to the object pool for re-use
	})
	if err != nil {
		b.Fatalf("failed to connect to gate websocket")
	}

	time.Sleep(time.Duration(1) * time.Second)
	close <- struct{}{}

	b.ResetTimer()
	msg := []byte("{\"method\": \"depth.update\", \"params\": [false, {\"asks\": [[\"7295.91\", \"0.144\"], [\"7296.35\", \"0\"]], \"bids\": [[\"7281.02\", \"0\"], [\"7275\", \"0.002\"]]}, \"BTC_USDT\"], \"id\": null}")
	for i := 0; i < b.N; i += 1 {
		ws.Conn.ReceiveMessage(msg)
	}	
}

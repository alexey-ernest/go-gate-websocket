[![Build Status](https://travis-ci.com/alexey-ernest/go-gate-websocket.svg?branch=master)](https://travis-ci.com/alexey-ernest/go-gate-websocket)

# go-gate-websocket
Gate.io websocket client with optimized latency by leveraging object pool and fast json deserializer

## Optimized latency
Leveraging fast json deserializer and object pool for good base performance (i5 6-core): ~700 ns/op or ~1.4M op/s
```
$ go test --bench=. --benchtime 30s --benchmem

BenchmarkGateMessageHandling-6   	50664753	       707 ns/op	      72 B/op	       9 allocs/op
```

## Example

```
import (
	. "github.com/alexey-ernest/go-gate-websocket"
	"log"
)

func main() {
	ws := NewGateWs()
	messages := make(chan *Depth, 10)
	err, _ := ws.SubscribeDepth("BTC_USDT", func (d *Depth) {
		messages <- d
	})

	if err != nil {
		log.Fatalf("failed to connect to gate websocket")
	}

	for m := range messages {
		log.Printf("%+v\n", m.RawDepth)
		m.DecrementReferenceCount() // return object back to the object pool
	}
}
```

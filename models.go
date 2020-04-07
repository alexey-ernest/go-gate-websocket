package gatewebsocket

import (
	"github.com/alexey-ernest/go-gate-websocket/pool"
	//"errors"
)

type UpdateMessage struct {
	Method string `json:"method"`
	Params []interface{} `json:"params"`
}

type subscribeMessage struct {
	Id int `json:"id"`
	Method string `json:"method"`
	Params []interface{} `json:"params"`
}

type RawDepth struct {
	Clean bool
	Bids [][2]string
	Asks [][2]string
}

// pooled depth message
type Depth struct {
	pool.ReferenceCounter `json:"-"`
	RawDepth
}

func (d *Depth) Reset() {
	d.Clean = false
	d.Bids = nil
	d.Asks = nil
}

// Used by reference countable pool
func ResetDepth(i interface{}) error {
	return nil
}

// depth pool
var depthPool = pool.NewReferenceCountedPool(
	func(counter pool.ReferenceCounter) pool.ReferenceCountable {
		d := new(Depth)
		d.ReferenceCounter = counter
		return d
	}, ResetDepth)

// Method to get new Depth
func AcquireDepth() *Depth {
	return depthPool.Get().(*Depth)
}

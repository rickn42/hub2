package hub2

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

type Message = interface{}

type Pipe = chan Message

type Filter = func(Message) (Message, bool)

type Done = chan struct{}

var DestroyedHub = errors.New("destroyed hub")

type Hub struct {
	broadcast Pipe
	destroyed Done

	inMu    sync.Mutex
	inPipes map[Pipe]struct{}

	outMu    sync.RWMutex
	outPipes map[Pipe]Pipe
}

func NewHub() (hub *Hub) {
	hub = &Hub{
		broadcast: make(Pipe),
		destroyed: make(Done),
		inPipes:   make(map[Pipe]struct{}),
		outPipes:  make(map[Pipe]Pipe),
	}

	// hub goroutine start.
	go func() {
		var v Message
		var buf Pipe

		// loop until broadcast channel closed.
		for v = range hub.broadcast {
			hub.outMu.RLock()
			for _, buf = range hub.outPipes {
				buf <- v
			}
			hub.outMu.RUnlock()
		}

		// close all out buffer pipes.
		hub.outMu.Lock()
		for _, buf = range hub.outPipes {
			close(buf)
		}
		hub.outMu.Unlock()
	}()

	return hub
}

func (hub *Hub) MakeInPipe(filters ...Filter) (in Pipe, err error) {

	in = make(Pipe)

	hub.inMu.Lock()
	select {
	case <-hub.destroyed:
		hub.inMu.Unlock()
		return nil, DestroyedHub
	default:
	}
	hub.inPipes[in] = struct{}{}
	hub.inMu.Unlock()

	connectPipe(connectConfig{
		In:             in,
		Out:            hub.broadcast,
		Filters:        filters,
		PropagateClose: false,
	})

	return in, nil
}

func (hub *Hub) DestroyInPipes(ps ...Pipe) {
	hub.inMu.Lock()
	for _, p := range ps {
		if _, ok := hub.inPipes[p]; ok {
			delete(hub.inPipes, p)
			close(p)
		}
	}
	hub.inMu.Unlock()
}

func (hub *Hub) MakeOutPipe(bufSize int, filters ...Filter) (out Pipe, err error) {

	bufSize = int(math.Max(float64(bufSize), 0))

	out = make(Pipe)
	buf := make(Pipe, bufSize)

	hub.outMu.Lock()
	select {
	case <-hub.destroyed:
		hub.outMu.Unlock()
		return nil, DestroyedHub
	default:
	}
	hub.outPipes[out] = buf
	hub.outMu.Unlock()

	connectPipe(connectConfig{
		In:             buf,
		Out:            out,
		PropagateClose: true,
		Filters:        filters,
	})

	return out, nil
}

func (hub *Hub) DestroyOutPipe(p Pipe) {
	hub.outMu.Lock()
	buf, ok := hub.outPipes[p]
	if ok {
		delete(hub.outPipes, p)
		close(buf)
	}
	hub.outMu.Unlock()
}

func (hub *Hub) Destroy() {
	hub.inMu.Lock()
	defer hub.inMu.Unlock()

	close(hub.destroyed)

	for p := range hub.inPipes {
		close(p)
	}

	close(hub.broadcast)
}

type connectConfig struct {
	In, Out        Pipe
	PropagateClose bool
	Filters        []Filter
}

func connectPipe(cfg connectConfig) {
	go func(in, out Pipe, propagateClose bool, filters []Filter) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recovered", r)
			}
			if propagateClose {
				close(out)
			}
		}()

		var ok bool
		var v Message
		var f Filter

		for v = range in {
			ok = true
			for _, f = range filters {
				v, ok = f(v)
				if !ok {
					goto CONTINUE
				}
			}
			out <- v
		CONTINUE:
		}
	}(cfg.In, cfg.Out, cfg.PropagateClose, cfg.Filters)
}

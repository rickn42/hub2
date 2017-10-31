package hub2

import (
	"sync"
)

type Message = interface{}

type Pipe = chan Message

type BufPipe = chan Message

type Done = chan struct{}

type Filter = func(interface{}) (interface{}, bool)

type Hub struct {
	broadcast chan Message

	mu      sync.Mutex
	inPipes map[Pipe]Done

	mu2      sync.RWMutex
	outPipes map[Pipe]BufPipe
}

func NewHub() (hub *Hub) {
	hub = &Hub{
		broadcast: make(chan Message),
		inPipes:   make(map[Pipe]Done),
		outPipes:  make(map[Pipe]BufPipe),
	}

	go func() {
		for {
			select {
			case v := <-hub.broadcast:
				hub.mu2.RLock()
				for _, buf := range hub.outPipes {
					buf <- v
				}
				hub.mu2.RUnlock()
			}
		}
	}()

	return hub
}

func (hub *Hub) MakeInPipe() (c Pipe) {
	c = make(Pipe)
	done := make(Done)

	hub.mu.Lock()
	hub.inPipes[c] = done
	hub.mu.Unlock()

	go func() {
		for v := range c {
			hub.broadcast <- v
		}
	}()

	return c
}

func (hub *Hub) RemoveInPipe(c Pipe) {
	hub.mu.Lock()
	if _, ok := hub.inPipes[c]; ok {
		delete(hub.inPipes, c)
		close(c)
	}
	hub.mu.Unlock()
}

func (hub *Hub) MakeOutPipe(bufSize int, filters ...Filter) (c Pipe) {
	c = make(Pipe)
	buf := make(BufPipe, bufSize)

	hub.mu2.Lock()
	hub.outPipes[c] = buf
	hub.mu2.Unlock()

	go func() {
		var ok bool
		for v := range buf {
			ok = false
			for _, f := range filters {
				v, ok = f(v)
				if !ok {
					goto CONTINUE
				}
			}

			c <- v

		CONTINUE:
		}
	}()

	return c
}

func (hub *Hub) RemoveOutPipe(c Pipe) {
	hub.mu2.Lock()
	if buf, ok := hub.outPipes[c]; ok {
		delete(hub.outPipes, c)
		close(buf)
	}
	hub.mu2.Unlock()
}

func (hub *Hub) Destroy() {
	close(hub.broadcast)
}

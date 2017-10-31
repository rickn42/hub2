package hub2_test

import (
	"fmt"
	"testing"

	"runtime"
	"time"

	"github.com/rickn42/hub2"
)

func TestHubMakeInPipe(t *testing.T) {

	double := func(i interface{}) (interface{}, bool) {
		return i.(int) * 2, true
	}

	hub := hub2.NewHub()
	in := hub.MakeInPipe()
	out := hub.MakeOutPipe(2)
	out2 := hub.MakeOutPipe(2, double)

	go func() {
		for {
			fmt.Println("out:", <-out, ", out2:", <-out2)
		}
	}()

	for i := 1; i < 100; i++ {
		in <- i
	}
}

func BenchmarkHub(b *testing.B) {
	hub := hub2.NewHub()
	const inCnt = 100
	const outCnt = 100

	ins := [inCnt]chan interface{}{}
	for i := range ins {
		ins[i] = hub.MakeInPipe()
	}

	outs := [outCnt]chan interface{}{}
	for i := range outs {
		outs[i] = hub.MakeOutPipe(100)
		go func(p chan interface{}) {
			for range p {
			}
		}(outs[i])
	}

	for i := 0; i < b.N; i++ {
		ins[i%inCnt] <- i
	}
}

func TestHubDestroy(t *testing.T) {
	start := runtime.NumGoroutine()

	h := hub2.NewHub()

	in1 := h.MakeInPipe()
	in2 := h.MakeInPipe()
	out1 := h.MakeOutPipe(1)
	out2 := h.MakeOutPipe(1)

	h.DestroyInPipe(in1)
	h.DestroyInPipe(in2)
	h.DestroyOutPipe(out1)
	h.DestroyOutPipe(out2)

	h.Destroy()

	time.Sleep(time.Millisecond)
	if cur := runtime.NumGoroutine(); cur != start {
		t.Error("Destroy not working", start, cur)
	}
}

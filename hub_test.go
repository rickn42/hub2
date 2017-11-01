package hub2_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/rickn42/hub2"
)

func even(v interface{}) (v2 interface{}, ok bool) {
	if v.(int)%2 == 0 {
		return v, true
	}
	return nil, false
}

func double(v interface{}) (interface{}, bool) {
	return v.(int) * 2, true
}

func TestHubMakeInPipe(t *testing.T) {

	hub := hub2.NewHub()
	in, _ := hub.MakeInPipe(even)
	out, _ := hub.MakeOutPipe(2)
	out2, _ := hub.MakeOutPipe(2, double)
	out3, _ := hub.MakeOutPipe(2, double, double)

	go func() {
		for {
			fmt.Println("out:", <-out, "\tout2:", <-out2, "\tout3:", <-out3)
		}
	}()

	for i := 1; i < 100; i++ {
		in <- i
	}
}

func BenchmarkHub(b *testing.B) {

	b.ReportAllocs()

	const inCnt = 100
	const outCnt = 100

	hub := hub2.NewHub()

	ins := [inCnt]chan interface{}{}
	for i := range ins {
		ins[i], _ = hub.MakeInPipe()
	}

	outs := [outCnt]chan interface{}{}
	for i := range outs {
		outs[i], _ = hub.MakeOutPipe(10)
		go func(p chan interface{}) {
			for range p {
				// for not to block
			}
		}(outs[i])
	}

	// init & cool time
	time.Sleep(time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ins[i%inCnt] <- &i
	}
}

func TestHubDestroy(t *testing.T) {
	start := runtime.NumGoroutine()

	h := hub2.NewHub()

	in1, _ := h.MakeInPipe()
	in2, _ := h.MakeInPipe()
	out1, _ := h.MakeOutPipe(1)
	out2, _ := h.MakeOutPipe(1)
	_, _, _, _ = in1, in2, out1, out2

	// Explicit destroy pipe. (closed)
	h.DestroyInPipes(in1)
	h.DestroyOutPipe(out1)

	// All remain pipes is also destroyed. (closed)
	h.Destroy()

	_, err := h.MakeInPipe()
	if err == nil {
		t.Error("MakeInPipe should be return error if hub destroyed.")
	}
	_, err = h.MakeOutPipe(0)
	if err == nil {
		t.Error("MakeInPipe should be return error if hub destroyed.")
	}

	// Wait a moment for all goroutines end.
	time.Sleep(time.Millisecond)
	if cur := runtime.NumGoroutine(); cur != start {
		t.Error("Destroy not working", start, cur)
	}
}

package hub2_test

import (
	"fmt"
	"testing"

	"github.com/rickn42/hub2"
)

func TestHub_MakeInPipe(t *testing.T) {

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

	const cnt = 1000

	ins := [cnt]chan interface{}{}
	for i := range ins {
		ins[i] = hub.MakeInPipe()
	}

	outs := [cnt]chan interface{}{}
	for i := range outs {
		outs[i] = hub.MakeOutPipe(100)
		go func(p chan interface{}) {
			for range p {
			}
		}(outs[i])
	}

	for i := 0; i < b.N; i++ {
		ins[i%cnt] <- i
	}
}

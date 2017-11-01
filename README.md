# Hub

channel input & output communication library

### How to use 

```go
double := func(i interface{}) (interface{}, bool) {
	return i.(int) * 2, true
}

bufsize := 2

hub := hub2.NewHub()
in, _ := hub.MakeInPipe()
out, _ := hub.MakeOutPipe(bufsize)
out2, _ := hub.MakeOutPipe(bufsize, double)

go func() {
	for {
		fmt.Println("out:", <-out, ", out2:", <-out2)
	}
}()

for i := 1; i < 100; i++ {
	in <- i
}

//out: 1 , out2: 2
//out: 2 , out2: 4
//out: 3 , out2: 6
//out: 4 , out2: 8
//out: 5 , out2: 10
//out: 6 , out2: 12
//out: 7 , out2: 14
//out: 8 , out2: 16
//...
```

### Destroy

WARN: if hub destroyed, all remain pipe in hub is also destroyed. (= closed)

```go
hub := hub2.NewHub()

in1, _ := hub.MakeInPipe()
in2, _ := hub.MakeInPipe()
out1, _ := hub.MakeOutPipe(1)
out2, _ := hub.MakeOutPipe(1)
_, _, _, _ = in1, in2, out1, out2

// Explicit pipe destroy (= close)
hub.DestroyInPipe(in1)
hub.DestroyOutPipe(out1)

// All remain pipes is also destroyed. (= close)
hub.Destroy()

_, err := hub.MakeInPipe()
```

### Benchmark

```go
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
```

```
inCnt = 10, outCnt = 10 
500000	      3254 ns/op	       0 B/op	       0 allocs/op

inCnt = 100, outCnt = 100
100000	     23210 ns/op	       0 B/op	       0 allocs/op 

inCnt = 1000, outCnt = 1000
10000	    227986 ns/op	       0 B/op	       0 allocs/op
```
# Hub

channel input & output communication library

### How to use 

```go
double := func(i interface{}) (interface{}, bool) {
	return i.(int) * 2, true
}

bufsize := 2

hub := hub2.NewHub()
in := hub.MakeInPipe()
out := hub.MakeOutPipe(bufsize)
out2 := hub.MakeOutPipe(bufsize, double)

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

### Benchmark

```go
func BenchmarkHub(b *testing.B) {
	
	const inCnt = 100
	const outCnt = 100
	
	hub := hub2.NewHub()
	
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
```

```
inCnt = 10, outCnt = 10 
300000	      3518 ns/op

inCnt = 100, outCnt = 100 
100000	     21198 ns/op

inCnt = 1000, outCnt = 1000
10000	    210393 ns/op
```
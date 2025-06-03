package main

import (
	"fmt"
	"time"
	"workerpool"
)

func main() {
	p := workerpool2.NewPool(5, workerpool2.WithPreAlloc(false), workerpool2.WithPreAlloc(true))
	time.Sleep(2 * time.Second)
	for i := 0; i < 10; i++ {
		if err := p.Schedule(func() {
			time.Sleep(3 * time.Second)
		}); err != nil {
			fmt.Printf("index: %d, err: %s\n", i, err.Error())
		}
	}
	p.Free()
}

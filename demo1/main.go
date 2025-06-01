package main

import (
	"fmt"
	"time"
	"workerpool1"
)

func main() {
	p := workerpool1.NewPool(5)
	for i := 0; i < 12; i++ {
		err := p.Schedule(func() {
			time.Sleep(3 * time.Second)
		})
		if err != nil {
			fmt.Printf("index: %d, err: %v\n", i, err)
		}
	}
	p.Free()
}

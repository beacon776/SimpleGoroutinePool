package workerpool1

import (
	"errors"
	"fmt"
	"sync"
)

const (
	defaultCapacity = 100
	maxCapacity     = 10000
)

type Task func()

type Pool struct {
	capacity int
	active   chan struct{}
	tasks    chan Task
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewPool(capacity int) *Pool {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}
	p := &Pool{
		capacity: capacity,
		active:   make(chan struct{}, capacity),
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
	}
	fmt.Printf("workerpool start\n")
	go p.run() // 开新的 goroutine，不要阻塞 main goroutine
	// 如果一个函数需要用死循环+特定条件退出来保证它的运行的话，请开一个新线程去运行它，在 main goroutine里运行会阻塞掉后面所有任务的哦
	return p
}

func (p *Pool) run() {

	index := 0
	for {
		select {
		case <-p.quit:
			return
		case p.active <- struct{}{}:
			index++
			go p.newWorker(index) // 在 子goroutine 里执行，不要阻塞 main goroutine
			// 如果一个函数需要用死循环+特定条件退出来保证它的运行的话，请开一个新线程去运行它，在 main goroutine里运行会阻塞掉后面所有任务的哦
		}
	}
}

func (p *Pool) newWorker(index int) {
	p.wg.Add(1)
	defer func() { // 处理错误 + p.wg.Done()这俩事情都得最后再做，直接defer函数里顺手一块做了
		if err := recover(); err != nil {
			fmt.Printf("worker %d: recover panic %s and exit\n", index, err)
			<-p.active // p 的可用 goroutine 少了一个
		}
		p.wg.Done()
	}()
	fmt.Printf("worker %d is ready", index)
	for {
		select {
		case <-p.quit:
			fmt.Printf("worker[%03d]: exit\n", index)
			<-p.active
			return
		case t := <-p.tasks:
			fmt.Printf("worker[%03d]: receive a task\n", index)
			t()
		}
	}
}

var FreedWorkerPool = errors.New("workepool is freed")

func (p *Pool) Schedule(t Task) error {
	select {
	case <-p.quit:
		return FreedWorkerPool
	case p.tasks <- t:
		return nil
	}
}

func (p *Pool) Free() {
	close(p.quit)
	p.wg.Wait()
	fmt.Printf("worker pool free\n")
}

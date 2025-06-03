package workerpool2

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
	preAlloc bool
	block    bool
}

func NewPool(capacity int, opts ...Option) *Pool {

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
	// 初始化*Pool类型的变量p
	for _, opt := range opts {
		opt(p)
	}
	fmt.Printf("workerpool start(preAlloc=%t)\n", p.preAlloc)
	// 根据 preAlloc字段的值 进行相应初始化
	if p.preAlloc {
		for i := 0; i < p.capacity; i++ {
			go p.newWorker(i + 1)
			p.active <- struct{}{}
		}
	}
	go p.run() // 开新的 goroutine，不要阻塞 main goroutine
	// 如果一个函数需要用死循环+特定条件退出来保证它的运行的话，请开一个新线程去运行它，在 main goroutine里运行会阻塞掉后面所有任务的哦
	return p
}

// run 启动 workerpool
func (p *Pool) run() {

	index := len(p.active)
	// 根据 preAlloc字段的值 进行相应初始化
	if !p.preAlloc {
	loop:
		for task := range p.tasks {
			p.returnTask(task)
			select {
			case <-p.quit:
				return
			case p.active <- struct{}{}:
				index++
				p.newWorker(index) // 不要在 子goroutine 里执行此函数！因为这个函数内部本身就开了一个子 goroutine！
			default:
				break loop
			}
		}
	}

	for {
		select {
		case <-p.quit:
			return
		case p.active <- struct{}{}:
			index++
			p.newWorker(index) // 不要在 子goroutine 里执行此函数！因为这个函数内部本身就开了一个子 goroutine！
			// 如果一个函数需要用死循环+特定条件退出来保证它的运行的话，请开一个新线程去运行它，在 main goroutine里运行会阻塞掉后面所有任务的哦
		}
	}
}

// newWorker 负责创建新的 worker goroutine
// 不要在 子goroutine 里执行此函数！因为这个函数内部本身就开了一个子 goroutine！
func (p *Pool) newWorker(index int) {
	p.wg.Add(1)
	go func() {
		defer func() { // 处理错误 + p.wg.Done()这俩事情都得最后再做，直接defer函数里顺手一块做了
			if err := recover(); err != nil {
				fmt.Printf("worker[%03d]: recover panic[%s] and exit\n", index, err)
				<-p.active // 释放当前 goroutine
			}
			p.wg.Done()
		}()
		fmt.Printf("worker[%03d] start\n", index)
		for {
			select {
			case <-p.quit: // 监视 quit channel，当接收到来自quit channel的退出“信号”时，这个worker就会结束运行
				fmt.Printf("worker[%03d]: exit\n", index)
				<-p.active
				return
			case t := <-p.tasks: // 监视 tasks channel，获取最新的 Task 并运行这个 Task
				fmt.Printf("worker[%03d]: receive a task\n", index)
				t()
			}
		}
	}()

}

var (
	FreedWorkerPool = errors.New("workerpool is freed")
	NoUsableWorker  = errors.New("no usable worker")
)

func (p *Pool) Schedule(t Task) error {

	select {
	case <-p.quit:
		return FreedWorkerPool
	case p.tasks <- t:
		return nil
	default:
		if p.block {
			p.tasks <- t // 阻塞直到任务能发送成功
			return nil
		}
		return NoUsableWorker
	}
}

func (p *Pool) Free() {
	close(p.quit)
	p.wg.Wait()
	fmt.Printf("worker pool free\n")
}

func (p *Pool) returnTask(t Task) {
	go func() {
		p.tasks <- t
	}()
}

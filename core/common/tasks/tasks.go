package tasks

import "sync"

type CocurrencyProc interface {
	//add task to background processor
	AddTask(f func())

	//wait until all remianing task done
	StopGracefully()
}

func NewCocurrencyProc(workercnt, pipsize int) CocurrencyProc {
	return &CocurrencyTasks{
		Cocurrentcies: workercnt,
		tasksPipSize:  pipsize,
	}
}

var DefaultCocurrencyProc CocurrencyProc = &CocurrencyTasks{tasksPipSize: 10000}

type CocurrencyTasks struct {
	Cocurrentcies int

	tasksPipSize int
	taskspip     chan func()
	wg           *sync.WaitGroup
}

func (c *CocurrencyTasks) AddTask(f func()) {
	if c.taskspip == nil {
		if c.Cocurrentcies <= 0 {
			c.Cocurrentcies = 3
		}
		if c.tasksPipSize <= 0 {
			c.tasksPipSize = c.Cocurrentcies * 3
		}

		c.taskspip = make(chan func(), c.tasksPipSize)
		c.wg = &sync.WaitGroup{}
		backend := func(id int) {
			defer c.wg.Done()
			for task := range c.taskspip {
				task()
			}
		}

		for i := 0; i < c.Cocurrentcies; i++ {
			c.wg.Add(1)
			go backend(i)
		}
	}

	c.taskspip <- f
}

func (c *CocurrencyTasks) StopGracefully() {
	if c.taskspip != nil {
		close(c.taskspip)
	}
	if c.wg != nil {
		c.wg.Wait()
	}
}

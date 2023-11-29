package task_pool

import (
	"sync"
	"go.uber.org/zap"
)

type TaskPool struct {
	threads   int
	taskQueue chan func() error
	wg        sync.WaitGroup
	errChanel chan error
	logger    *zap.Logger
}

func NewTaskPool(threads int, logger *zap.Logger) *TaskPool {
	return &TaskPool{
		threads:   threads,
		taskQueue: make(chan func() error),
		errChanel: make(chan error),
		logger:    logger,
	}
}

func (pool *TaskPool) run() {
	for task := range pool.taskQueue {
		err := task()
		if err != nil {
			pool.errChanel <- err
		}
		pool.wg.Done()
	}
}

func (pool *TaskPool) Start() {
	for i := 0; i < pool.threads; i++ {
		go pool.run()
	}
	pool.logError()
}

func (pool *TaskPool) AddTask(task func() error) {
	pool.wg.Add(1)
	pool.taskQueue <- task
}

func (pool *TaskPool) Stop() {
	close(pool.taskQueue)
	close(pool.errChanel)
}

func (pool *TaskPool) Wait() {
	pool.wg.Wait()
}


func (pool *TaskPool) logError() {
	for i := 0; i < pool.threads; i++ {
		go func() {
			for err := range pool.errChanel {
				pool.logger.Error("taskPool error", zap.Error(err))
			}
		}()
	}
}
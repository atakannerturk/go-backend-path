package worker

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type Task func(ctx context.Context) error

type WorkerPool struct {
	workers      int
	taskQueue    chan Task
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	tasksTotal   atomic.Int64
	tasksSuccess atomic.Int64
	tasksFailed  atomic.Int64
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (wp *WorkerPool) Start() {
	logger.Info("Starting worker pool", "workers", wp.workers)

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	logger.Info("Worker started", "worker_id", id)

	for {
		select {
		case <-wp.ctx.Done():
			logger.Info("Worker stopped", "worker_id", id)
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				logger.Info("Worker task queue closed", "worker_id", id)
				return
			}

			wp.tasksTotal.Add(1)

			if err := task(wp.ctx); err != nil {
				wp.tasksFailed.Add(1)
				logger.Error("Worker task failed", "worker_id", id, "error", err)
			} else {
				wp.tasksSuccess.Add(1)
			}
		}
	}
}

func (wp *WorkerPool) Submit(task Task) bool {
	select {
	case <-wp.ctx.Done():
		return false
	case wp.taskQueue <- task:
		return true
	default:
		logger.Warn("Worker pool queue is full, task rejected")
		return false
	}
}

func (wp *WorkerPool) Stop() {
	logger.Info("Stopping worker pool")
	close(wp.taskQueue)
	wp.cancel()
	wp.wg.Wait()
	logger.Info("Worker pool stopped",
		"total_tasks", wp.tasksTotal.Load(),
		"successful_tasks", wp.tasksSuccess.Load(),
		"failed_tasks", wp.tasksFailed.Load())
}

func (wp *WorkerPool) Stats() (total, success, failed int64) {
	return wp.tasksTotal.Load(), wp.tasksSuccess.Load(), wp.tasksFailed.Load()
}

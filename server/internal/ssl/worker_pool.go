package ssl

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Task struct {
	Domain   string
	DomainID int
	UserID   int
}

type Result struct {
	Task        Task
	Certificate *SSLCertificate
	Error       error
	CheckedAt   time.Time
}

type WorkerPool struct {
	tasks   chan Task
	results chan Result
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		tasks:   make(chan Task, 100),
		results: make(chan Result, 100),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (wp *WorkerPool) processTask(task Task) Result {
	hostname, err := NewHostname(task.Domain)
	if err != nil {
		return Result{
			Task:      task,
			Error:     err,
			CheckedAt: time.Now(),
		}
	}
	ctx, cancel := context.WithTimeout(wp.ctx, 10*time.Second)
	defer cancel()

	certificate, err := CheckSSLCertificate(ctx, hostname)
	return Result{
		Task:        task,
		Certificate: certificate,
		Error:       err,
		CheckedAt:   time.Now(),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	slog.Info("Worker pool started", "workers", wp.workers)
}

func (wp *WorkerPool) Stop() {
	close(wp.tasks)
	wp.wg.Wait()
	close(wp.results)
	wp.cancel()
	slog.Info("Worker pool stopped")
}

func (wp *WorkerPool) AddTask(task Task) {
	select {
	case wp.tasks <- task:
	case <-wp.ctx.Done():
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for task := range wp.tasks {
		result := wp.processTask(task)
		select {
		case wp.results <- result:
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) GetResults() <-chan Result {
	return wp.results
}

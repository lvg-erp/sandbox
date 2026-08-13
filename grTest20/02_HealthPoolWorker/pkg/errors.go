package pkg

import "errors"

var (
	ErrShuttingDown = errors.New("worker pool is shutting down")
	ErrQueueFull    = errors.New("task queue is full")
	ErrWorkerPanic  = errors.New("worker panicked")
	ErrTimeout      = errors.New("operation timed out")
	ErrTaskFailed   = errors.New("task execution failed")
)

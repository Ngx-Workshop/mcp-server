package tasks

// Dispatch queue (in-memory + adapter interface)
// This file implements a task dispatch queue with in-memory implementation
// and adapter interface for pluggable queue backends

import "context"

type Task struct {
	ID      string
	Type    string
	Payload map[string]any
}

type Result struct {
	TaskID string
	Output map[string]any
	Err    error
}

type Queue interface {
	Enqueue(ctx context.Context, t Task) error
	Dequeue(ctx context.Context) (Task, error)
	Ack(ctx context.Context, taskID string, res Result) error
}

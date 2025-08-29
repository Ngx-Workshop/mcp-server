package agents

import "context"

type Task struct {
	ID         string
	Type       string
	Payload    map[string]any
}

type Result struct {
	TaskID string
	Status string // "ok", "failed", "partial"
	Output map[string]any
	Error  string
}

type Agent interface {
	Name() string
	CanHandle(taskType string) bool
	Execute(ctx context.Context, t Task) (Result, error)
}

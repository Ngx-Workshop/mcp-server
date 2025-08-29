package orchestrator

// orchestration engine entry point

import (
	"context"

	"github.com/ngx-workshop/mcp-server/internal/agents"
	"github.com/ngx-workshop/mcp-server/internal/criteria"
	"github.com/ngx-workshop/mcp-server/internal/tasks"
)

type Orchestrator struct {
	Planner  Planner
	Queue    tasks.Queue
	Registry AgentRegistry
}

type Planner interface {
	Plan(ctx context.Context, c criteria.Criteria) ([]tasks.Task, error)
}

type AgentRegistry interface {
	Select(taskType string) (agents.Agent, bool)
}

package instancer

import (
	"context"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
)

type TaskType int

const (
	TaskTypeStart TaskType = iota
	TaskTypeDestroy
)

type InstanceTask interface {
	Type() TaskType
	ProblemConfig() common.ProblemConfig
	InstanceId() string
	ProblemData() string
	Patch(ctx context.Context, patch *client.PatchInstanceTaskRequest) error
	Complete(ctx context.Context, req *client.CompleteTaskRequest) error
}

type InstanceAdapter interface {
	Name() string
	StartInstance(ctx context.Context, task InstanceTask) error
	DestroyInstance(ctx context.Context, task InstanceTask) error
}

var adapters = make(map[string]InstanceAdapter)

func RegisterAdapter(adapter InstanceAdapter) {
	if _, ok := adapters[adapter.Name()]; ok {
		panic("adapter already registered")
	}
	adapters[adapter.Name()] = adapter
}

func GetAdapter(name string) (InstanceAdapter, bool) {
	adapter, ok := adapters[name]
	return adapter, ok
}

func GetAdapterNames() []string {
	names := make([]string, 0, len(adapters))
	for name := range adapters {
		names = append(names, name)
	}
	return names
}

type RemoteInstanceTask struct {
	taskType    TaskType
	config      common.ProblemConfig
	problemData string
	instanceId  string
}

func (t *RemoteInstanceTask) Type() TaskType {
	return t.taskType
}

func (t *RemoteInstanceTask) ProblemConfig() common.ProblemConfig {
	return t.config
}

func (t *RemoteInstanceTask) ProblemData() string {
	return t.problemData
}

func (t *RemoteInstanceTask) InstanceId() string {
	return t.instanceId
}

func (t *RemoteInstanceTask) Patch(ctx context.Context, patch *client.PatchInstanceTaskRequest) error {
	return client.PatchInstanceTask(ctx, patch)
}

func (t *RemoteInstanceTask) Complete(ctx context.Context, req *client.CompleteTaskRequest) error {
	return client.CompleteTask(ctx, req)
}

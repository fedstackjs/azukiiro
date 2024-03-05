package judge

import (
	"context"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
)

type JudgeTask interface {
	Config() common.ProblemConfig
	ProblemData() string
	SolutionData() string
	Update(ctx context.Context, update *common.SolutionInfo) error
	UploadDetails(ctx context.Context, details *common.SolutionDetails) error
}

type JudgeAdapter interface {
	Name() string
	Judge(ctx context.Context, task JudgeTask) error
}

var adapters = make(map[string]JudgeAdapter)

func RegisterAdapter(adapter JudgeAdapter) {
	if _, ok := adapters[adapter.Name()]; ok {
		panic("adapter already registered")
	}
	adapters[adapter.Name()] = adapter
}

func GetAdapter(name string) (JudgeAdapter, bool) {
	adapter, ok := adapters[name]
	return adapter, ok
}

type judgeTask struct {
	config       common.ProblemConfig
	problemData  string
	solutionData string
}

func (t *judgeTask) Config() common.ProblemConfig {
	return t.config
}

func (t *judgeTask) ProblemData() string {
	return t.problemData
}

func (t *judgeTask) SolutionData() string {
	return t.solutionData
}

func (t *judgeTask) Update(ctx context.Context, update *common.SolutionInfo) error {
	return client.PatchSolutionTask(ctx, update)
}

func (t *judgeTask) UploadDetails(ctx context.Context, details *common.SolutionDetails) error {
	return client.SaveSolutionDetails(ctx, details)
}

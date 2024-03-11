package judge

import (
	"context"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
)

type JudgeTask interface {
	Config() common.ProblemConfig
	Env() map[string]string
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

type RemoteJudgeTask struct {
	config       common.ProblemConfig
	problemData  string
	solutionData string
	solutionId   string
	taskId       string
	env          map[string]string
}

func (t *RemoteJudgeTask) Config() common.ProblemConfig {
	return t.config
}

func (t *RemoteJudgeTask) Env() map[string]string {
	return t.env
}

func (t *RemoteJudgeTask) ProblemData() string {
	return t.problemData
}

func (t *RemoteJudgeTask) SolutionData() string {
	return t.solutionData
}

func (t *RemoteJudgeTask) Update(ctx context.Context, update *common.SolutionInfo) error {
	return client.PatchSolutionTask(ctx, update)
}

func (t *RemoteJudgeTask) UploadDetails(ctx context.Context, details *common.SolutionDetails) error {
	return client.SaveSolutionDetails(ctx, details)
}

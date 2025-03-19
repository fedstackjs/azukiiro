package client

import (
	"context"

	"github.com/fedstackjs/azukiiro/common"
)

type instanceContextKey int

const instanceInjectionKey instanceContextKey = iota

type instanceContext struct {
	InstanceId string
	TaskId     string
}

func WithInstanceTask(ctx context.Context, instanceId string, taskId string) context.Context {
	ctx = context.WithValue(ctx, instanceInjectionKey, instanceContext{
		InstanceId: instanceId,
		TaskId:     taskId,
	})
	return ctx
}

func LoadInstanceTask(ctx context.Context) (string, string) {
	inj := ctx.Value(instanceInjectionKey).(instanceContext)
	return inj.InstanceId, inj.TaskId
}

type PollInstanceRequest struct {
}

type PollInstanceResponse struct {
	TaskId          string               `json:"taskId"`
	InstanceId      string               `json:"instanceId"`
	OrgId           string               `json:"orgId"`
	UserId          string               `json:"userId"`
	ProblemId       string               `json:"problemId"`
	ContestId       string               `json:"contestId"`
	State           int                  `json:"state"`
	ProblemConfig   common.ProblemConfig `json:"problemConfig"`
	ProblemDataUrl  string               `json:"problemDataUrl"`
	ProblemDataHash string               `json:"problemDataHash"`
	ErrMsg          string               `json:"errMsg"`
}

func PollInstance(ctx context.Context, req *PollInstanceRequest) (*PollInstanceResponse, error) {
	res := &PollInstanceResponse{}
	raw, err := http.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(res).
		Post("/api/runner/instance/poll")
	err = loadError(raw, err)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type PatchInstanceTaskRequest struct {
	Message *string `json:"message,omitempty"`
}

const (
	InstanceStateDestroyed  = 0
	InstanceStateDestroying = 1
	InstanceStateAllocated  = 2
	InstanceStateAllocating = 3
	InstanceStateError      = 4
)

const (
	InstanceTaskStatePending    = 0
	InstanceTaskStateQueued     = 1
	InstanceTaskStateInProgress = 2
)

func PatchInstanceTask(ctx context.Context, req *PatchInstanceTaskRequest) error {
	instanceId, taskId := LoadInstanceTask(ctx)
	raw, err := http.R().
		SetContext(ctx).
		SetBody(req).
		Patch("/api/runner/instance/task/" + instanceId + "/" + taskId)
	return loadError(raw, err)
}

type CompleteTaskRequest struct {
	Succeeded bool    `json:"succeeded"`
	Message   *string `json:"message,omitempty"`
}

func CompleteTask(ctx context.Context, req *CompleteTaskRequest) error {
	instanceId, taskId := LoadInstanceTask(ctx)
	raw, err := http.R().
		SetContext(ctx).
		SetBody(req).
		Post("/api/runner/instance/task/" + instanceId + "/" + taskId + "/complete")
	return loadError(raw, err)
}

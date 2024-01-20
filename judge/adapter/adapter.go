package adapter

import (
	"context"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge/adapter/dummy"
	"github.com/fedstackjs/azukiiro/judge/adapter/uoj"
)

type JudgeAdapter interface {
	Name() string
	Judge(ctx context.Context, config common.ProblemConfig, problemData string, solutionData string) error
}

var adapters = make(map[string]JudgeAdapter)

func Register(adapter JudgeAdapter) {
	adapters[adapter.Name()] = adapter
}

func Get(name string) (JudgeAdapter, bool) {
	adapter, ok := adapters[name]
	return adapter, ok
}

func init() {
	Register(&dummy.DummyAdapter{})
	Register(&uoj.UojAdapter{})
}

package judge

import (
	"fmt"

	"github.com/fedstackjs/azukiiro/common"
)

type JudgeError interface {
	error
	Info() *common.SolutionInfo
	Details() *common.SolutionDetails
}

type SimpleSolutionError struct {
	S string
	M string
	D string
}

func (e *SimpleSolutionError) Error() string {
	return e.M
}

func (e *SimpleSolutionError) Info() *common.SolutionInfo {
	return &common.SolutionInfo{
		Status:  e.S,
		Message: e.M,
	}
}

func (e *SimpleSolutionError) Details() *common.SolutionDetails {
	return &common.SolutionDetails{
		Version: 1,
		Summary: fmt.Sprintf("An Error has occurred:\n\n```%s```", e.D),
	}
}

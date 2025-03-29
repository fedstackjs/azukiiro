package flag

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/fedstackjs/azukiiro/utils"
)

func init() {
	judge.RegisterAdapter(&FlagAdapter{})
}

type FlagAdapterConfig struct {
	Flag string `json:"flag"`
}

type FlagAdapter struct{}

func (g *FlagAdapter) Name() string {
	return "flag"
}

type FlagAnswer struct {
	Flag string `json:"flag"`
}

func (g *FlagAdapter) Judge(ctx context.Context, task judge.JudgeTask) error {
	config := task.Config()

	adapterConfig := FlagAdapterConfig{}
	if err := json.Unmarshal([]byte(config.Judge.Config), &adapterConfig); err != nil {
		return err
	}

	solutionDir, err := utils.UnzipTemp(task.SolutionData(), "solution-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(solutionDir)

	answerPath := filepath.Join(solutionDir, "answer.json")
	answer, err := os.ReadFile(answerPath)
	if err != nil {
		task.Update(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Wrong Answer",
			Message: "",
		})
		task.UploadDetails(ctx, &common.SolutionDetails{
			Version: 1,
			Summary: "answer.json not found",
		})
		return nil
	}
	var flagAnswer FlagAnswer
	if err := json.Unmarshal(answer, &flagAnswer); err != nil {
		task.Update(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Wrong Answer",
			Message: "",
		})
		task.UploadDetails(ctx, &common.SolutionDetails{
			Version: 1,
			Summary: "answer.json is not valid",
		})
	}
	if flagAnswer.Flag == adapterConfig.Flag {
		task.Update(ctx, &common.SolutionInfo{
			Score:   100,
			Status:  "Accepted",
			Message: "",
		})
		task.UploadDetails(ctx, &common.SolutionDetails{
			Version: 1,
			Summary: "Accepted",
		})
	} else {
		task.Update(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Wrong Answer",
			Message: "",
		})
		task.UploadDetails(ctx, &common.SolutionDetails{
			Version: 1,
			Summary: "Wrong Answer",
		})
	}

	return nil
}

package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type judgeArgs struct {
	problemConfig string
	problemData   string
	solutionData  string
}

type localJudgeTask struct {
	config       common.ProblemConfig
	problemData  string
	solutionData string
}

func (t *localJudgeTask) Config() common.ProblemConfig {
	return t.config
}

func (t *localJudgeTask) ProblemData() string {
	return t.problemData
}

func (t *localJudgeTask) SolutionData() string {
	return t.solutionData
}

func (t *localJudgeTask) Update(ctx context.Context, update *common.SolutionInfo) error {
	str, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		logrus.Warnf("Failed to marshal update: %v", err)
	}
	logrus.Infof("Update: \n%s\n", str)
	return nil
}

func (t *localJudgeTask) UploadDetails(ctx context.Context, details *common.SolutionDetails) error {
	str, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		logrus.Warnf("Failed to marshal details: %v", err)
	}
	logrus.Infof("Details: \n%s\n", str)
	return nil
}

func runJudge(ctx context.Context, regArgs *judgeArgs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		content, err := os.ReadFile(regArgs.problemConfig)
		if err != nil {
			return err
		}
		problemConfig := common.ProblemConfig{}
		if err := json.Unmarshal(content, &problemConfig); err != nil {
			return err
		}
		regArgs.problemData, err = filepath.Abs(regArgs.problemData)
		if err != nil {
			return err
		}
		regArgs.solutionData, err = filepath.Abs(regArgs.solutionData)
		if err != nil {
			return err
		}
		task := &localJudgeTask{
			config:       problemConfig,
			problemData:  regArgs.problemData,
			solutionData: regArgs.solutionData,
		}
		adapter, ok := judge.GetAdapter(problemConfig.Judge.Adapter)
		if !ok {
			logrus.Fatalf("Judge adapter %v not found", problemConfig.Judge.Adapter)
		}
		return adapter.Judge(ctx, task)
	}
}

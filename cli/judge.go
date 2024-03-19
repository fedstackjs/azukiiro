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

func init() {
	commands = append(commands, &judgeCmd{})
}

type judgeCmd struct{}

func (c *judgeCmd) Mount(ctx context.Context, root *cobra.Command) {
	var judgeArgs judgeArgs
	judgeCmd := &cobra.Command{
		Use:   "judge",
		Short: "Run the judge locally",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runJudge(ctx, &judgeArgs),
	}
	judgeCmd.Flags().StringVar(&judgeArgs.problemConfig, "problem-config", "", "Problem config file")
	judgeCmd.MarkFlagRequired("problem-config")
	judgeCmd.Flags().StringVar(&judgeArgs.problemData, "problem-data", "", "Problem data file")
	judgeCmd.MarkFlagRequired("problem-data")
	judgeCmd.Flags().StringVar(&judgeArgs.solutionData, "solution-data", "", "Solution data file")
	judgeCmd.MarkFlagRequired("solution-data")
	judgeCmd.Flags().StringVar(&judgeArgs.env, "env", "{}", "Environment variables")
	root.AddCommand(judgeCmd)
}

type judgeArgs struct {
	problemConfig string
	problemData   string
	solutionData  string
	env           string
}

type localJudgeTask struct {
	config       common.ProblemConfig
	problemData  string
	solutionData string
	env          map[string]string
}

func (t *localJudgeTask) Config() common.ProblemConfig {
	return t.config
}

func (t *localJudgeTask) Env() map[string]string {
	return t.env
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
	logrus.Infof("Update: \n%s\n\n", str)
	return nil
}

func (t *localJudgeTask) UploadDetails(ctx context.Context, details *common.SolutionDetails) error {
	str, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		logrus.Warnf("Failed to marshal details: %v", err)
	}
	logrus.Infof("Details: \n%s\n\n", str)
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
		env := make(map[string]string)
		if err := json.Unmarshal([]byte(regArgs.env), &env); err != nil {
			logrus.Warnf("Failed to parse env: %v", err)
		}
		task := &localJudgeTask{
			config:       problemConfig,
			problemData:  regArgs.problemData,
			solutionData: regArgs.solutionData,
			env:          env,
		}
		adapter, ok := judge.GetAdapter(problemConfig.Judge.Adapter)
		if !ok {
			logrus.Fatalf("Judge adapter %v not found", problemConfig.Judge.Adapter)
		}
		return adapter.Judge(ctx, task)
	}
}

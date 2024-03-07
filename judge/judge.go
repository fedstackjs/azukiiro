package judge

import (
	"context"
	"fmt"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/sirupsen/logrus"
)

func judge(ctx context.Context, res *client.PollSolutionResponse) error {
	err := client.PatchSolutionTask(ctx, &common.SolutionInfo{
		Score:   0,
		Status:  "Running",
		Message: "Preparing solution",
	})
	if err != nil {
		return err
	}
	problemData, err := storage.PrepareFile(ctx, res.ProblemDataUrl, res.ProblemDataHash)
	if err != nil {
		return err
	}
	solutionData, err := storage.PrepareFile(ctx, res.SolutionDataUrl, res.SolutionDataHash)
	if err != nil {
		return err
	}
	err = client.PatchSolutionTask(ctx, &common.SolutionInfo{
		Score:   0,
		Status:  "Running",
		Message: "Judging",
	})
	if err != nil {
		return err
	}
	adapter, ok := GetAdapter(res.ProblemConfig.Judge.Adapter)
	if !ok {
		return client.PatchSolutionTask(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Error",
			Message: "Judge adapter not found",
		})
	}
	task := &RemoteJudgeTask{
		config:       res.ProblemConfig,
		problemData:  problemData,
		solutionData: solutionData,
	}
	return adapter.Judge(ctx, task)
}

func Poll(ctx context.Context) (bool, error) {
	res, err := client.PollSolution(ctx, &client.PollSolutionRequest{})
	if err != nil {
		return false, err
	}

	if res.TaskId == "" {
		// No pending tasks
		return false, nil
	}

	ctx = client.WithSolutionTask(ctx, res.SolutionId, res.TaskId)

	if res.ErrMsg != "" {
		// Server side error occurred
		client.PatchSolutionTask(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Error",
			Message: "Server side error occurred",
		})
		client.CompleteSolutionTask(ctx)
		return true, nil
	}

	logrus.Println("Got task:", res.TaskId)
	logrus.Println("SolutionId:", res.SolutionId)

	err = judge(ctx, res)
	if err != nil {
		logrus.Println("Judge finished with error:", err)
		err = client.SaveSolutionDetails(ctx, &common.SolutionDetails{
			Version: 1,
			Jobs:    []*common.SolutionDetailsJob{},
			Summary: fmt.Sprintf("An Error has occurred:\n\n```\n%s\n```", err),
		})
		if err != nil {
			logrus.Println("Save details failed:", err)
		}
		err = client.PatchSolutionTask(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Error",
			Message: "Judge error",
		})
		if err != nil {
			logrus.Println("Patch task failed:", err)
		}
	} else {
		logrus.Println("Judge finished")
	}
	err = client.CompleteSolutionTask(ctx)
	if err != nil {
		logrus.Println("Complete task failed:", err)
	}

	return true, nil
}

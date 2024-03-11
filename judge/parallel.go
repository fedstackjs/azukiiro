package judge

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/sirupsen/logrus"
)

func parallelPoll(ctx context.Context) (*RemoteJudgeTask, bool, error) {
	res, err := client.PollSolution(ctx, &client.PollSolutionRequest{})
	if err != nil {
		return nil, false, err
	}

	if res.TaskId == "" {
		// No pending tasks
		return nil, false, nil
	}

	ctx = client.WithSolutionTask(ctx, res.SolutionId, res.TaskId)

	logrus.Println("Got task:", res.TaskId)
	logrus.Println("SolutionId:", res.SolutionId)

	task := &RemoteJudgeTask{
		config:       res.ProblemConfig,
		problemData:  "",
		solutionData: "",
		solutionId:   res.SolutionId,
		taskId:       res.TaskId,
		env: map[string]string{
			"userId": res.UserId,
		},
	}

	if res.ErrMsg != "" {
		return task, true, fmt.Errorf("server side error occurred: %s", res.ErrMsg)
	}

	_, ok := GetAdapter(res.ProblemConfig.Judge.Adapter)
	if !ok {
		err := client.PatchSolutionTask(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Error",
			Message: "Judge adapter not found",
		})
		return task, true, err
	}

	err = client.PatchSolutionTask(ctx, &common.SolutionInfo{
		Score:   0,
		Status:  "Queued",
		Message: "Preparing solution",
	})
	if err != nil {
		return task, true, err
	}

	task.problemData, err = storage.PrepareFile(ctx, res.ProblemDataUrl, res.ProblemDataHash)
	if err != nil {
		return task, true, err
	}
	task.solutionData, err = storage.PrepareFile(ctx, res.SolutionDataUrl, res.SolutionDataHash)
	if err != nil {
		return task, true, err
	}

	err = client.PatchSolutionTask(ctx, &common.SolutionInfo{
		Score:   0,
		Status:  "Queued",
		Message: "Waiting for judge",
	})
	if err != nil {
		return task, true, err
	}

	return task, true, nil
}

func ParallelPoller(ctx context.Context, pollInterval float32, queue chan<- *RemoteJudgeTask) {
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	for {
		task, cont, err := parallelPoll(ctx)
		if task != nil {
			if err != nil {
				ctx := client.WithSolutionTask(ctx, task.solutionId, task.taskId)
				logrus.Println("Judge skipped with error:", err)

				if err = client.SaveSolutionDetails(ctx, &common.SolutionDetails{
					Version: 1,
					Jobs:    []*common.SolutionDetailsJob{},
					Summary: fmt.Sprintf("An Error has occurred:\n\n```\n%s\n```", err),
				}); err != nil {
					logrus.Warnln("Save details failed:", err)
				}

				if err := client.PatchSolutionTask(ctx, &common.SolutionInfo{
					Score:   0,
					Status:  "Error",
					Message: "Judge error",
				}); err != nil {
					logrus.Warnln("Patch task failed:", err)
				}

				if err := client.CompleteSolutionTask(ctx); err != nil {
					logrus.Warnln("Complete task failed:", err)
				}
			} else {
				queue <- task
			}
		} else {
			if err != nil {
				logrus.Warnln("Failed to poll:", err)
			}
		}
		if cont {
			continue
		}
		waitDur := time.Duration(0)
		if !cont {
			waitDur = time.Duration(pollInterval) * time.Second
		}
		timer := time.NewTimer(waitDur)
		select {
		case <-signalCtx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			logrus.Info("Stopping parallel poller")
			close(queue)
			return
		case <-timer.C:
		}
	}
}

func parallelJudge(ctx context.Context, task *RemoteJudgeTask) error {
	adapter, ok := GetAdapter(task.config.Judge.Adapter)
	if !ok {
		return client.PatchSolutionTask(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Error",
			Message: "Judge adapter not found",
		})
	}
	return adapter.Judge(ctx, task)
}

func ParallelJudger(ctx context.Context, queue <-chan *RemoteJudgeTask) {
	for task := range queue {
		ctx := client.WithSolutionTask(ctx, task.solutionId, task.taskId)
		err := parallelJudge(ctx, task)
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
	}
	logrus.Info("Stopping parallel judger")
}

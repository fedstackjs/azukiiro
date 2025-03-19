package instancer

import (
	"context"
	"fmt"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/fedstackjs/azukiiro/utils"
	"github.com/sirupsen/logrus"
)

func startInstance(ctx context.Context, res *client.PollInstanceResponse) error {
	message := "Starting instance\n"
	client.PatchInstanceTask(ctx, &client.PatchInstanceTaskRequest{
		Message: &message,
	})

	updateMessage := func() {
		message += " ✅\n"
		client.PatchInstanceTask(ctx, &client.PatchInstanceTaskRequest{
			Message: &message,
		})
	}

	updateError := func(err error) error {
		message += fmt.Sprintf(" ❌\n\nError:\n\n```%s```\n", err)
		client.CompleteTask(ctx, &client.CompleteTaskRequest{
			Succeeded: false,
			Message:   &message,
		})
		return nil
	}

	message += "- Prepare problem data"
	problemData, err := storage.PrepareFile(ctx, res.ProblemDataUrl, res.ProblemDataHash)
	if err != nil {
		logrus.Infof("Failed to prepare problem data: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Prepare adapter"
	if res.ProblemConfig.InstanceLabel == nil || res.ProblemConfig.Instance == nil {
		return updateError(fmt.Errorf("instance not configured"))
	}
	adapter, ok := GetAdapter(*res.ProblemConfig.InstanceLabel)
	if !ok {
		return updateError(fmt.Errorf("adapter not found for label: %s", *res.ProblemConfig.InstanceLabel))
	}
	updateMessage()

	task := &RemoteInstanceTask{
		taskType:    TaskTypeStart,
		config:      res.ProblemConfig,
		problemData: problemData,
		instanceId:  res.InstanceId,
	}

	err = adapter.StartInstance(ctx, task)
	if err != nil {
		return updateError(err)
	}

	client.CompleteTask(ctx, &client.CompleteTaskRequest{
		Succeeded: true,
		Message:   utils.ToPtr(fmt.Sprintf("%s✅\nInstance started successfully\n", message)),
	})
	return nil
}

func destroyInstance(ctx context.Context, res *client.PollInstanceResponse) error {
	message := "Destroying instance\n"
	client.PatchInstanceTask(ctx, &client.PatchInstanceTaskRequest{
		Message: &message,
	})

	updateMessage := func() {
		message += " ✅\n"
		client.PatchInstanceTask(ctx, &client.PatchInstanceTaskRequest{
			Message: &message,
		})
	}

	updateError := func(err error) error {
		message += fmt.Sprintf(" ❌\n\nError:\n\n```%s```\n", err)
		client.CompleteTask(ctx, &client.CompleteTaskRequest{
			Succeeded: false,
			Message:   &message,
		})
		return nil
	}

	message += "- Prepare adapter"
	if res.ProblemConfig.InstanceLabel == nil || res.ProblemConfig.Instance == nil {
		return updateError(fmt.Errorf("instance not configured"))
	}
	adapter, ok := GetAdapter(*res.ProblemConfig.InstanceLabel)
	if !ok {
		return updateError(fmt.Errorf("adapter not found for label: %s", *res.ProblemConfig.InstanceLabel))
	}
	updateMessage()

	task := &RemoteInstanceTask{
		taskType:    TaskTypeDestroy,
		config:      res.ProblemConfig,
		problemData: "",
		instanceId:  res.InstanceId,
	}

	err := adapter.DestroyInstance(ctx, task)
	if err != nil {
		return updateError(err)
	}

	client.CompleteTask(ctx, &client.CompleteTaskRequest{
		Succeeded: true,
		Message:   utils.ToPtr(fmt.Sprintf("%s✅\nInstance destroyed successfully\n", message)),
	})
	return nil
}

func handlePollError(ctx context.Context, err string) error {
	client.CompleteTask(ctx, &client.CompleteTaskRequest{
		Succeeded: false,
		Message:   utils.ToPtr("Server side error occurred"),
	})
	return nil
}

func Poll(ctx context.Context) (bool, error) {
	res, err := client.PollInstance(ctx, &client.PollInstanceRequest{})
	if err != nil {
		return false, err
	}
	if res.TaskId == "" {
		return false, nil
	}

	ctx = client.WithInstanceTask(ctx, res.InstanceId, res.TaskId)

	if res.ErrMsg != "" {
		return true, handlePollError(ctx, res.ErrMsg)
	}

	logrus.Println("Got task   :", res.TaskId)
	logrus.Println("- ProblemId:", res.ProblemId)
	logrus.Println("- Hash     :", res.ProblemDataHash)
	logrus.Println("- Label    :", res.ProblemConfig.InstanceLabel)
	logrus.Println("- State    :", res.State)

	var actionErr error
	switch res.State {
	case client.InstanceStateAllocating:
		actionErr = startInstance(ctx, res)
	case client.InstanceStateDestroying:
		actionErr = destroyInstance(ctx, res)
	default:
		actionErr = fmt.Errorf("unexpected instance state: %d", res.State)
	}

	if actionErr != nil {
		logrus.Printf("Failed to handle instance task: %v", actionErr)
		client.CompleteTask(ctx, &client.CompleteTaskRequest{
			Succeeded: false,
			Message:   utils.ToPtr(fmt.Sprintf("Task error:\n```%s```", actionErr)),
		})
	}
	return true, nil
}

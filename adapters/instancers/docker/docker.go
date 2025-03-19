package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/instancer"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/fedstackjs/azukiiro/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	instancer.RegisterAdapter(&DockerAdapter{})
}

type DockerAdapterConfig struct {
	StartTimeout int    `json:"startTimeout"`
	DomainSuffix string `json:"domainSuffix"`
	NetworkName  string `json:"networkName"`
}

type DockerAdapter struct{}

func (d *DockerAdapter) Name() string {
	return "docker"
}

func (d *DockerAdapter) StartInstance(ctx context.Context, task instancer.InstanceTask) error {
	message := "Starting docker instance\n"
	task.Patch(ctx, &client.PatchInstanceTaskRequest{
		Message: &message,
	})

	updateMessage := func() {
		message += " ✅\n"
		task.Patch(ctx, &client.PatchInstanceTaskRequest{
			Message: &message,
		})
	}

	updateError := func(err error) error {
		message += fmt.Sprintf(" ❌\n\nError:\n\n```%s```\n", err)
		task.Complete(ctx, &client.CompleteTaskRequest{
			Succeeded: false,
			Message:   &message,
		})
		return nil
	}

	message += "- Parse problem config"
	viper.SetDefault("instancer.docker.startTimeout", 30)
	viper.SetDefault("instancer.docker.domainSuffix", ".inst.localhost")
	viper.SetDefault("instancer.docker.networkName", "caddy")
	config := &DockerAdapterConfig{
		StartTimeout: viper.GetInt("instancer.docker.startTimeout"),
		DomainSuffix: viper.GetString("instancer.docker.domainSuffix"),
		NetworkName:  viper.GetString("instancer.docker.networkName"),
	}
	// Currently, do not load config from problem
	// if err := json.Unmarshal([]byte(task.ProblemConfig().Instance.Config), config); err != nil {
	// 	logrus.Infof("Failed to parse problem config: %v", err)
	// 	return updateError(err)
	// }
	updateMessage()

	message += "- Extract problem data"
	problemDir, err := utils.UnzipTemp(task.ProblemData(), "problem")
	if err != nil {
		logrus.Infof("Failed to extract problem data: %v", err)
		return updateError(err)
	}
	defer os.RemoveAll(problemDir)
	updateMessage()

	message += "- Load Docker Compose Project"
	composeTemplateDir := filepath.Join(problemDir, "compose")
	projectName := getProjectNameForTask(task, config)
	projectDomain := getProjectDomainForTask(task, config)
	project, err := LoadComposeProject(ctx, composeTemplateDir, projectName)
	if err != nil {
		logrus.Infof("Failed to load Docker Compose project: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Initialize Docker Compose Instance"
	err = InitComposeInstance(ctx, task.InstanceId(), composeTemplateDir)
	if err != nil {
		logrus.Infof("Failed to initialize Docker Compose instance: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Transform Docker Compose Project"
	err = TransformComposeProject(ctx, project, config, projectDomain)
	if err != nil {
		logrus.Infof("Failed to transform Docker Compose project: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Write Docker Compose Project"
	err = WriteComposeProject(ctx, task.InstanceId(), project)
	if err != nil {
		logrus.Infof("Failed to write Docker Compose project: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Start Docker Compose Project"
	err = StartComposeProject(ctx, task.InstanceId(), config.StartTimeout)
	if err != nil {
		logrus.Infof("Failed to start Docker Compose project: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "\nInstance info:\n"
	message += fmt.Sprintf("- Domain: [%s](http://%s)\n", projectDomain, projectDomain)

	task.Complete(ctx, &client.CompleteTaskRequest{
		Succeeded: true,
		Message:   &message,
	})

	return nil
}

func (d *DockerAdapter) DestroyInstance(ctx context.Context, task instancer.InstanceTask) error {
	message := "Destroying docker instance\n"
	task.Patch(ctx, &client.PatchInstanceTaskRequest{
		Message: &message,
	})

	updateMessage := func() {
		message += " ✅\n"
		task.Patch(ctx, &client.PatchInstanceTaskRequest{
			Message: &message,
		})
	}

	updateError := func(err error) error {
		message += fmt.Sprintf(" ❌\n\nError:\n\n```%s```\n", err)
		task.Complete(ctx, &client.CompleteTaskRequest{
			Succeeded: false,
			Message:   &message,
		})
		return nil
	}

	// Check if instance directory exists
	instancePath := filepath.Join(storage.GetRootPath(), "instances", task.InstanceId())
	_, err := os.Stat(instancePath)
	if os.IsNotExist(err) {
		message += "Instance directory does not exist, already cleaned up\n"
		task.Complete(ctx, &client.CompleteTaskRequest{
			Succeeded: true,
			Message:   &message,
		})
		return nil
	}
	if err != nil {
		return updateError(fmt.Errorf("failed to check instance directory: %w", err))
	}

	message += "- Stop Docker Compose Project"
	err = StopComposeProject(ctx, task.InstanceId())
	if err != nil {
		logrus.Infof("Failed to stop Docker Compose project: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Remove Docker Compose Project"
	err = RemoveComposeProject(ctx, task.InstanceId())
	if err != nil {
		logrus.Infof("Failed to remove Docker Compose project: %v", err)
		return updateError(err)
	}
	updateMessage()

	message += "- Clean up instance directory"
	err = CleanupInstanceDirectory(ctx, task.InstanceId())
	if err != nil {
		logrus.Infof("Failed to clean up instance directory: %v", err)
		return updateError(err)
	}
	updateMessage()

	task.Complete(ctx, &client.CompleteTaskRequest{
		Succeeded: true,
		Message:   &message,
	})

	return nil
}

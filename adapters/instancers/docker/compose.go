package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/sirupsen/logrus"
)

func LoadComposeProject(ctx context.Context, path string, projectName string) (*types.Project, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", path, err)
	}

	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}

	composePath := filepath.Join(path, "compose.yml")
	if _, err := os.Stat(composePath); err != nil {
		return nil, fmt.Errorf("no compose.yml found in directory %s", path)
	}

	options, err := cli.NewProjectOptions(
		[]string{composePath},
		cli.WithWorkingDirectory(path),
		cli.WithName(projectName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return options.LoadProject(ctx)
}

func InitComposeInstance(ctx context.Context, instanceId string, path string) error {
	instancePath := filepath.Join(storage.GetRootPath(), "instances", instanceId)
	err := os.MkdirAll(instancePath, 0700)
	if err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return fmt.Errorf("failed to open root: %w", err)
	}
	defer root.Close()
	if err := os.CopyFS(instancePath, root.FS()); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}
	return nil
}

func TransformComposeProject(ctx context.Context, project *types.Project, config *DockerAdapterConfig, domain string) error {
	ingressServiceName := ""
	for serviceName, service := range project.Services {
		if service.ContainerName != "" {
			return fmt.Errorf("container_name is not supported in service %s", serviceName)
		}
		for label := range service.Labels {
			if strings.HasPrefix(label, "caddy_") {
				return fmt.Errorf("multiple caddy config is not supported in service %s: %s", serviceName, label)
			}
			if strings.HasPrefix(label, "caddy") {
				if ingressServiceName != "" && ingressServiceName != serviceName {
					return fmt.Errorf("multiple ingress service found: %s, %s", ingressServiceName, serviceName)
				}
				ingressServiceName = serviceName
			}
		}
	}
	for serviceName, service := range project.Services {
		if serviceName == ingressServiceName {
			service.Labels["caddy"] = fmt.Sprintf("http://%s", domain)
		} else {
			for label := range service.Labels {
				if strings.HasPrefix(label, "caddy") {
					logrus.Infof("removing label %s from service %s", label, serviceName)
					delete(service.Labels, label)
				}
			}
		}
	}
	caddy, ok := project.Networks["caddy"]
	if !ok {
		return fmt.Errorf("caddy network not found")
	}
	caddy.External = true
	caddy.Name = config.NetworkName
	return nil
}

func WriteComposeProject(ctx context.Context, instanceId string, project *types.Project) error {
	instancePath := filepath.Join(storage.GetRootPath(), "instances", instanceId)
	composePath := filepath.Join(instancePath, "compose.yml")

	projectYAML, err := project.MarshalYAML()
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}

	err = os.WriteFile(composePath, projectYAML, 0600)
	if err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

func StartComposeProject(ctx context.Context, instanceId string, timeout int) error {
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	name := "docker"
	args := []string{"compose", "-f", "compose.yml", "up", "-d"}

	cmd := exec.CommandContext(execCtx, name, args...)
	cmd.Dir = filepath.Join(storage.GetRootPath(), "instances", instanceId)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start compose project: %w", err)
	}

	return nil
}

func StopComposeProject(ctx context.Context, instanceId string) error {
	name := "docker"
	args := []string{"compose", "-f", "compose.yml", "down"}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = filepath.Join(storage.GetRootPath(), "instances", instanceId)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop compose project: %w", err)
	}

	return nil
}

func RemoveComposeProject(ctx context.Context, instanceId string) error {
	name := "docker"
	args := []string{"compose", "-f", "compose.yml", "down", "-v", "--remove-orphans"}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = filepath.Join(storage.GetRootPath(), "instances", instanceId)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove compose project: %w", err)
	}

	return nil
}

func CleanupInstanceDirectory(ctx context.Context, instanceId string) error {
	instancePath := filepath.Join(storage.GetRootPath(), "instances", instanceId)
	if err := os.RemoveAll(instancePath); err != nil {
		return fmt.Errorf("failed to cleanup instance directory: %w", err)
	}
	return nil
}

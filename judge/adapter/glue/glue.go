package glue

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/sirupsen/logrus"
)

const (
	scriptHeader = "#!/bin/bash\n\nset -ex\n\n"
)

type GlueAdapterConfig struct {
	Command []string `json:"command"`
	Run     string   `json:"run"`
	Timeout int      `json:"timeout"`
}

type GlueAdapter struct{}

func (g *GlueAdapter) Name() string {
	return "glue"
}

func (g *GlueAdapter) reportHandler(ctx context.Context, pipe *os.File) {
	reader := bufio.NewReader(pipe)
	request := client.PatchSolutionTaskRequest{}
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}
		k, v, err := parseKVLine(line)
		if err != nil {
			logrus.Warnf("Failed to parse report line: %v", err)
			continue
		}
		switch k {
		case "score":
			score, err := strconv.ParseFloat(v, 64)
			if err != nil {
				logrus.Warnf("Failed to parse score: %v", err)
				continue
			}
			if score < 0 || score > 100 {
				logrus.Warnf("Invalid score: %v", score)
				continue
			}
			request.Score = score
		case "status":
			request.Status = v
		case "message":
			request.Message = v
		case "metrics":
			metrics := make(map[string]float64)
			if err := json.Unmarshal([]byte(v), &metrics); err != nil {
				logrus.Warnf("Failed to parse metrics: %v", err)
				continue
			}
			request.Metrics = &metrics
		case "commit":
			if err := client.PatchSolutionTask(ctx, &request); err != nil {
				logrus.Warnf("Failed to commit report: %v", err)
			}
		}
	}
}

func (g *GlueAdapter) Judge(ctx context.Context, config common.ProblemConfig, problemData string, solutionData string) error {
	adapterConfig := GlueAdapterConfig{}
	if err := json.Unmarshal([]byte(config.Judge.Config), &adapterConfig); err != nil {
		return err
	}

	dir, err := storage.MkdirTemp("glue-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	scriptPath := filepath.Join(dir, "run.sh")
	if adapterConfig.Run != "" {
		fullScript := scriptHeader + adapterConfig.Run
		os.WriteFile(scriptPath, []byte(fullScript), 0700)
		// Warn if Command is set
		if len(adapterConfig.Command) > 0 {
			logrus.Warnf("Command is set, run script will be ignored")
		} else {
			adapterConfig.Command = []string{scriptPath}
		}
	}

	detailsPath := filepath.Join(dir, "details.json")
	details := common.SolutionDetails{
		Jobs:    []*common.SolutionDetailsJob{},
		Summary: "",
	}
	detailsJson, err := json.Marshal(details)
	if err != nil {
		return err
	}
	if err := os.WriteFile(detailsPath, detailsJson, 0600); err != nil {
		return err
	}

	reportPath := filepath.Join(dir, "report")
	if err := syscall.Mkfifo(reportPath, 0600); err != nil {
		return err
	}
	defer os.Remove(reportPath)
	pipe, err := os.OpenFile(reportPath, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer pipe.Close()

	go g.reportHandler(ctx, pipe)

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(adapterConfig.Timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(execCtx, adapterConfig.Command[0], adapterConfig.Command[1:]...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, "GLUE_PROBLEM_DATA="+problemData)
	cmd.Env = append(cmd.Env, "GLUE_SOLUTION_DATA="+solutionData)
	cmd.Env = append(cmd.Env, "GLUE_REPORT="+reportPath)
	cmd.Env = append(cmd.Env, "GLUE_DETAILS="+detailsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()

	detailsJson, err = os.ReadFile(detailsPath)
	if err != nil {
		logrus.Warnf("Failed to read details: %v", err)
	}
	if err := json.Unmarshal(detailsJson, &details); err != nil {
		logrus.Warnf("Failed to unmarshal details: %v", err)
	}

	if cmdErr != nil {
		if err := client.PatchSolutionTask(ctx, &client.PatchSolutionTaskRequest{
			Score:   0,
			Status:  "Judge Error",
			Message: "Judge process exited abnormally",
		}); err != nil {
			logrus.Warnf("Failed to report error: %v", err)
		}

		details.Summary += "\n\n"
		details.Summary += fmt.Sprintf("Judge process exited abnormally: %v", cmdErr)
	}

	if err := client.SaveSolutionDetails(ctx, &details); err != nil {
		logrus.Warnf("Failed to save details: %v", err)
	}

	return nil
}

//go:build !windows

package deno

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/fedstackjs/azukiiro/utils"
	"github.com/sirupsen/logrus"
)

func init() {
	judge.RegisterAdapter(&DenoAdapter{})
}

type DenoAdapterConfig struct {
	Script  string `json:"script"`
	Timeout int    `json:"timeout"`
}

type DenoAdapter struct{}

func (g *DenoAdapter) Name() string {
	return "deno"
}

func (g *DenoAdapter) reportHandler(ctx context.Context, task judge.JudgeTask, pipe *os.File) {
	reader := bufio.NewReader(pipe)
	request := common.SolutionInfo{}
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
			if err := task.Update(ctx, &request); err != nil {
				logrus.Warnf("Failed to commit report: %v", err)
			}
		}
	}
}

func (g *DenoAdapter) Judge(ctx context.Context, task judge.JudgeTask) error {
	config := task.Config()

	adapterConfig := DenoAdapterConfig{}
	if err := json.Unmarshal([]byte(config.Judge.Config), &adapterConfig); err != nil {
		return err
	}

	problemDir, err := utils.UnzipTemp(task.ProblemData(), "problem-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(problemDir)

	scriptPath := filepath.Join(problemDir, adapterConfig.Script)
	if _, err := os.Stat(scriptPath); err != nil {
		return err
	}

	solutionDir, err := utils.UnzipTemp(task.SolutionData(), "solution-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(solutionDir)

	workDir, err := storage.MkdirTemp("work-deno-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	detailsPath := filepath.Join(workDir, "details.json")
	details := common.SolutionDetails{
		Version: 1,
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

	reportPath := filepath.Join(workDir, "report")
	if err := syscall.Mkfifo(reportPath, 0600); err != nil {
		return err
	}
	defer os.Remove(reportPath)
	pipe, err := os.OpenFile(reportPath, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer pipe.Close()

	go g.reportHandler(ctx, task, pipe)

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(adapterConfig.Timeout)*time.Second)
	defer cancel()
	name := "deno"
	args := []string{"run"}
	args = append(args, "--allow-read="+strings.Join([]string{problemDir, solutionDir}, ","))
	args = append(args, "--allow-write="+strings.Join([]string{reportPath, detailsPath}, ","))
	args = append(args, "--no-prompt")
	args = append(args, additionalDenoArgs...)
	args = append(args, scriptPath)

	cmd := exec.CommandContext(execCtx, name, args...)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "AZUKIIRO_PROBLEM_DATA_DIR="+problemDir)
	cmd.Env = append(cmd.Env, "AZUKIIRO_SOLUTION_DATA_DIR="+solutionDir)
	cmd.Env = append(cmd.Env, "AZUKIIRO_REPORT="+reportPath)
	cmd.Env = append(cmd.Env, "AZUKIIRO_DETAILS="+detailsPath)
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
		if err := task.Update(ctx, &common.SolutionInfo{
			Score:   0,
			Status:  "Judge Error",
			Message: "Judge process exited abnormally",
		}); err != nil {
			logrus.Warnf("Failed to report error: %v", err)
		}

		details.Summary += "\n\n"
		details.Summary += fmt.Sprintf("Judge process exited abnormally: %v", cmdErr)
	}

	if err := task.UploadDetails(ctx, &details); err != nil {
		logrus.Warnf("Failed to save details: %v", err)
	}

	return nil
}

func parseKVLine(line []byte) (string, string, error) {
	parts := bytes.SplitN(line, []byte("="), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid line: does not contain key and value separated by '='")
	}
	key := string(parts[0])
	value := string(parts[1])
	return key, value, nil
}

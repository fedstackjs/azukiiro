//go:build !windows

package uoj

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/fedstackjs/azukiiro/utils"
	"github.com/sirupsen/logrus"
)

func init() {
	judge.RegisterAdapter(&UojAdapter{})
}

type UojAdapter struct{}

func (u *UojAdapter) Name() string {
	return "uoj"
}

type Test struct {
	Num    int    `xml:"num,attr"`
	Score  int    `xml:"score,attr"`
	Info   string `xml:"info,attr"`
	Time   int    `xml:"time,attr"`
	Memory int    `xml:"memory,attr"`
	In     string `xml:"in"`
	Out    string `xml:"out"`
	Res    string `xml:"res"`
}

type Details struct {
	Tests []Test `xml:"test"`
	Error string `xml:"error"`
}

type Result struct {
	XMLName xml.Name `xml:"result"`
	Score   int      `xml:"score"`
	Time    int      `xml:"time"`
	Memory  int      `xml:"memory"`
	Error   string   `xml:"error"`
	Details Details  `xml:"details"`
}

func toCodeBlock(v interface{}) string {
	return fmt.Sprintf("```\n%s\n```", v)
}

func ReadResult(resultDir string) (common.SolutionInfo, common.SolutionDetails, error) {
	resultPath := filepath.Join(resultDir, "result.txt")
	// read result
	resultFile, err := os.ReadFile(resultPath)
	if err != nil {
		return common.SolutionInfo{
				Score: 0,
				Metrics: &map[string]float64{
					"cpu": 0,
					"mem": 0,
				},
				Status:  "Unknown Error",
				Message: "An unknown error occurred when reading the result",
			}, common.SolutionDetails{
				Version: 1,
				Jobs:    nil,
				Summary: "Unknown Error",
			}, err
	}

	// unmarshal XML
	var result Result
	if err := xml.Unmarshal(resultFile, &result); err != nil {
		return common.SolutionInfo{
				Score: 0,
				Metrics: &map[string]float64{
					"cpu": 0,
					"mem": 0,
				},
				Status:  "Unknown Error",
				Message: "An unknown error occurred when unmarshaling the result",
			}, common.SolutionDetails{
				Version: 1,
				Jobs:    nil,
				Summary: "Unknown Error",
			}, err
	}

	// Result -> common.SolutionDetails
	var testsResult []*common.SolutionDetailsTest
	status := "Accepted"
	if result.Error != "" {
		status = result.Error
	} else {
		for _, r := range result.Details.Tests {
			if r.Info == "Extra Test Passed" {
				r.Info = "Accepted"
			}
			if status == "Accepted" && r.Info != "Accepted" {
				status = r.Info
			}
			testsResult = append(testsResult, &common.SolutionDetailsTest{
				Name:    "Test " + fmt.Sprint(r.Num),
				Score:   float64(r.Score),
				Status:  r.Info,
				Summary: "Time: `" + fmt.Sprint(r.Time) + "`\tMemory: `" + fmt.Sprint(r.Memory) + "`\n\nInput:\n\n" + toCodeBlock(r.In) + "\n\nOutput:\n\n" + toCodeBlock(r.Out) + "\n\nResult:\n\n" + toCodeBlock(r.Res),
			})
		}
	}
	return common.SolutionInfo{
			Score: float64(result.Score),
			Metrics: &map[string]float64{
				"cpu": float64(result.Time),
				"mem": float64(result.Memory),
			},
			Status:  status,
			Message: "UOJ Judger finished with exit code 0",
		}, common.SolutionDetails{
			Version: 1,
			Jobs: []*common.SolutionDetailsJob{
				{
					Name:       "default",
					Score:      float64(result.Score),
					ScoreScale: 100,
					Status:     status,
					Tests:      testsResult,
					Summary:    "The default subtask",
				},
			},
			Summary: fmt.Sprintf("Error:\n\n%s", toCodeBlock(result.Details.Error)),
		}, nil
}

type UOJAdapterConfig struct {
	SandboxMode string `json:"sandbox_mode"`
}

type SolutionMetadata struct {
	Language string `json:"language"`
}

func (u *UojAdapter) Judge(ctx context.Context, task judge.JudgeTask) error {
	config := task.Config()
	problemData := task.ProblemData()
	solutionData := task.SolutionData()

	adapterConfig := UOJAdapterConfig{
		SandboxMode: "bwrap",
	}
	if err := json.Unmarshal([]byte(config.Judge.Config), &adapterConfig); err != nil {
		return err
	}

	judgerPath := "/opt/uoj_judger"

	// unzip data
	problemDir, err := utils.UnzipTemp(problemData, "problem-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(problemDir)
	solutionDir, err := utils.UnzipTemp(solutionData, "solution-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(solutionDir)
	workDir, err := storage.MkdirTemp("work-uoj-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)
	judgerWorkDir := filepath.Join(workDir, "work")
	if err := os.Mkdir(judgerWorkDir, 0750); err != nil {
		return err
	}
	judgerResultDir := filepath.Join(workDir, "result")
	if err := os.Mkdir(judgerResultDir, 0750); err != nil {
		return err
	}

	language := "C++14"
	if content, err := os.ReadFile(solutionDir + "/.metadata.json"); err == nil {
		var metadata SolutionMetadata
		json.Unmarshal(content, &metadata)
		languages := []string{"C++", "C++11", "C++14", "Python2", "Python3"}
		if slices.Contains(languages, metadata.Language) {
			language = metadata.Language
		}
	}

	if err := os.WriteFile(solutionDir+"/submission.conf", []byte("answer_language "+language), 0666); err != nil {
		return err
	}

	// run judger
	cmd := exec.Command("bwrap",
		"--dir", "/tmp",
		"--dir", "/var",
		"--bind", solutionDir, "/tmp/solution",
		"--ro-bind", problemDir, "/tmp/problem",
		"--ro-bind", judgerPath, "/opt/uoj_judger",
		"--bind", judgerResultDir, "/opt/uoj_judger/result",
		"--bind", judgerWorkDir, "/opt/uoj_judger/work",
		"--ro-bind", "/usr", "/usr",
		"--symlink", "../tmp", "var/tmp",
		"--proc", "/proc",
		"--dev", "/dev",
		"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
		"--symlink", "usr/lib", "/lib",
		"--symlink", "usr/lib64", "/lib64",
		"--symlink", "usr/bin", "/bin",
		"--symlink", "usr/sbin", "/sbin",
		"--chdir", "/opt/uoj_judger",
		"--unshare-all",
		"--die-with-parent",
		"/opt/uoj_judger/main_judger", "/tmp/solution", "/tmp/problem")
	cmd.Dir = judgerPath
	logrus.Infof("Running %s", cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	// read & report result
	result, resultDetails, _ := ReadResult(judgerResultDir)
	task.Update(ctx, &result)
	task.UploadDetails(ctx, &resultDetails)

	return nil
}

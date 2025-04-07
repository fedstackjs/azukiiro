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
	"strconv"
	"strings"

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

func parseProblemConf(problemDir string) (map[string]string, error) {
	problemConfPath := filepath.Join(problemDir, "problem.conf")
	problemConfFile, err := os.ReadFile(problemConfPath)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	lines := strings.Split(string(problemConfFile), "\n")
	for _, line := range lines {
		key, value, found := strings.Cut(line, " ")
		if found {
			result[key] = value
		}
	}
	return result, nil
}

type Test struct {
	Num    int     `xml:"num,attr"`
	Score  float64 `xml:"score,attr"`
	Info   string  `xml:"info,attr"`
	Time   float64 `xml:"time,attr"`
	Memory float64 `xml:"memory,attr"`
	In     string  `xml:"in"`
	Out    string  `xml:"out"`
	Res    string  `xml:"res"`
}

type Subtask struct {
	Num    int     `xml:"num,attr"`
	Score  float64 `xml:"score,attr"`
	Info   string  `xml:"info,attr"`
	Time   float64 `xml:"time,attr"`
	Memory float64 `xml:"memory,attr"`
	Type   string  `xml:"type,attr"`
	Tests  []Test  `xml:"test"`
}

type Details struct {
	Tests    []Test    `xml:"test"`
	Subtasks []Subtask `xml:"subtask"`
	Error    string    `xml:"error"`
}

type Result struct {
	XMLName xml.Name `xml:"result"`
	Score   float64  `xml:"score"`
	Time    float64  `xml:"time"`
	Memory  float64  `xml:"memory"`
	Error   string   `xml:"error"`
	Details Details  `xml:"details"`
}

func toCodeBlock(v interface{}) string {
	return fmt.Sprintf("```\n%s\n```", v)
}

func errorDetails(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Error Details:\n\n%s", toCodeBlock(err))
}

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	str := fmt.Sprintf("%s", err)
	parts := strings.SplitN(str, ":", 2)
	return strings.TrimSpace(parts[0])
}

func testsToJob(subtask Subtask, problemConf map[string]string) (*common.SolutionDetailsJob, error) {
	str, ok := problemConf[fmt.Sprintf("subtask_score_%d", subtask.Num)]
	if !ok {
		return nil, fmt.Errorf("Subtask %v not found in conf", subtask.Num)
	}
	scoreScale, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, err
	}
	result := &common.SolutionDetailsJob{
		Name:       fmt.Sprintf("Subtask %d", subtask.Num),
		Score:      subtask.Score / scoreScale * 100,
		ScoreScale: scoreScale,
		Status:     subtask.Info,
		Tests:      []*common.SolutionDetailsTest{},
		Summary:    "",
	}
	for _, r := range subtask.Tests {
		result.Tests = append(result.Tests, &common.SolutionDetailsTest{
			Name:    "Test " + fmt.Sprint(r.Num),
			Score:   float64(r.Score),
			Status:  r.Info,
			Summary: "Time: `" + fmt.Sprint(r.Time) + "`\tMemory: `" + fmt.Sprint(r.Memory) + "`\n\nInput:\n\n" + toCodeBlock(r.In) + "\n\nOutput:\n\n" + toCodeBlock(r.Out) + "\n\nResult:\n\n" + toCodeBlock(r.Res),
		})
	}
	return result, nil
}

func GenerateErrorResult(err error) (common.SolutionInfo, common.SolutionDetails, error) {
	return common.SolutionInfo{
			Score: 0,
			Metrics: &map[string]float64{
				"cpu": 0,
				"mem": 0,
			},
			Status:  "Judge Error",
			Message: errorMsg(err),
		}, common.SolutionDetails{
			Version: 1,
			Jobs:    nil,
			Summary: errorDetails(err),
		},
		err
}

func ReadResult(resultDir string, problemConf map[string]string) (common.SolutionInfo, common.SolutionDetails, error) {
	resultPath := filepath.Join(resultDir, "result.txt")
	resultFile, err := os.ReadFile(resultPath)
	if err != nil {
		return GenerateErrorResult(fmt.Errorf("failed to read UOJ result"))
	}

	var result Result
	if err := xml.Unmarshal(resultFile, &result); err != nil {
		return GenerateErrorResult(fmt.Errorf("failed to parse UOJ result:\n\n%s", string(resultFile)))
	}

	info := common.SolutionInfo{
		Score: float64(result.Score),
		Metrics: &map[string]float64{
			"cpu": float64(result.Time),
			"mem": float64(result.Memory),
		},
		Status:  "Accepted",
		Message: "UOJ Judger OK",
	}
	details := common.SolutionDetails{
		Version: 1,
		Jobs:    []*common.SolutionDetailsJob{},
		Summary: "",
	}

	// Result -> common.SolutionDetails
	if result.Error != "" {
		info.Status = result.Error
	} else {
		for _, subtask := range result.Details.Subtasks {
			job, err := testsToJob(subtask, problemConf)
			if err != nil {
				return GenerateErrorResult(err)
			}
			details.Jobs = append(details.Jobs, job)
		}
		if len(result.Details.Tests) > 0 {
			job := &common.SolutionDetailsJob{
				Name:       "Default",
				Score:      result.Score,
				ScoreScale: 100,
				Status:     "Accepted",
				Tests:      []*common.SolutionDetailsTest{},
				Summary:    "",
			}
			for _, r := range result.Details.Tests {
				if r.Info == "Extra Test Passed" {
					r.Info = "Accepted"
				}
				if job.Status == "Accepted" && r.Info != "Accepted" {
					job.Status = r.Info
				}
				job.Tests = append(job.Tests, &common.SolutionDetailsTest{
					Name:    "Test " + fmt.Sprint(r.Num),
					Score:   float64(r.Score),
					Status:  r.Info,
					Summary: "Time: `" + fmt.Sprint(r.Time) + "`\tMemory: `" + fmt.Sprint(r.Memory) + "`\n\nInput:\n\n" + toCodeBlock(r.In) + "\n\nOutput:\n\n" + toCodeBlock(r.Out) + "\n\nResult:\n\n" + toCodeBlock(r.Res),
				})
			}
			details.Jobs = append(details.Jobs, job)
		}
		for _, job := range details.Jobs {
			if info.Status == "Accepted" && job.Status != "Accepted" {
				info.Status = job.Status
			}
		}
	}
	return info, details, nil
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
	problemConf, err := parseProblemConf(problemDir)
	if err != nil {
		return err
	}
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
	result, resultDetails, _ := ReadResult(judgerResultDir, problemConf)
	task.Update(ctx, &result)
	task.UploadDetails(ctx, &resultDetails)

	return nil
}

package vjudge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/fedstackjs/azukiiro/utils"
	"github.com/go-resty/resty/v2"
)

func init() {
	judge.RegisterAdapter(&VjudgeAdapter{
		client: resty.New(),
	})
}

type VjSolution struct {
	Memory                    int    `json:"memory"`
	Code                      string `json:"code"`
	StatusType                int    `json:"statusType"`
	Author                    string `json:"author"`
	Length                    int    `json:"length"`
	Runtime                   int    `json:"runtime"`
	Language                  string `json:"language"`
	StatusCanonical           string `json:"statusCanonical"`
	HasSubmissionOriginViewer bool   `json:"hasSubmissionOriginViewer"`
	AuthorId                  int    `json:"authorId"`
	PrismClass                string `json:"prismClass"`
	SubmitTime                int64  `json:"submitTime"`
	IsOpen                    int    `json:"isOpen"`
	Processing                bool   `json:"processing"`
	RunId                     int    `json:"runId"`
	Oj                        string `json:"oj"`
	RemoteRunId               string `json:"remoteRunId"`
	ProbNum                   string `json:"probNum"`
	Status                    string `json:"status"`
	AdditionalInfo            string `json:"additionalInfo"`
}

type VjudgeSolutionMetadata struct {
	Url string `json:"url"`
}

type VjudgeConfig struct {
	VjProblemId string `json:"vjProblemId"`
}

type VjudgeAdapter struct {
	client *resty.Client
}

func (d *VjudgeAdapter) Name() string {
	return "vjudge"
}

func (d *VjudgeAdapter) getSolution(solutionId string, shareCode string) (result VjSolution, err error) {
	_, err = d.client.R().
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,ja;q=0.6").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8").
		SetHeader("Pragma", "no-cache").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		SetFormData(map[string]string{
			"shareCode": shareCode,
		}).
		SetResult(&result).
		SetPathParam("solutionId", solutionId).
		SetQueryParam("inPage", "true").
		Post("https://vjudge.net/solution/data/{solutionId}")
	return
}

func (d *VjudgeAdapter) getUserId(userName string) (string, error) {
	resp, err := d.client.R().
		SetPathParam("userName", userName).
		Get("https://vjudge.net/user/{userName}")
	if err != nil {
		return "", err
	}
	// AOI_User_ID=d9dd05ff-6a8d-4e29-a8e3-1c1844212850
	re := regexp.MustCompile(`AOI_User_ID=([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`)
	matches := re.FindStringSubmatch(resp.String())
	if len(matches) != 2 {
		return "", fmt.Errorf("failed to parse user id")
	}
	return matches[1], nil
}

func (d *VjudgeAdapter) Judge(ctx context.Context, task judge.JudgeTask) error {
	config := task.Config()
	adapterConfig := VjudgeConfig{}
	if err := json.Unmarshal(config.Judge.Config, &adapterConfig); err != nil {
		return err
	}

	solutionDir, err := utils.Unzip(task.SolutionData(), "solution")
	if err != nil {
		return err
	}
	defer os.RemoveAll(solutionDir)

	metadata := VjudgeSolutionMetadata{}
	metadataContent, err := os.ReadFile(solutionDir + "/.metadata.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(metadataContent, &metadata); err != nil {
		return &judge.SimpleSolutionError{
			S: "Bad Solution",
			M: "Failed to parse metadata",
			D: fmt.Sprintf("Failed to parse metadata: %s", err.Error()),
		}
	}

	// url is like https://vjudge.net/solution/17219310/hs5WMYRMvUdCmRV9LccL
	// Parse and get the solution id and share code using regex
	var solutionId, shareCode string
	re := regexp.MustCompile(`https://vjudge.net/solution/(\d+)/(\w+)`)
	matches := re.FindStringSubmatch(metadata.Url)
	if len(matches) == 3 {
		solutionId = matches[1]
		shareCode = matches[2]
	} else {
		return &judge.SimpleSolutionError{
			S: "Bad Solution",
			M: "Failed to parse url",
			D: "Failed to parse vjudge url",
		}
	}

	result, err := d.getSolution(solutionId, shareCode)
	if err != nil {
		return err
	}
	userId, err := d.getUserId(result.Author)
	if err != nil {
		return err
	}
	matchUserId := task.Env()["userId"]
	if userId != matchUserId {
		return &judge.SimpleSolutionError{
			S: "Bad Solution",
			M: "User id mismatch",
			D: fmt.Sprintf("User id mismatch: %s != %s", userId, matchUserId),
		}
	}

	task.Update(ctx, &common.SolutionInfo{
		Score: parseScore(result.AdditionalInfo),
		Metrics: &map[string]float64{
			"cpu": float64(result.Runtime),
			"mem": float64(result.Memory),
		},
		Status:  getMappedStatus(result.Status),
		Message: "Vjudge solution sync ok",
	})
	task.UploadDetails(ctx, &common.SolutionDetails{
		Version: 1,
		Jobs:    []*common.SolutionDetailsJob{},
		Summary: generateVjMd(result),
	})
	return nil
}

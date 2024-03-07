package dummy

import (
	"context"
	"encoding/json"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/judge"
)

func init() {
	judge.RegisterAdapter(&DummyAdapter{})
}

type DummyConfig struct {
	Ping string `json:"ping"`
}

type DummyAdapter struct{}

func (d *DummyAdapter) Name() string {
	return "dummy"
}

func (d *DummyAdapter) Judge(ctx context.Context, task judge.JudgeTask) error {
	config := task.Config()

	adapterConfig := DummyConfig{}
	json.Unmarshal(config.Judge.Config, &adapterConfig)
	task.Update(ctx, &common.SolutionInfo{
		Score: 100,
		Metrics: &map[string]float64{
			"cpu": 0,
			"mem": 0,
		},
		Status:  "AC",
		Message: "Well Done! Accepted",
	})
	task.UploadDetails(ctx, &common.SolutionDetails{
		Version: 1,
		Jobs: []*common.SolutionDetailsJob{
			{
				Name:       "Group 1",
				Score:      100,
				ScoreScale: 100,
				Status:     "AC",
				Tests: []*common.SolutionDetailsTest{
					{
						Name:    "Test 1",
						Score:   100,
						Status:  "AC",
						Summary: "Accepted",
					},
				},
				Summary: "Accepted",
			},
		},
		Summary: "Accepted\nPing is: `" + adapterConfig.Ping + "`",
	})
	return nil
}

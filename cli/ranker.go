package cli

import (
	"context"
	"time"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/db"
	"github.com/fedstackjs/azukiiro/ranker"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func init() {
	commands = append(commands, &rankerCmd{})
}

type rankerCmd struct{}

func (c *rankerCmd) Mount(ctx context.Context, root *cobra.Command) {
	var rankerArgs rankerArgs
	rankerCmd := &cobra.Command{
		Use:   "ranker",
		Short: "Run the ranker",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runRanker(ctx, &rankerArgs),
	}
	rankerCmd.Flags().Float32Var(&rankerArgs.pollInterval, "poll-interval", 1, "Poll interval in seconds")
	root.AddCommand(rankerCmd)
}

type rankerArgs struct {
	pollInterval float32
}

func runRanker(ctx context.Context, rankerArgs *rankerArgs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logrus.Println("Starting ranker")
		ctx, cleanup := db.WithMongo(ctx)
		defer cleanup()

		client.InitFromConfig()
		for {
			cont, err := ranker.Poll(ctx)
			if err != nil {
				logrus.Println("Error:", err)
			}
			waitDur := time.Duration(0)
			if !cont {
				waitDur = time.Duration(rankerArgs.pollInterval) * time.Second
			}
			timer := time.NewTimer(waitDur)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return nil
			case <-timer.C:
			}
		}
	}
}

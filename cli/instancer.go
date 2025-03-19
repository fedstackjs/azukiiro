package cli

import (
	"context"
	"time"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/instancer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	commands = append(commands, &instancerCommand{})
}

type instancerCommand struct{}

func (c *instancerCommand) Mount(ctx context.Context, root *cobra.Command) {
	var instancerArgs instancerArgs
	instancerCmd := &cobra.Command{
		Use:   "instancer",
		Short: "Run the instancer",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runInstancer(ctx, &instancerArgs),
	}
	instancerCmd.Flags().Float32Var(&instancerArgs.pollInterval, "poll-interval", 1, "Poll interval in seconds")
	root.AddCommand(instancerCmd)
}

type instancerArgs struct {
	pollInterval float32
}

func runInstancer(ctx context.Context, instancerArgs *instancerArgs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logrus.Println("Starting instancer")
		client.InitFromConfig()
		logrus.Infoln("Serial poller started")
		for {
			cont, err := instancer.Poll(ctx)
			if err != nil {
				logrus.Println("Error:", err)
			}
			waitDur := time.Duration(0)
			if !cont {
				waitDur = time.Duration(instancerArgs.pollInterval) * time.Second
			}
			timer := time.NewTimer(waitDur)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				logrus.Infoln("Serial poller stopped")
				return nil
			case <-timer.C:
			}
		}
	}
}

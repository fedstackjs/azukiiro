package cli

import (
	"context"
	"time"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type daemonArgs struct {
	pollInterval float32
}

func runDaemon(ctx context.Context, daemonArgs *daemonArgs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logrus.Println("Starting daemon")
		client.InitFromConfig()
		for {
			cont, err := judge.Poll(ctx)
			if err != nil {
				logrus.Println("Error:", err)
			}
			waitDur := time.Duration(0)
			if !cont {
				waitDur = time.Duration(daemonArgs.pollInterval) * time.Second
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

package cli

import (
	"context"
	"sync"
	"time"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	commands = append(commands, &daemonCommand{})
}

type daemonCommand struct{}

func (c *daemonCommand) Mount(ctx context.Context, root *cobra.Command) {
	var daemonArgs daemonArgs
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run the daemon",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runDaemon(ctx, &daemonArgs),
	}
	daemonCmd.Flags().IntVar(&daemonArgs.concurrency, "concurrency", 1, "Concurrency")
	daemonCmd.Flags().Float32Var(&daemonArgs.pollInterval, "poll-interval", 1, "Poll interval in seconds")
	root.AddCommand(daemonCmd)
}

type daemonArgs struct {
	concurrency  int
	pollInterval float32
}

func runDaemon(ctx context.Context, daemonArgs *daemonArgs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logrus.Println("Starting daemon")
		client.InitFromConfig()
		if daemonArgs.concurrency == 1 {
			logrus.Infoln("Serial poller started")
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
					logrus.Infoln("Serial poller stopped")
					return nil
				case <-timer.C:
				}
			}
		} else {
			wg := sync.WaitGroup{}
			queue := make(chan *judge.RemoteJudgeTask)
			wg.Add(1)
			go func() {
				judge.ParallelPoller(ctx, daemonArgs.pollInterval, queue)
				wg.Done()
			}()
			for i := 0; i < daemonArgs.concurrency; i++ {
				wg.Add(1)
				go func() {
					judge.ParallelJudger(ctx, queue)
					wg.Done()
				}()
			}
			wg.Wait()
			return nil
		}
	}
}

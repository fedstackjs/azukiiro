package cli

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fedstackjs/azukiiro/client"
	"github.com/fedstackjs/azukiiro/common"
)

func init() {
	commands = append(commands, &registerCmd{})
}

type registerCmd struct{}

func (c *registerCmd) Mount(ctx context.Context, root *cobra.Command) {
	var regArgs registerArgs
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Register a runner to the server",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runRegister(ctx, &regArgs), // must use a pointer to regArgs
	}
	registerCmd.Flags().StringVar(&regArgs.ServerAddr, "server", "", "AOI Server address")
	registerCmd.Flags().BoolVar(&regArgs.Force, "force", false, "Force register")
	registerCmd.Flags().StringVar(&regArgs.Token, "token", "", "Runner token")
	registerCmd.Flags().StringVar(&regArgs.Name, "name", "", "Runner name")
	registerCmd.Flags().StringVar(&regArgs.Labels, "labels", "", "Runner tags, comma separated")
	registerCmd.MarkFlagRequired("server")
	registerCmd.MarkFlagRequired("token")
	root.AddCommand(registerCmd)
}

type registerArgs struct {
	ServerAddr string
	Force      bool
	Name       string
	Labels     string
	Token      string
}

func splitLabels(input string) []string {
	return strings.Split(input, ",")
}

func runRegister(ctx context.Context, regArgs *registerArgs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logrus.Println("Registering runner")
		logrus.Println("ServerAddr:", regArgs.ServerAddr)

		runnerKey := viper.GetString("runnerKey")
		if runnerKey != "" && !regArgs.Force {
			logrus.Println("Runner already registered, exiting...")
			return nil
		}

		viper.Set("serverAddr", regArgs.ServerAddr)

		http := client.GetDefaultHTTPClient()
		http.SetBaseURL(regArgs.ServerAddr)

		if regArgs.Name == "" {
			name, err := os.Hostname()
			if err != nil {
				logrus.Fatalln(err)
			}
			regArgs.Name = name
		}

		if regArgs.Labels == "" {
			regArgs.Labels = "default"
		}

		req := &client.RegisterRequest{
			Name:              regArgs.Name,
			Version:           common.GetVersion(),
			Labels:            splitLabels(regArgs.Labels),
			RegistrationToken: regArgs.Token,
		}

		res, err := client.Register(ctx, req)

		if err != nil {
			logrus.Fatalln(err)
		}

		logrus.Println("RunnerId:", res.RunnerId)
		viper.Set("runnerId", res.RunnerId)
		viper.Set("runnerKey", res.RunnerKey)
		viper.WriteConfig()

		logrus.Println("Runner registered successfully")

		return nil
	}
}

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fedstackjs/azukiiro/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute(ctx context.Context) {
	configFile := ""

	cobra.OnInitialize(func() {
		if configFile != "" {
			viper.SetConfigFile(configFile)
		} else {
			viper.AddConfigPath("/etc/azukiiro/")
			viper.AddConfigPath(".")
			viper.SetConfigName("config")
		}

		viper.SetDefault("storagePath", "/var/lib/azukiiro")

		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}

		storage.Initialize()
	})

	// ./azukiiro
	rootCmd := &cobra.Command{
		Use:   "azukiiro [command]",
		Short: "Judge runner for the AOI Project",
		Args:  cobra.MaximumNArgs(1),
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path")

	// ./azukiiro register
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
	rootCmd.AddCommand(registerCmd)

	// ./azukiiro daemon
	var daemonArgs daemonArgs
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run the daemon",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runDaemon(ctx, &daemonArgs),
	}
	daemonCmd.Flags().Float32Var(&daemonArgs.pollInterval, "poll-interval", 1, "Poll interval in seconds")
	rootCmd.AddCommand(daemonCmd)

	// ./azukiiro ranker
	var rankerArgs rankerArgs
	rankerCmd := &cobra.Command{
		Use:   "ranker",
		Short: "Run the ranker",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runRanker(ctx, &rankerArgs),
	}
	rankerCmd.Flags().Float32Var(&rankerArgs.pollInterval, "poll-interval", 1, "Poll interval in seconds")
	rootCmd.AddCommand(rankerCmd)

	var judgeArgs judgeArgs
	judgeCmd := &cobra.Command{
		Use:   "judge",
		Short: "Run the judge locally",
		Args:  cobra.MaximumNArgs(0),
		RunE:  runJudge(ctx, &judgeArgs),
	}
	judgeCmd.Flags().StringVar(&judgeArgs.problemConfig, "problem-config", "", "Problem config file")
	judgeCmd.MarkFlagRequired("problem-config")
	judgeCmd.Flags().StringVar(&judgeArgs.problemData, "problem-data", "", "Problem data file")
	judgeCmd.MarkFlagRequired("problem-data")
	judgeCmd.Flags().StringVar(&judgeArgs.solutionData, "solution-data", "", "Solution data file")
	judgeCmd.MarkFlagRequired("solution-data")
	rootCmd.AddCommand(judgeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

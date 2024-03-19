package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fedstackjs/azukiiro/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Command interface {
	Mount(ctx context.Context, root *cobra.Command)
}

var commands = []Command{}

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
		Short: "Official runner for the AOI Project",
		Args:  cobra.MaximumNArgs(1),
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path")

	for _, c := range commands {
		c.Mount(ctx, rootCmd)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

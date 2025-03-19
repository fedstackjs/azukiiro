package cli

import (
	"context"

	"github.com/fedstackjs/azukiiro/common"
	"github.com/fedstackjs/azukiiro/instancer"
	"github.com/fedstackjs/azukiiro/judge"
	"github.com/fedstackjs/azukiiro/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	commands = append(commands, &infoCmd{})
}

type infoCmd struct{}

func (c *infoCmd) Mount(ctx context.Context, root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show azukiiro info",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infoln("Version           :", common.GetVersion())
			logrus.Infoln("Judge adapters    :", judge.GetAdapterNames())
			logrus.Infoln("Instance adapters :", instancer.GetAdapterNames())
			logrus.Infoln("Config path       :", viper.ConfigFileUsed())
			logrus.Infoln("Storage path      :", storage.GetRootPath())
			logrus.Infoln("Server Address    :", viper.GetString("serverAddr"))
			logrus.Infoln("Runner ID         :", viper.GetString("runnerId"))
			return nil
		},
	}
	root.AddCommand(cmd)
}

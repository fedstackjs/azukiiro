package main

import (
	"context"
	"fmt"
	"os/signal"
	"runtime/debug"
	"syscall"

	_ "github.com/fedstackjs/azukiiro/adapters"
	"github.com/fedstackjs/azukiiro/cli"
	"github.com/fedstackjs/azukiiro/common"
)

var (
	version string
)

func init() {
	if version != "" {
		common.Version = fmt.Sprintf("azukiiro/%s", version)
		if info, ok := debug.ReadBuildInfo(); ok {
			version = " ("
			version += info.Main.Path
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					version += "; " + setting.Value[:7]
				case "vcs.modified":
					if setting.Value == "true" {
						version += "; modified"
					}
				}
			}
			version += ")"
			common.Version += version
		} else {
			common.Version += " (unknown)"
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cli.Execute(ctx)
}

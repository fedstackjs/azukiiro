package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/fedstackjs/azukiiro/adapters"
	"github.com/fedstackjs/azukiiro/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cli.Execute(ctx)
}

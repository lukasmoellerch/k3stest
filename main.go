package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/lukasmoellerch/k3stest/pkg/k3t"
)

func main() {
	bg := context.Background()
	ctx, stop := signal.NotifyContext(bg, os.Interrupt)
	defer stop()

	cluster := k3t.NewClusterFromEnv(7443)
	data, err := cluster.Start(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", data)

	<-ctx.Done()
	cluster.Logger.Info().
		Msg("shutting down")
	if err := cluster.Stop(bg); err != nil {
		panic(err)
	}
}

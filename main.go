package main

import (
	"context"
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/worlpaker/gitty/cmd"
)

//go:embed VERSION
var version string

// run executes the program and handles graceful shutdown.
func run() int {
	// Graceful shutdown for Ctrl+C.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c: // First signal, soft exit.
			cancel()
		case <-ctx.Done():
		}
		<-c // Second signal, hard exit.
		os.Exit(1)
	}()

	// Execute the program.
	if err := cmd.Execute(ctx, version); err != nil {
		return 1
	}

	return 0
}

func main() {
	os.Exit(run())
}

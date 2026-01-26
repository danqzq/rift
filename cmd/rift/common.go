package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// setupContext creates a context with cancellation and sets up signal handling
// for SIGINT and SIGTERM. Returns the context and cancel function.
func setupContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}

// arrayFlags allows multiple flag values to be collected.
type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

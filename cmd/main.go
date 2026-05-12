package main

import (
	"context"
	"os"

	"github.com/gi8lino/loggo/internal/app"
)

var (
	// Version is the build version set via ldflags.
	Version string = "dev"
	// Commit is the build commit set via ldflags.
	Commit string = "none"
)

// main sets up the application context and runs the main loop.
func main() {
	ctx := context.Background()

	if err := app.Run(ctx, Version, Commit, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv); err != nil {
		os.Exit(1)
	}
}

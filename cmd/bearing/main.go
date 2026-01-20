package main

import (
	"os"

	"github.com/joshribakoff/bearing/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"github.com/hchenc/iceberg/cmd/controller-manager/app"
	"os"
)

func main() {
	cmd := app.NewControllerManagerCommandOptions()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

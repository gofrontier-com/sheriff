package main

import (
	"os"

	"github.com/frontierdigital/sheriff/pkg/cmd/sheriff"
	"github.com/frontierdigital/utils/output"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	command := sheriff.NewRootCmd(version, commit, date)
	if err := command.Execute(); err != nil {
		output.PrintlnError(err)
		os.Exit(1)
	}
}

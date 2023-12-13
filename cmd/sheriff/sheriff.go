package main

import (
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/frontierdigital/sheriff/pkg/cmd/sheriff"
	"github.com/frontierdigital/utils/output"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	myFigure := figure.NewFigure("Sheriff for Azure", "doom", true)
	myFigure.Print()
	output.PrintlnInfo()
	command := sheriff.NewRootCmd(version, commit, date)
	if err := command.Execute(); err != nil {
		output.PrintlnError(err)
		os.Exit(1)
	}
}

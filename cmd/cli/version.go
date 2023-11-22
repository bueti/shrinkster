package main

import (
	"fmt"

	"github.com/bueti/shrinkster/internal/config"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func (app *application) version(context *cli.Context) {
	fmt.Printf("%s %s, commit %s, built at %s\n", config.AppName, version, commit, date)
}

package main

import (
	"fmt"

	"github.com/bueti/shrinkster/internal/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func (app *application) version() {
	fmt.Printf("%s %s, commit %s, built at %s\n", config.AppName, version, commit, date)
}

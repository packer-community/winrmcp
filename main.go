package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
)

func main() {
	ui := &cli.BasicUi{Writer: os.Stdout}
	commands := map[string]cli.CommandFactory{
		"ls": func() (cli.Command, error) {
			return &lsCommand{
				ui: ui,
			}, nil
		},

		"cp": func() (cli.Command, error) {
			return &cpCommand{
				ui: ui,
			}, nil
		},

		"info": func() (cli.Command, error) {
			return &infoCommand{
				ui: ui,
			}, nil
		},
	}

	app := &cli.CLI{
		Args:     os.Args[1:],
		Commands: commands,
		HelpFunc: cli.BasicHelpFunc("winrmfs"),
	}

	status, err := app.Run()
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	os.Exit(status)
}

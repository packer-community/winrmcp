package main

import (
	"flag"
	"strings"

	"github.com/dylanmei/winrmcp/winrmcp"
	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/cli"
)

type cpCommand struct {
	shutdown <-chan struct{}
	ui       cli.Ui
}

func (c *cpCommand) Help() string {
	text := `
Usage: winrmfs cp [options] from to

  Copies a local file to a remote directory.

Options:

  -addr=127.0.0.1:5985    Host and port of the remote machine
  -user=""                Name of the user to authenticate as
  -pass=""                Password to authenticate with
`
	return strings.TrimSpace(text)
}

func (c *cpCommand) Synopsis() string {
	return "Copies a local file to a remote directory"
}

func (c *cpCommand) Run(args []string) int {
	var user string
	var pass string

	flags := flag.NewFlagSet("ls", flag.ContinueOnError)
	flags.Usage = func() { c.ui.Output(c.Help()) }
	flags.StringVar(&user, "user", "", "auth name")
	flags.StringVar(&pass, "pass", "", "auth password")
	addr := addrFlag(flags)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) < 1 {
		c.ui.Error("A source directory is required.\n")
		c.ui.Error(c.Help())
		return 1
	}
	if len(args) < 2 {
		c.ui.Error("A remote directory is required.\n")
		c.ui.Error(c.Help())
		return 1
	}
	if len(args) > 2 {
		c.ui.Error("Too many arguments. Only a source and remote directory are required.\n")
		c.ui.Error(c.Help())
		return 1
	}

	endpoint, err := parseEndpoint(*addr)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	sourcePath := args[0]
	remotePath := args[1]

	client := winrm.NewClient(endpoint, user, pass)
	cp := winrmcp.New(client)

	err = cp.Copy(sourcePath, remotePath)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	return 0
}

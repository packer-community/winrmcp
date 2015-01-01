package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/dylanmei/winrmfs/winrmfs"
	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/cli"
)

type infoCommand struct {
	shutdown <-chan struct{}
	ui       cli.Ui
}

func (c *infoCommand) Help() string {
	text := `
Usage: winrmfs info [options]

  Show status and info about the remote.

Options:

  -addr=127.0.0.1:5985    Host and port of the remote machine
  -user=""                Name of the user to authenticate as
  -pass=""                Password to authenticate with
`
	return strings.TrimSpace(text)
}

func (c *infoCommand) Synopsis() string {
	return "Show status and info about the remote"
}

func (c *infoCommand) Run(args []string) int {
	var user string
	var pass string

	flags := flag.NewFlagSet("info", flag.ContinueOnError)
	flags.Usage = func() { c.ui.Output(c.Help()) }
	flags.StringVar(&user, "user", "", "auth name")
	flags.StringVar(&pass, "pass", "", "auth password")
	addr := addrFlag(flags)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	endpoint, err := parseEndpoint(*addr)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	client := winrm.NewClient(endpoint, user, pass)
	fs := winrmfs.New(client)

	info, err := fs.Info()
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	c.ui.Output("Auth")
	c.ui.Output(fmt.Sprintf("\tUser: %s", user))
	c.ui.Output(fmt.Sprintf("\tBasic: %v", true))
	c.ui.Output("WinRM Config")
	c.ui.Output(fmt.Sprintf("\t%s: %d", "MaxEnvelopeSizeKB", info.WinRM.MaxEnvelopeSizekb))
	c.ui.Output(fmt.Sprintf("\t%s: %d", "Service/MaxConcurrentOperationsPerUser", info.WinRM.Service.MaxConcurrentOperationsPerUser))
	c.ui.Output(fmt.Sprintf("\t%s: %d", "Winrs/MaxMemoryPerShellMB", info.WinRM.Winrs.MaxMemoryPerShellMB))

	return 0
}

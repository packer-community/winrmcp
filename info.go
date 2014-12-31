package main

import (
	"bytes"
	"flag"
	"fmt"
	"strings"

	"github.com/masterzen/winrm/winrm"
	"github.com/masterzen/xmlpath"
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
	stdout, stderr, err := client.RunWithString("winrm get winrm/config -format:xml", "")

	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	if stdout != "" {
		node, err := xmlpath.Parse(bytes.NewBuffer([]byte(stdout)))
		if err != nil {
			c.ui.Error(err.Error())
			return 1
		}

		c.ui.Output(fmt.Sprintf("%s: %s", "MaxEnvelopeSizekb", parseConfig(node, "/cfg:Config/cfg:MaxEnvelopeSizekb")))
		c.ui.Output(fmt.Sprintf("%s: %s", "MaxTimeoutms", parseConfig(node, "/cfg:Config/cfg:MaxTimeoutms")))
		c.ui.Output(fmt.Sprintf("%s: %s", "MaxBatchItems", parseConfig(node, "/cfg:Config/cfg:MaxBatchItems")))
		c.ui.Output(fmt.Sprintf("%s: %s", "Service/MaxConcurrentOperations", parseConfig(node, "/cfg:Config/cfg:Service/cfg:MaxConcurrentOperations")))
		c.ui.Output(fmt.Sprintf("%s: %s", "Service/MaxConcurrentOperationsPerUser", parseConfig(node, "/cfg:Config/cfg:Service/cfg:MaxConcurrentOperationsPerUser")))
		c.ui.Output(fmt.Sprintf("%s: %s", "Winrs/IdleTimeout", parseConfig(node, "/cfg:Config/cfg:Winrs/cfg:IdleTimeout")))
		c.ui.Output(fmt.Sprintf("%s: %s", "Winrs/MaxConcurrentUsers", parseConfig(node, "/cfg:Config/cfg:Winrs/cfg:MaxConcurrentUsers")))
	}

	if stderr != "" {
		println(stderr)
	}

	return 0
}

func parseConfig(config *xmlpath.Node, selector string) string {
	path, _ := xmlpath.CompileWithNamespace(selector, []xmlpath.Namespace{
		{"cfg", "http://schemas.microsoft.com/wbem/wsman/1/config"},
	})

	value, ok := path.String(config)
	if !ok {
		return ""
	}

	return value
}

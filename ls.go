package main

import (
	"flag"
	"io"
	"os"
	"strings"

	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/cli"
)

type lsCommand struct {
	shutdown <-chan struct{}
	ui       cli.Ui
}

func (c *lsCommand) Help() string {
	text := `
Usage: winrmfs ls [options] directory

  List files of a remote directory.

Options:

  -addr=127.0.0.1:5985    Host and port of the remote machine
  -user=""                Name of the user to authenticate as
  -pass=""                Password to authenticate with
`
	return strings.TrimSpace(text)
}

func (c *lsCommand) Synopsis() string {
	return "List files of a remote directory"
}

func (c *lsCommand) Run(args []string) int {
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
		c.ui.Error("A directory is required.\n")
		c.ui.Error(c.Help())
		return 1
	}
	if len(args) > 1 {
		c.ui.Error("Too many arguments. Only a directory is required.\n")
		c.ui.Error(c.Help())
		return 1
	}

	dir := args[0]
	endpoint, err := parseEndpoint(*addr)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	client := winrm.NewClient(endpoint, user, pass)
	shell, err := client.CreateShell()

	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	defer shell.Close()
	cmd, err := shell.Execute("powershell", "Get-ChildItem", friendlyPath(dir))
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	go io.Copy(os.Stdout, cmd.Stdout)
	go io.Copy(os.Stderr, cmd.Stderr)
	cmd.Wait()
	cmd.Close()

	return cmd.ExitCode()
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/packer-community/winrmcp/winrmcp"
)

var usage string = `
Usage: winrmcp [options] [-help | <from> <to>]

  Copy a local file or directory to a remote directory.

Options:

  -addr=localhost:5985  Host and port of the remote machine
  -user=""              Name of the user to authenticate as
  -pass=""              Password to authenticate with

`

func main() {
	if hasSwitch("-help") {
		fmt.Print(usage)
		os.Exit(0)
	}
	if err := runMain(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func runMain() error {
	flags := flag.NewFlagSet("cli", flag.ContinueOnError)
	flags.Usage = func() { fmt.Print(usage) }
	user := flags.String("user", "vagrant", "winrm admin username")
	pass := flags.String("pass", "vagrant", "winrm admin password")
	addr := flags.String("addr", "localhost:5985", "winrm remote host:port")
	flags.Parse(os.Args[1:])

	client, err := winrmcp.New(*addr, &winrmcp.Config{
		Auth: winrmcp.Auth{*user, *pass},
	})
	if err != nil {
		return err
	}

	args := flags.Args()
	if len(args) < 1 {
		return errors.New("Source directory is required.")
	}
	if len(args) < 2 {
		return errors.New("Remote directory is required.")
	}
	if len(args) > 2 {
		return errors.New("Too many arguments.")
	}

	return client.Copy(args[0], args[1])
}

func hasSwitch(name string) bool {
	for _, arg := range os.Args[1:] {
		if arg == name {
			return true
		}
	}
	return false
}

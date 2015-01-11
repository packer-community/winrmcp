package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/masterzen/winrm/winrm"
	"github.com/packer-community/winrmcp/winrmcp"
)

var usage string = `
Usage: winrmcp [options] [-help | -info | <from> <to>]

  Copy a local file or directory to a remote directory.

Options:

  -addr=localhost:5985  Host and port of the remote machine
  -user=""              Name of the user to authenticate as
  -pass=""              Password to authenticate with

`

func main() {
	if hasSwitch("-help") {
		fmt.Print(usage)
		return
	}

	flags := flag.NewFlagSet("cli", flag.ContinueOnError)
	flags.Usage = func() { fmt.Print(usage) }
	info := flags.Bool("info", false, "")
	user := flags.String("user", "vagrant", "winrm admin username")
	pass := flags.String("pass", "vagrant", "winrm admin password")
	addr := addrFlag(flags)
	flags.Parse(os.Args[1:])

	endpoint, err := parseEndpoint(*addr)
	if err != nil {
		fmt.Printf("Couldn't parse addr: %v\n", err)
		os.Exit(1)
	}

	client := winrm.NewClient(endpoint, *user, *pass)
	var exitCode int
	if *info {
		exitCode := runInfo(client, *user, *addr)
		os.Exit(exitCode)
	}

	args := flags.Args()
	if len(args) < 1 {
		fmt.Println("Source directory is required.")
		exitCode = 1
	} else if len(args) < 2 {
		fmt.Println("Remote directory is required.\n")
		exitCode = 1
	} else if len(args) > 2 {
		fmt.Println("Too many arguments. Only a source and remote directory are required.\n")
		exitCode = 1
	}
	if exitCode != 0 {
		fmt.Print(usage)
	} else {
		exitCode = runCopy(client, args[0], args[1])
	}

	os.Exit(exitCode)
}

func hasSwitch(name string) bool {
	for _, arg := range os.Args[1:] {
		if arg == name {
			return true
		}
	}
	return false
}

func runCopy(client *winrm.Client, fromPath, toPath string) int {
	cp := winrmcp.New(client)

	err := cp.Copy(fromPath, toPath)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	return 0
}

func runInfo(client *winrm.Client, user, addr string) int {
	cp := winrmcp.New(client)

	info, err := cp.Info()
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	fmt.Println("Client")
	fmt.Printf("    Addr: %s\n", addr)
	fmt.Printf("    Auth: %s", "Basic\n")
	fmt.Printf("    User: %s\n", user)
	fmt.Println("WinRM Config")
	fmt.Printf("    %s: %d\n", "MaxEnvelopeSizeKB", info.WinRM.MaxEnvelopeSizeKB)
	fmt.Printf("    %s: %d\n", "MaxTimeoutMS", info.WinRM.MaxTimeoutMS)
	fmt.Printf("    %s: %d\n", "Service/MaxConcurrentOperations", info.WinRM.Service.MaxConcurrentOperations)
	fmt.Printf("    %s: %d\n", "Service/MaxConcurrentOperationsPerUser", info.WinRM.Service.MaxConcurrentOperationsPerUser)
	fmt.Printf("    %s: %d\n", "Service/MaxConnections", info.WinRM.Service.MaxConnections)
	fmt.Printf("    %s: %d\n", "Winrs/MaxConcurrentUsers", info.WinRM.Winrs.MaxConcurrentUsers)
	fmt.Printf("    %s: %d\n", "Winrs/MaxProcessesPerShell", info.WinRM.Winrs.MaxProcessesPerShell)
	fmt.Printf("    %s: %d\n", "Winrs/MaxMemoryPerShellMB", info.WinRM.Winrs.MaxMemoryPerShellMB)
	fmt.Printf("    %s: %d\n", "Winrs/MaxShellsPerUser", info.WinRM.Winrs.MaxShellsPerUser)
	return 0
}

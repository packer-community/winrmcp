package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dylanmei/winrmcp/winrmcp"
	"github.com/masterzen/winrm/winrm"
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
	if runSwitch("-help") {
		return
	}
	if runSwitch("-info") {
		return
	}

	exitCode := runMain()
	os.Exit(exitCode)
}

func runSwitch(name string) bool {
	for _, arg := range os.Args[1:] {
		if arg != name {
			continue
		}
		switch arg {
		case "-help":
			fmt.Print(usage)
			return true
		case "-info":
			exitCode := runInfo()
			os.Exit(exitCode)
			return true
		}
	}

	return false
}

func runMain() int {
	var user string
	var pass string

	flags := flag.NewFlagSet("cli", flag.ContinueOnError)
	flags.Usage = func() { fmt.Print(usage) }
	flags.StringVar(&user, "user", "vagrant", "winrm admin username")
	flags.StringVar(&pass, "pass", "vagrant", "winrm admin password")
	addr := addrFlag(flags)
	flags.Parse(os.Args[1:])

	args := flags.Args()
	if len(args) < 1 {
		fmt.Println("Source directory is required.")
		fmt.Print(usage)
		return 1
	}
	if len(args) < 2 {
		fmt.Println("Remote directory is required.\n")
		fmt.Print(usage)
		return 1
	}
	if len(args) > 2 {
		fmt.Println("Too many arguments. Only a source and remote directory are required.\n")
		fmt.Print(usage)
		return 1
	}

	endpoint, err := parseEndpoint(*addr)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	fromPath := args[0]
	toPath := args[1]

	client := winrm.NewClient(endpoint, user, pass)
	cp := winrmcp.New(client)

	err = cp.Copy(fromPath, toPath)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	return 0
}

func runInfo() int {
	var user string
	var pass string

	flags := flag.NewFlagSet("cli", flag.ContinueOnError)
	flags.Usage = func() { fmt.Print(usage) }
	flags.StringVar(&user, "user", "vagrant", "winrm admin username")
	flags.StringVar(&pass, "pass", "vagrant", "winrm admin password")
	addr := addrFlag(flags)
	flags.Parse(os.Args)

	endpoint, err := parseEndpoint(*addr)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	client := winrm.NewClient(endpoint, user, pass)
	cp := winrmcp.New(client)

	info, err := cp.Info()
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	fmt.Println("Client")
	fmt.Printf("    Addr: %s:%d\n", endpoint.Host, endpoint.Port)
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

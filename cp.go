package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/packer/common/uuid"
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
	tempFile := fmt.Sprintf("winrmfs-%s.tmp", uuid.TimeOrderedUUID())
	tempPath := "$env:TEMP\\" + tempFile

	file, err := os.Open(sourcePath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error opening file %s: %v", sourcePath, err))
		return 1
	}

	defer file.Close()
	client := winrm.NewClient(endpoint, user, pass)

	shell, err := client.CreateShell()
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	defer shell.Close()

	c.ui.Info(fmt.Sprintf("Copying file from %s to %s", sourcePath, tempPath))
	err = uploadContent(shell, "%TEMP%\\"+tempFile, file)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error uploading file from %s to %s: %v", sourcePath, tempPath, err))
		return 1
	}

	c.ui.Info(fmt.Sprintf("Moving file from %s to %s", tempPath, remotePath))
	err = restoreContent(client, tempPath, friendlyPath(remotePath))
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error restoring file from %s to %s: %v", tempPath, remotePath, err))
		return 1
	}

	err = cleanupContent(client, tempPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error deleting temporary file %s: %v", tempPath, err))
		return 1
	}

	return 0
}

func uploadContent(shell *winrm.Shell, filePath string, reader io.Reader) error {
	// Upload the file in chunks to get around the Windows command line size limit.
	//   Base64 encodes each set of three bytes into four bytes.
	//   In addition the output is padded to always be a multiple of four.
	//
	//   ceil(n / 3) * 4 = m1 - m2
	//
	//   where:
	//     n  = bytes
	//     m1 = max (cmd.exe has a 8192 character command line limit. powershell?)
	//     m2 = len(filePath)

	chunkSize := ((8000 - len(filePath)) / 4) * 3
	chunk := make([]byte, chunkSize)

	for {
		n, err := reader.Read(chunk)

		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			return nil
		}

		content := base64.StdEncoding.EncodeToString(chunk[:n])
		if err = appendContent(shell, filePath, content); err != nil {
			return err
		}
	}

	return nil
}

func restoreContent(client *winrm.Client, fromPath, toPath string) error {
	shell, err := client.CreateShell()
	if err != nil {
		return err
	}

	defer shell.Close()
	script := fmt.Sprintf(`
		$tmp_file_path = [System.IO.Path]::GetFullPath("%s")
		$dest_file_path = [System.IO.Path]::GetFullPath("%s")
		if (Test-Path $dest_file_path) {
			rm $dest_file_path
		}
		else {
			$dest_dir = ([System.IO.Path]::GetDirectoryName($dest_file_path))
			New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path $dest_dir | Out-Null
		}

		if (Test-Path $tmp_file_path) {
			$base64_lines = Get-Content $tmp_file_path
			$base64_string = [string]::join("",$base64_lines)
			$bytes = [System.Convert]::FromBase64String($base64_string) 
			[System.IO.File]::WriteAllBytes($dest_file_path, $bytes)
		} else {
			echo $null > $dest_file_path
		}
	`, fromPath, toPath)

	cmd, err := shell.Execute(winrm.Powershell(script))
	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, cmd.Stdout)
	go io.Copy(os.Stderr, cmd.Stderr)

	cmd.Wait()
	cmd.Close()

	if cmd.ExitCode() != 0 {
		return errors.New(fmt.Sprintf("restore operation returned code=%d", cmd.ExitCode()))
	}
	return nil
}

func cleanupContent(client *winrm.Client, filePath string) error {
	shell, err := client.CreateShell()
	if err != nil {
		return err
	}

	defer shell.Close()
	cmd, _ := shell.Execute("powershell", "Remove-Item", filePath, "-ErrorAction SilentlyContinue")

	cmd.Wait()
	cmd.Close()
	return nil
}

func appendContent(shell *winrm.Shell, filePath, content string) error {
	//cmd, err := shell.Execute("powershell", "Add-Content", friendlyPath(filePath), "-value", content)
	cmd, err := shell.Execute(fmt.Sprintf("echo %s >> \"%s\"", content, filePath))

	if err != nil {
		return err
	}

	defer cmd.Close()
	go io.Copy(os.Stdout, cmd.Stdout)
	go io.Copy(os.Stderr, cmd.Stderr)
	cmd.Wait()

	if cmd.ExitCode() != 0 {
		return errors.New(fmt.Sprintf("upload operation returned code=%d", cmd.ExitCode()))
	}

	return nil
}

package winrmfs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/packer/common/uuid"
)

func doCopy(client *winrm.Client, fromPath, toPath string) error {
	tempFile := fmt.Sprintf("winrmfs-%s.tmp", uuid.TimeOrderedUUID())
	tempPath := "$env:TEMP\\" + tempFile

	file, err := os.Open(fromPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't open file %s: %v", fromPath, err))
	}

	defer file.Close()
	shell, err := client.CreateShell()
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't create shell: %v", fromPath, err))
	}
	defer shell.Close()

	if os.Getenv("WINRMFS_DEBUG") != "" {
		log.Printf("Copying file from %s to %s\n", fromPath, tempPath)
	}

	err = uploadContent(shell, "%TEMP%\\"+tempFile, file)
	if err != nil {
		return errors.New(fmt.Sprintf("Error uploading file from %s to %s: %v", fromPath, tempPath, err))
	}

	if os.Getenv("WINRMFS_DEBUG") != "" {
		log.Printf("Moving file from %s to %s", tempPath, toPath)
	}

	err = restoreContent(client, tempPath, toPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error restoring file from %s to %s: %v", tempPath, toPath, err))
	}

	if os.Getenv("WINRMFS_DEBUG") != "" {
		log.Printf("Removing temporary file %s", tempPath)
	}

	err = cleanupContent(client, tempPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error removing temporary file %s: %v", tempPath, err))
	}

	return nil
}

func uploadContent(shell *winrm.Shell, filePath string, reader io.Reader) error {
	// Upload the file in chunks to get around the Windows command line size limit.
	// Base64 encodes each set of three bytes into four bytes. In addition the output
	// is padded to always be a multiple of four.
	//
	//   ceil(n / 3) * 4 = m1 - m2
	//
	//   where:
	//     n  = bytes
	//     m1 = max (8192 character command limit.)
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

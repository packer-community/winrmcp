package winrmcp

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/packer/common/uuid"
)

func doCopy(client *winrm.Client, config *Config, in io.Reader, toPath string) error {
	tempFile := fmt.Sprintf("winrmcp-%s.tmp", uuid.TimeOrderedUUID())
	tempPath := "$env:TEMP\\" + tempFile

	if os.Getenv("WINRMCP_DEBUG") != "" {
		log.Printf("Copying file to %s\n", tempPath)
	}

	err := uploadContent(client, config.MaxShells, config.MaxOperationsPerShell, "%TEMP%\\"+tempFile, in)
	if err != nil {
		return errors.New(fmt.Sprintf("Error uploading file to %s: %v", tempPath, err))
	}

	if os.Getenv("WINRMCP_DEBUG") != "" {
		log.Printf("Moving file from %s to %s", tempPath, toPath)
	}

	err = restoreContent(client, tempFile, toPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error restoring file from %s to %s: %v", tempPath, toPath, err))
	}

	if os.Getenv("WINRMCP_DEBUG") != "" {
		log.Printf("Removing temporary file %s", tempPath)
	}

	err = cleanupContent(client, fmt.Sprintf("%s.tmp", toPath))
	if err != nil {
		return errors.New(fmt.Sprintf("Error removing temporary file %s: %v", tempPath, err))
	}

	return nil
}

func uploadContent(client *winrm.Client, maxShell int, maxChunks int, filePath string, reader io.Reader) error {
	var err error
	var piece = 0
	var wg sync.WaitGroup

	if maxChunks == 0 {
		maxChunks = 1
	}

	if maxChunks == 0 {
		maxChunks = 10
	}

	// Create 4 Parallel workers
	for p := 0; p < maxShell; p++ {
		done := make(chan bool, 1)
		// Add worker to the WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
		Loop:
			for {
				select {
				case <-done:
					break Loop
				default:
					// Create a shell
					shell, gerr := client.CreateShell()
					if gerr != nil {
						err = gerr
						break Loop
					}
					defer shell.Close()

					// Each shell can do X amount of chunks per session
					for c := 0; c < maxChunks; c++ {
						// Read a chunk
						piece++
						content, finished, gerr := getChunk(reader, filePath)
						if gerr != nil {
							err = gerr
							done <- true
							shell.Close()
							break
						}
						if finished {
							done <- true
							shell.Close()
							break
						}

						gerr = uploadChunks(shell, fmt.Sprintf("%v.%v", filePath, piece), content)
						if gerr != nil {
							err = gerr
							done <- true
						}
					}
					shell.Close()
				}
			}
		}()
	}
	wg.Wait()
	return err
}
func getChunk(reader io.Reader, filePath string) (string, bool, error) {
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

	n, err := reader.Read(chunk)
	if err != nil && err != io.EOF {
		return "", false, err
	}
	if n == 0 {
		return "", true, nil
	}

	content := base64.StdEncoding.EncodeToString(chunk[:n])

	return content, false, nil

}

func uploadChunks(shell *winrm.Shell, filePath string, content string) error {
	// Upload chunk
	err := appendContent(shell, filePath, content)
	return err
}

func restoreContent(client *winrm.Client, fileLike, toPath string) error {
	shell, err := client.CreateShell()
	if err != nil {
		return err
	}

	defer shell.Close()
	script := fmt.Sprintf(`
		Write-Host ""
		$dest_file_path = [System.IO.Path]::GetFullPath("%s")
		$dest_file_path_temp = [System.IO.Path]::GetFullPath("$dest_file_path.tmp")
		if (Test-Path $dest_file_path) {
			rm $dest_file_path
		}
		else {
			$dest_dir = ([System.IO.Path]::GetDirectoryName($dest_file_path))
			New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path $dest_dir | Out-Null
		}

		$file_list = Get-ChildItem $env:TEMP |
		where {$_.Name -like "%s*"}

		# Get the number from the last part of the file and sort
		$file_list | foreach {
				$_ | Add-Member -Name Number -MemberType NoteProperty -Value -1
				$_.Number = [int]$_.Name.Substring($_.Name.IndexOf("tmp.")+4)
		}
		$file_list = $file_list | sort { $_.Number }

		if (Test-Path $dest_file_path_temp) {
			rm $dest_file_path_temp
		}
		# For each file in the list, add it to a main file
		$file_list | foreach {
				$tmp_file_path = [System.IO.Path]::GetFullPath($_.FullName)
				$content = Get-Content $tmp_file_path
				Add-Content -Path $dest_file_path_temp -Value $content
				rm $tmp_file_path
		}

		$base64_lines = Get-Content $dest_file_path_temp
		$base64_string = [string]::join("",$base64_lines)
		$bytes = [System.Convert]::FromBase64String($base64_string)
		[System.IO.File]::WriteAllBytes($dest_file_path, $bytes)
	`, toPath, fileLike)

	cmd, err := shell.Execute(winrm.Powershell(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return errors.New(fmt.Sprintf("restore operation returned code=%d", cmd.ExitCode()))
	}
	return nil
}

func cleanupContent(client *winrm.Client, toPath string) error {
	shell, err := client.CreateShell()
	if err != nil {
		return err
	}

	defer shell.Close()
	cmd, _ := shell.Execute("powershell", "Remove-Item", toPath, "-ErrorAction SilentlyContinue")
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
	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return errors.New(fmt.Sprintf("upload operation returned code=%d", cmd.ExitCode()))
	}

	return nil
}

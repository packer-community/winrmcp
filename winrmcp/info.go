package winrmcp

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/masterzen/winrm/winrm"
)

type Info struct {
	WinRM      WinrmConfig
	PowerShell PsSettings
}

type WinrmConfig struct {
	MaxEnvelopeSizeKB int `xml:"MaxEnvelopeSizekb"`
	MaxTimeoutMS      int `xml:"MaxTimeoutms"`
	Service           WinrmServiceConfig
	Winrs             WinrmWinrsConfig
}

type WinrmServiceConfig struct {
	MaxConnections                 int
	MaxConcurrentOperations        int
	MaxConcurrentOperationsPerUser int
}

type WinrmWinrsConfig struct {
	MaxMemoryPerShellMB  int
	MaxShellsPerUser     int
	MaxConcurrentUsers   int
	MaxProcessesPerShell int
}

type PsSettings struct {
	Version         string
	ExecutionPolicy string
}

func fetchInfo(client *winrm.Client) (*Info, error) {
	var err error
	info := &Info{
		WinRM:      WinrmConfig{},
		PowerShell: PsSettings{},
	}

	err = runPsVersion(client, &info.PowerShell)
	if err != nil {
		return info, err
	}

	err = runPsExecutionPolicy(client, &info.PowerShell)
	if err != nil {
		return info, err
	}

	err = runWinrmConfig(client, &info.WinRM)
	if err != nil {
		return info, err
	}

	return info, nil
}

func runWinrmConfig(client *winrm.Client, config *WinrmConfig) error {
	// generate command
	var err error
	var buffer bytes.Buffer
	err = elevatedTemplate.Execute(&buffer, struct {
		Command, User, Password string
	}{
		"winrm get winrm/config -format:xml", "packer", "packer",
	})

	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't compile elevated command: %v", err))
	}

	// execute command
	stdout, stderr, err := client.RunWithString(
		fmt.Sprintf("powershell -EncodedCommand %s", psencode(buffer.Bytes())), "")
	if err != nil {
		err = errors.New(fmt.Sprintf("Couldn't execute elevated command: %v", err))
	}

	if stderr != "" {
		if os.Getenv("WINRMCP_DEBUG") != "" {
			log.Printf("STDERR returned: %s\n", stderr)
		}
	}

	if err == nil {
		if stdout != "" {
			err = xml.Unmarshal([]byte(stdout), config)
		}
	}

	return err
}

func runPsVersion(client *winrm.Client, settings *PsSettings) error {
	script := "$PSVersionTable.PSVersion | ConvertTo-Xml -NoTypeInformation -As String"
	stdout, stderr, err := client.RunWithString("powershell -Command \""+script+"\"", "")

	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't execute script %s: %v", script, err))
	}

	if stderr != "" {
		if os.Getenv("WINRMCP_DEBUG") != "" {
			log.Printf("STDERR returned: %s\n", stderr)
		}
	}

	if stdout != "" {
		doc := pslist{}
		err := xml.Unmarshal([]byte(stdout), &doc)
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't parse results: %v", err))
		}

		settings.Version = doc.Objects[0].Value
	}

	return nil
}

func runPsExecutionPolicy(client *winrm.Client, settings *PsSettings) error {
	script := "Get-ExecutionPolicy | Select-Object"
	stdout, stderr, err := client.RunWithString("powershell -Command \""+script+"\"", "")

	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't execute script %s: %v", script, err))
	}

	if stderr != "" {
		if os.Getenv("WINRMCP_DEBUG") != "" {
			log.Printf("STDERR returned: %s\n", stderr)
		}
	}

	if stdout != "" {
		settings.ExecutionPolicy = strings.Trim(stdout, "\n")
	}

	return nil
}

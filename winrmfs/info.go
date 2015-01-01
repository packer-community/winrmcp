package winrmfs

import (
	"encoding/xml"

	"github.com/masterzen/winrm/winrm"
)

type Info struct {
	WinRM WinrmConfig
}

type WinrmConfig struct {
	MaxEnvelopeSizekb int
	Service           WinrmServiceConfig
	Winrs             WinrmWinrsConfig
}

type WinrmServiceConfig struct {
	MaxConcurrentOperationsPerUser int
}

type WinrmWinrsConfig struct {
	MaxMemoryPerShellMB int
}

func fetchInfo(client *winrm.Client) (*Info, error) {
	stdout, stderr, err := client.RunWithString("winrm get winrm/config -format:xml", "")
	if err != nil {
		return nil, err
	}

	info := &Info{
		WinRM: WinrmConfig{},
	}

	if stdout != "" {
		err := xml.Unmarshal([]byte(stdout), &info.WinRM)
		return info, err
	}

	if stderr != "" {
		println(stderr)
	}

	return info, nil
}

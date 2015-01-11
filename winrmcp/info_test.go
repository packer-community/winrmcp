package winrmcp

import (
	"io"
	"testing"

	"github.com/dylanmei/winrmtest"
	"github.com/masterzen/winrm/winrm"
)

func Test_fetching_powershell_version(t *testing.T) {
	h := winrmtest.NewRemote()
	client := winrm.NewClient(&winrm.Endpoint{h.Host, h.Port}, "test", "test")
	defer h.Close()

	script := "$PSVersionTable.PSVersion | ConvertTo-Xml -NoTypeInformation -As String"
	h.CommandFunc("powershell -Command \""+script+"\"", func(out, err io.Writer) int {
		out.Write([]byte(`<?xml version="1.0"?>
		<Objects>
			<Object>1.234</Object>
		</Objects>`))
		return 0
	})

	psSettings := PsSettings{}
	err := runPsVersion(client, &psSettings)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if psSettings.Version != "1.234" {
		t.Errorf("expected PowerShell version to be 1.234 but was %s", psSettings.Version)
	}
}

func Test_fetching_powershell_execution_policy(t *testing.T) {
	h := winrmtest.NewRemote()
	client := winrm.NewClient(&winrm.Endpoint{h.Host, h.Port}, "test", "test")
	defer h.Close()

	h.CommandFunc("powershell -Command \"Get-ExecutionPolicy | Select-Object\"", func(out, err io.Writer) int {
		out.Write([]byte("TestPolicy\n"))
		return 0
	})

	psSettings := PsSettings{}
	err := runPsExecutionPolicy(client, &psSettings)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if psSettings.ExecutionPolicy != "TestPolicy" {
		t.Errorf("expected PowerShell ExecutionPolicy to be TestPolicy but was %s", psSettings.ExecutionPolicy)
	}
}

func Test_fetching_winrm_info(t *testing.T) {
	t.Skip("Not compatible with elevated script")
	h := winrmtest.NewRemote()
	client := winrm.NewClient(&winrm.Endpoint{h.Host, h.Port}, "test", "test")
	defer h.Close()

	h.CommandFunc("winrm get winrm/config -format:xml", func(out, err io.Writer) int {
		out.Write([]byte(`<?xml version="1.0"?>
<cfg:Config xmlns:cfg="http://schemas.microsoft.com/wbem/wsman/1/config">
  <cfg:MaxEnvelopeSizekb>1</cfg:MaxEnvelopeSizekb>
  <cfg:MaxTimeoutms>2</cfg:MaxTimeoutms>
  <cfg:Service>
    <cfg:MaxConcurrentOperations>1</cfg:MaxConcurrentOperations>
    <cfg:MaxConcurrentOperationsPerUser>2</cfg:MaxConcurrentOperationsPerUser>
    <cfg:MaxConnections>3</cfg:MaxConnections>
  </cfg:Service>
  <cfg:Winrs>
    <cfg:MaxConcurrentUsers>1</cfg:MaxConcurrentUsers>
    <cfg:MaxProcessesPerShell>2</cfg:MaxProcessesPerShell>
    <cfg:MaxMemoryPerShellMB>3</cfg:MaxMemoryPerShellMB>
    <cfg:MaxShellsPerUser>4</cfg:MaxShellsPerUser>
  </cfg:Winrs>
</cfg:Config>`))
		return 0
	})

	winrmConfig := WinrmConfig{}
	err := runWinrmConfig(client, &winrmConfig)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if winrmConfig.MaxEnvelopeSizeKB != 1 {
		t.Errorf("expected WinRM/MaxEnvelopeSizeKB to be 1 but was %d", winrmConfig.MaxEnvelopeSizeKB)
	}
	if winrmConfig.MaxTimeoutMS != 2 {
		t.Errorf("expected WinRM/MaxTimeoutMS to be 2 but was %d", winrmConfig.MaxTimeoutMS)
	}
	if winrmConfig.Service.MaxConcurrentOperations != 1 {
		t.Errorf("expected WinRM/Service/MaxConcurrentOperations to be 1 but was %d", winrmConfig.Service.MaxConcurrentOperations)
	}
	if winrmConfig.Service.MaxConcurrentOperationsPerUser != 2 {
		t.Errorf("expected WinRM/Service/MaxConcurrentOperationsPerUser to be 2 but was %d", winrmConfig.Service.MaxConcurrentOperationsPerUser)
	}
	if winrmConfig.Service.MaxConnections != 3 {
		t.Errorf("expected WinRM/Service/MaxConnections to be 3 but was %d", winrmConfig.Service.MaxConnections)
	}
	if winrmConfig.Winrs.MaxConcurrentUsers != 1 {
		t.Errorf("expected WinRM/Winrs/MaxConcurrentUsers to be 1 but was %d", winrmConfig.Winrs.MaxConcurrentUsers)
	}
	if winrmConfig.Winrs.MaxProcessesPerShell != 2 {
		t.Errorf("expected WinRM/Winrs/MaxProcessesPerShell to be 2 but was %d", winrmConfig.Winrs.MaxProcessesPerShell)
	}
	if winrmConfig.Winrs.MaxMemoryPerShellMB != 3 {
		t.Errorf("expected WinRM/Winrs/MaxMemoryPerShellMB to be 3 but was %d", winrmConfig.Winrs.MaxMemoryPerShellMB)
	}
	if winrmConfig.Winrs.MaxShellsPerUser != 4 {
		t.Errorf("expected WinRM/Winrs/MaxShellsPerUser to be 4 but was %d", winrmConfig.Winrs.MaxShellsPerUser)
	}
}

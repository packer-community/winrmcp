package winrmcp

import (
	"io"
	"testing"

	"github.com/dylanmei/winrmtest"
	"github.com/masterzen/winrm/winrm"
)

func Test_fetching_winrm_info(t *testing.T) {
	h := winrmtest.NewRemote()
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

	client := winrm.NewClient(&winrm.Endpoint{h.Host, h.Port}, "test", "test")
	info, err := fetchInfo(client)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info.WinRM.MaxEnvelopeSizeKB != 1 {
		t.Errorf("expected WinRM/MaxEnvelopeSizeKB to be 1 but was %d", info.WinRM.MaxEnvelopeSizeKB)
	}
	if info.WinRM.MaxTimeoutMS != 2 {
		t.Errorf("expected WinRM/MaxTimeoutMS to be 2 but was %d", info.WinRM.MaxTimeoutMS)
	}
	if info.WinRM.Service.MaxConcurrentOperations != 1 {
		t.Errorf("expected WinRM/Service/MaxConcurrentOperations to be 1 but was %d", info.WinRM.Service.MaxConcurrentOperations)
	}
	if info.WinRM.Service.MaxConcurrentOperationsPerUser != 2 {
		t.Errorf("expected WinRM/Service/MaxConcurrentOperationsPerUser to be 2 but was %d", info.WinRM.Service.MaxConcurrentOperationsPerUser)
	}
	if info.WinRM.Service.MaxConnections != 3 {
		t.Errorf("expected WinRM/Service/MaxConnections to be 3 but was %d", info.WinRM.Service.MaxConnections)
	}
	if info.WinRM.Winrs.MaxConcurrentUsers != 1 {
		t.Errorf("expected WinRM/Winrs/MaxConcurrentUsers to be 1 but was %d", info.WinRM.Winrs.MaxConcurrentUsers)
	}
	if info.WinRM.Winrs.MaxProcessesPerShell != 2 {
		t.Errorf("expected WinRM/Winrs/MaxProcessesPerShell to be 2 but was %d", info.WinRM.Winrs.MaxProcessesPerShell)
	}
	if info.WinRM.Winrs.MaxMemoryPerShellMB != 3 {
		t.Errorf("expected WinRM/Winrs/MaxMemoryPerShellMB to be 3 but was %d", info.WinRM.Winrs.MaxMemoryPerShellMB)
	}
	if info.WinRM.Winrs.MaxShellsPerUser != 4 {
		t.Errorf("expected WinRM/Winrs/MaxShellsPerUser to be 4 but was %d", info.WinRM.Winrs.MaxShellsPerUser)
	}

	//MaxConnections                 int
	//MaxConcurrentOperations        int
	//MaxConcurrentOperationsPerUser int
}

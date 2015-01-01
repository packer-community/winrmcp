package winrmfs

import "github.com/masterzen/winrm/winrm"

type Winrmfs struct {
	client *winrm.Client
}

func New(client *winrm.Client) *Winrmfs {
	return &Winrmfs{client}
}

func (fs *Winrmfs) Info() (*Info, error) {
	return fetchInfo(fs.client)
}

func (fs *Winrmfs) Copy(fromPath, toPath string) error {
	return doCopy(fs.client, fromPath, toPath)
}


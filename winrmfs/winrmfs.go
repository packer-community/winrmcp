package winrmfs

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/masterzen/winrm/winrm"
)

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
	file, err := os.Open(fromPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't open file %s: %v", fromPath, err))
	}

	defer file.Close()
	return fs.Write(toPath, file)
}

func (fs *Winrmfs) Write(toPath string, src io.Reader) error {
	return doCopy(fs.client, src, winPath(toPath))
}

func (fs *Winrmfs) List(remotePath string) ([]FileItem, error) {
	return fetchList(fs.client, winPath(remotePath))
}

package winrmfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	f, err := os.Open(fromPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't read file %s: %v", fromPath, err))
	}

	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't stat file %s: %v", fromPath, err))
	}

	if !fi.IsDir() {
		return fs.Write(toPath, f)
	} else {
		fw := fileWalker{
			client:  fs.client,
			toDir:   toPath,
			fromDir: fromPath,
		}
		return filepath.Walk(fromPath, fw.copyFile)
	}
}

func (fs *Winrmfs) Write(toPath string, src io.Reader) error {
	return doCopy(fs.client, src, winPath(toPath))
}

func (fs *Winrmfs) List(remotePath string) ([]FileItem, error) {
	return fetchList(fs.client, winPath(remotePath))
}

type fileWalker struct {
	client  *winrm.Client
	toDir   string
	fromDir string
}

func (fw *fileWalker) copyFile(fromPath string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !shouldUploadFile(fi) {
		return nil
	}

	relPath := filepath.Dir(fromPath[len(fw.toDir):len(fromPath)])
	toPath := filepath.Join(fw.toDir, relPath, fi.Name())

	f, err := os.Open(fromPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't read file %s: %v", fromPath, err))
	}

	return doCopy(fw.client, f, winPath(toPath))
}

func shouldUploadFile(fi os.FileInfo) bool {
	// Ignore dir entries and OS X special hidden file
	return !fi.IsDir() && ".DS_Store" != fi.Name()
}

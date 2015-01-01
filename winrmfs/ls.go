package winrmfs

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/masterzen/winrm/winrm"
)

type FileItem struct {
	Name          string
	Path          string
	Mode          string
	LastWriteTime string
	Length        int
}

type list struct {
	Objects []object `xml:"Object"`
}

type object struct {
	Properties []objectProperty `xml:"Property"`
}

type objectProperty struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:",innerxml"`
}

func fetchList(client *winrm.Client, remotePath string) ([]FileItem, error) {
	shell, err := client.CreateShell()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't create shell: %v", err))
	}

	defer shell.Close()
	script := fmt.Sprintf("Get-ChildItem %s", remotePath)
	stdout, stderr, err := client.RunWithString("powershell -Command \""+script+" | ConvertTo-Xml -NoTypeInformation -As String\"", "")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't execute script %s: %v", script, err))
	}

	if stderr != "" {
		if os.Getenv("WINRMFS_DEBUG") != "" {
			log.Printf("STDERR returned: %s\n", stderr)
		}
	}

	if stdout != "" {
		doc := list{}
		err := xml.Unmarshal([]byte(stdout), &doc)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't parse results: %v", err))
		}

		return convertObjects(doc.Objects), nil
	}

	return []FileItem{}, nil
}

func convertObjects(objects []object) []FileItem {
	items := make([]FileItem, len(objects))

	for i, object := range objects {
		for _, property := range object.Properties {
			switch property.Name {
			case "Name":
				items[i].Name = property.Value
			case "Mode":
				items[i].Mode = property.Value
			case "FullName":
				items[i].Path = property.Value
			case "Length":
				if n, err := strconv.Atoi(property.Value); err == nil {
					items[i].Length = n
				}
			case "LastWriteTime":
				items[i].LastWriteTime = property.Value
			}
		}
	}

	return items
}

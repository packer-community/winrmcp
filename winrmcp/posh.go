package winrmcp

import (
	"encoding/base64"
)

type pslist struct {
	Objects []psobject `xml:"Object"`
}

type psobject struct {
	Properties []psproperty `xml:"Property"`
	Value      string       `xml:",innerxml"`
}

type psproperty struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:",innerxml"`
}

func psencode(buffer []byte) string {
	// 2 byte chars to make PowerShell happy
	wideCmd := ""
	for _, b := range buffer {
		wideCmd += string(b) + "\x00"
	}

	// Base64 encode the command
	input := []uint8(wideCmd)
	return base64.StdEncoding.EncodeToString(input)
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/masterzen/winrm/winrm"
)

func addrFlag(f *flag.FlagSet) *string {
	defaultAddr := os.Getenv("WINRMCP_ADDR")
	if defaultAddr == "" {
		defaultAddr = "localhost:5985"
	}

	return f.String("addr", defaultAddr,
		"Remote address of the target machine")
}

func parseEndpoint(addr string) (*winrm.Endpoint, error) {
	host, port, err := parseHostPort(addr)
	if err != nil {
		return nil, err
	}
	if port == 0 {
		port = 5985
	}
	return &winrm.Endpoint{
		Host: host, Port: port,
	}, nil
}

func parseHostPort(addr string) (host string, port int, err error) {
	host = ""
	port = 0

	if addr == "" {
		err = errors.New("Couldn't convert empty argument to an address.")
	}

	segments := strings.Split(addr, ":")

	if len(segments) == 1 {
		host = segments[0]
		return
	}

	if len(segments) == 2 {
		convPort, convErr := strconv.Atoi(segments[1])
		if convErr == nil {
			host = segments[0]
			port = convPort
		} else {
			err = errors.New(fmt.Sprintf("Couldn't convert \"%s\" to a port number.", segments[1]))
		}
		return
	}

	err = errors.New(fmt.Sprintf("Couldn't convert \"%s\" to an address.", addr))
	return
}

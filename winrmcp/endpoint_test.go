package winrmcp

import (
	"testing"
)

func Test_parsing_an_addr_to_a_winrm_endpoint(t *testing.T) {
	endpoint, err := parseEndpoint("1.2.3.4:1234")

	if err != nil {
		t.Fatalf("Should not have been an error: %v", err)
	}
	if endpoint == nil {
		t.Error("Endpoint should not be nil")
	}
	if endpoint.Host != "1.2.3.4" {
		t.Error("Host should be 1.2.3.4")
	}
	if endpoint.Port != 1234 {
		t.Error("Port should be 1234")
	}
}

func Test_parsing_an_addr_without_a_port_to_a_winrm_endpoint(t *testing.T) {
	endpoint, err := parseEndpoint("1.2.3.4")

	if err != nil {
		t.Fatalf("Should not have been an error: %v", err)
	}
	if endpoint == nil {
		t.Error("Endpoint should not be nil")
	}
	if endpoint.Host != "1.2.3.4" {
		t.Error("Host should be 1.2.3.4")
	}
	if endpoint.Port != 5985 {
		t.Error("Port should be 5985")
	}
}

func Test_parsing_an_empty_addr_to_a_winrm_endpoint(t *testing.T) {
	endpoint, err := parseEndpoint("")

	if endpoint != nil {
		t.Error("Endpoint should be nil")
	}
	if err == nil {
		t.Error("Expected an error")
	}
}

func Test_parsing_an_addr_with_a_bad_port(t *testing.T) {
	endpoint, err := parseEndpoint("1.2.3.4:ABCD")

	if endpoint != nil {
		t.Error("Endpoint should be nil")
	}
	if err == nil {
		t.Error("Expected an error")
	}
}

package resolver

import (
	"fmt"

	"github.com/jinliming2/encrypt-dns/config"
	"github.com/miekg/dns"
)

// TLSDNSClient resolves DNS with DNS-over-TLS
type TLSDNSClient struct {
	host    []string
	port    uint16
	timeout uint
	config.DNSSettings
}

// NewTLSDNSClient returns a new TLS DNS client
func NewTLSDNSClient(host []string, port uint16, timeout uint, settings config.DNSSettings) *TLSDNSClient {
	return &TLSDNSClient{
		host:        host,
		port:        port,
		timeout:     timeout,
		DNSSettings: settings,
	}
}

func (client *TLSDNSClient) String() string {
	return fmt.Sprintf("tls://%s:%d", client.host, client.port)
}

// Resolve DNS
func (client *TLSDNSClient) Resolve(request *dns.Msg, useTCP bool) *dns.Msg {
	return nil
}

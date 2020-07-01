package client

import (
	"fmt"

	"github.com/jinliming2/encrypt-dns/config"
	"github.com/miekg/dns"
)

type tlsDNSClient struct {
	host    []string
	port    uint16
	timeout uint
	config.DNSSettings
}

func newTLSDNSClient(host []string, port uint16, timeout uint, settings config.DNSSettings) *tlsDNSClient {
	return &tlsDNSClient{
		host:        host,
		port:        port,
		timeout:     timeout,
		DNSSettings: settings,
	}
}

func (client *tlsDNSClient) string() string {
	return fmt.Sprintf("tls://%s:%d", client.host, client.port)
}

func (client *tlsDNSClient) resolve(request *dns.Msg, useTCP bool) *dns.Msg {
	return nil
}

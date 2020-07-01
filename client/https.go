package client

import (
	"fmt"

	"github.com/jinliming2/encrypt-dns/config"
	"github.com/miekg/dns"
)

type httpsDNSClient struct {
	host    []string
	port    uint16
	path    string
	timeout uint
	config.DNSSettings
}

func newHTTPSDNSClient(host []string, port uint16, path string, google, cookie bool, timeout uint, settings config.DNSSettings) *httpsDNSClient {
	return &httpsDNSClient{
		host:        host,
		port:        port,
		path:        path,
		timeout:     timeout,
		DNSSettings: settings,
	}
}

func (client *httpsDNSClient) string() string {
	return fmt.Sprintf("https://%s:%d%s", client.host, client.port, client.path)
}

func (client *httpsDNSClient) resolve(request *dns.Msg, useTCP bool) *dns.Msg {
	return nil
}

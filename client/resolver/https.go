package resolver

import (
	"fmt"

	"github.com/jinliming2/encrypt-dns/config"
	"github.com/miekg/dns"
)

// HTTPSDNSClient resolves DNS with DNS-over-HTTPS
type HTTPSDNSClient struct {
	host    []string
	port    uint16
	path    string
	timeout uint
	config.DNSSettings
}

// NewHTTPSDNSClient returns a new HTTPS DNS client
func NewHTTPSDNSClient(host []string, port uint16, hostname string, path string, cookie bool, timeout uint, settings config.DNSSettings, bootstrap DNSClient) *HTTPSDNSClient {
	return &HTTPSDNSClient{
		host:        host,
		port:        port,
		path:        path,
		timeout:     timeout,
		DNSSettings: settings,
	}
}

func (client *HTTPSDNSClient) String() string {
	return fmt.Sprintf("https://%s:%d%s", client.host, client.port, client.path)
}

// Resolve DNS
func (client *HTTPSDNSClient) Resolve(request *dns.Msg, useTCP bool) (*dns.Msg, error) {
	return request, nil
}

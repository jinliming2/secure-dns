package resolver

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/jinliming2/secure-dns/config"
	"github.com/miekg/dns"
)

// HTTPSGoogleDNSClient resolves DNS with DNS-over-HTTPS Google API
type HTTPSGoogleDNSClient struct {
	host      []string
	port      uint16
	hostnames []*hostnameAddress
	client    *http.Client
	path      string
	timeout   uint
	config.DNSSettings
}

// NewHTTPSGoogleDNSClient returns a new HTTPS DNS client using Google API
func NewHTTPSGoogleDNSClient(host []string, port uint16, hostname string, path string, cookie bool, timeout uint, settings config.DNSSettings, bootstrap DNSClient) (*HTTPSGoogleDNSClient, error) {
	hostnames, err := resolveTLS(host, port, hostname, bootstrap, "HTTPS Client")
	if err != nil {
		return nil, err
	}

	var jar *cookiejar.Jar
	if cookie {
		jar, _ = cookiejar.New(nil)
	}

	return &HTTPSGoogleDNSClient{
		host:      host,
		port:      port,
		hostnames: hostnames,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Duration(timeout),
		},
		path:        path,
		timeout:     timeout,
		DNSSettings: settings,
	}, nil
}

func (client *HTTPSGoogleDNSClient) String() string {
	return fmt.Sprintf("https+google://%s:%d%s", client.host, client.port, client.path)
}

// Resolve DNS
func (client *HTTPSGoogleDNSClient) Resolve(request *dns.Msg, useTCP bool) (*dns.Msg, error) {
	return request, nil
}

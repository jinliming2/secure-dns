package resolver

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/miekg/dns"
)

// TLSDNSClient resolves DNS with DNS-over-TLS
type TLSDNSClient struct {
	host      []string
	port      uint16
	hostnames []*hostnameAddress
	clients   map[string]*dns.Client
	timeout   uint
	config.DNSSettings
}

// NewTLSDNSClient returns a new TLS DNS client
func NewTLSDNSClient(
	host []string,
	port uint16,
	hostname string,
	timeout uint,
	settings config.DNSSettings,
	bootstrap DNSClient,
) (*TLSDNSClient, error) {
	hostnames, err := resolveTLS(host, port, hostname, bootstrap, "TLS Client")
	if err != nil {
		return nil, err
	}

	clients := map[string]*dns.Client{}
	for _, hip := range hostnames {
		if c, ok := clients[hip.hostname]; !ok || c == nil {
			clients[hip.hostname] = &dns.Client{
				Net: "tcp-tls",
				TLSConfig: &tls.Config{
					ServerName:         hip.hostname,
					ClientSessionCache: tls.NewLRUClientSessionCache(-1),
				},
				Timeout: time.Duration(timeout),
			}
		}
	}

	return &TLSDNSClient{
		host:        host,
		port:        port,
		hostnames:   hostnames,
		clients:     clients,
		timeout:     timeout,
		DNSSettings: settings,
	}, nil
}

func (client *TLSDNSClient) String() string {
	return fmt.Sprintf("tls://%s:%d", client.host, client.port)
}

// Resolve DNS
func (client *TLSDNSClient) Resolve(request *dns.Msg, useTCP bool) (*dns.Msg, error) {
	ecs.SetECS(request, client.NoECS, client.CustomECS)
	// TODO: use random hostname
	address := client.hostnames[0]
	// TODO: use random address
	res, _, err := client.clients[address.hostname].Exchange(request, address.address[0])
	if err != nil {
		return getEmptyErrorResponse(request), fmt.Errorf("Failed to resolve %s using %s", request.Question[0].Name, client.String())
	}
	return res, nil
}

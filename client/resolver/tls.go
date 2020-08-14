package resolver

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/miekg/dns"
)

// TLSDNSClient resolves DNS with DNS-over-TLS
type TLSDNSClient struct {
	host      []string
	port      uint16
	addresses []string
	client    *dns.Client
	timeout   time.Duration
	config.DNSSettings
}

// NewTLSDNSClient returns a new TLS DNS client
func NewTLSDNSClient(
	host []string,
	port uint16,
	hostname string,
	timeout time.Duration,
	settings config.DNSSettings,
	bootstrap *net.Resolver,
) (*TLSDNSClient, error) {

	addresses := make([]string, len(host))
	for index, h := range host {
		if ip := net.ParseIP(h); ip != nil && ip.To4() == nil {
			addresses[index] = fmt.Sprintf("[%s]:%d", h, port)
		} else {
			addresses[index] = fmt.Sprintf("%s:%d", h, port)
		}
	}

	return &TLSDNSClient{
		host:      host,
		port:      port,
		addresses: addresses,
		client: &dns.Client{
			Net: "tcp-tls",
			TLSConfig: &tls.Config{
				ServerName:         hostname,
				ClientSessionCache: tls.NewLRUClientSessionCache(-1),
			},
			Dialer: &net.Dialer{
				Resolver: bootstrap,
			},
			Timeout:        timeout,
			SingleInflight: !settings.NoSingleInflight,
		},
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
	res, _, err := client.client.Exchange(request, client.addresses[randomSource.Intn(len(client.addresses))])
	if err != nil {
		return getEmptyErrorResponse(request), fmt.Errorf("Failed to resolve %s using %s", request.Question[0].Name, client.String())
	}
	// https://github.com/miekg/dns/issues/1145
	res.Id = request.Id
	return res, nil
}

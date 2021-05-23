package resolver

import (
	"fmt"
	"net"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/miekg/dns"
)

// TraditionalDNSClient resolves DNS with traditional DNS client
type TraditionalDNSClient struct {
	host      []string
	port      uint16
	addresses []string
	udpClient *dns.Client
	tcpClient *dns.Client
	timeout   time.Duration
	config.DNSSettings
}

// NewTraditionalDNSClient returns a new traditional DNS client
func NewTraditionalDNSClient(host []string, port uint16, timeout time.Duration, settings config.DNSSettings) *TraditionalDNSClient {
	addresses := make([]string, len(host))
	for index, h := range host {
		if ip := net.ParseIP(h); ip != nil && ip.To4() == nil {
			addresses[index] = fmt.Sprintf("[%s]:%d", h, port)
		} else {
			addresses[index] = fmt.Sprintf("%s:%d", h, port)
		}
	}
	return &TraditionalDNSClient{
		host:      host,
		port:      port,
		addresses: addresses,
		udpClient: &dns.Client{
			Net:            "udp",
			UDPSize:        dns.DefaultMsgSize,
			Timeout:        timeout,
			SingleInflight: !settings.NoSingleInflight,
		},
		tcpClient: &dns.Client{
			Net:            "tcp",
			Timeout:        timeout,
			SingleInflight: !settings.NoSingleInflight,
		},
		timeout:     timeout,
		DNSSettings: settings,
	}
}

func (client *TraditionalDNSClient) String() string {
	return fmt.Sprintf("dns://%s:%d", client.host, client.port)
}

func (client *TraditionalDNSClient) FallbackNoECSEnabled() bool {
	return client.FallbackNoECS
}

// Resolve DNS
func (client *TraditionalDNSClient) Resolve(request *dns.Msg, useTCP bool, forceNoECS bool) (*dns.Msg, error) {
	var c *dns.Client
	if useTCP {
		c = client.tcpClient
	} else {
		c = client.udpClient
	}
	ecs.SetECS(request, forceNoECS || client.NoECS, client.CustomECS)
	res, _, err := c.Exchange(request, client.addresses[randomSource.Intn(len(client.addresses))])
	if err != nil {
		return getEmptyErrorResponse(request), fmt.Errorf("Failed to resolve %s using %s", request.Question[0].Name, client.String())
	}
	// https://github.com/miekg/dns/issues/1145
	res.Id = request.Id
	return res, nil
}

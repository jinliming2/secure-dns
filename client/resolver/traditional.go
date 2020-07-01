package resolver

import (
	"fmt"
	"time"

	"github.com/jinliming2/encrypt-dns/client/ecs"
	"github.com/jinliming2/encrypt-dns/config"
	"github.com/miekg/dns"
)

// TraditionalDNSClient resolves DNS with traditional DNS client
type TraditionalDNSClient struct {
	host      []string
	port      uint16
	addresses []string
	udpClient *dns.Client
	tcpClient *dns.Client
	timeout   uint
	config.DNSSettings
}

// NewTraditionalDNSClient returns a new traditional DNS client
func NewTraditionalDNSClient(host []string, port uint16, timeout uint, settings config.DNSSettings) *TraditionalDNSClient {
	addresses := make([]string, len(host))
	for index, h := range host {
		addresses[index] = fmt.Sprintf("%s:%d", h, port)
	}
	return &TraditionalDNSClient{
		host:      host,
		port:      port,
		addresses: addresses,
		udpClient: &dns.Client{
			Net:     "udp",
			UDPSize: dns.DefaultMsgSize,
			Timeout: time.Duration(timeout),
		},
		tcpClient: &dns.Client{
			Net:     "tcp",
			Timeout: time.Duration(timeout),
		},
		timeout:     timeout,
		DNSSettings: settings,
	}
}

func (client *TraditionalDNSClient) String() string {
	return fmt.Sprintf("dns://%s:%d", client.host, client.port)
}

// Resolve DNS
func (client *TraditionalDNSClient) Resolve(request *dns.Msg, useTCP bool) *dns.Msg {
	var c *dns.Client
	if useTCP {
		c = client.tcpClient
	} else {
		c = client.udpClient
	}
	ecs.SetECS(request, client.NoECS, client.CustomECS)
	// return request
	// TODO: use random address
	res, _, err := c.Exchange(request, client.addresses[0])
	if err != nil {
		reply := &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Id:       request.Id,
				Response: true,
				Opcode:   request.Opcode,
				Rcode:    dns.RcodeServerFailure,
			},
			Compress: true,
			Question: make([]dns.Question, len(request.Question)),
		}
		copy(reply.Question, request.Question)
		return reply
	}
	return res
}

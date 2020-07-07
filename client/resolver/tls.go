package resolver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/miekg/dns"
)

type hostnameAddress struct {
	hostname string
	address  []string
}

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
func NewTLSDNSClient(host []string, port uint16, hostname string, timeout uint, settings config.DNSSettings, bootstrap DNSClient) (*TLSDNSClient, error) {
	var hostnames []*hostnameAddress
	var defaultAddress []string
	for _, h := range host {
		ip := net.ParseIP(h)
		if ip != nil {
			if ip.To4() != nil {
				defaultAddress = append(defaultAddress, fmt.Sprintf("%s:%d", ip, port))
			} else {
				defaultAddress = append(defaultAddress, fmt.Sprintf("[%s]:%d", ip, port))
			}
			continue
		}
		if bootstrap == nil {
			return nil, errors.New("No bootstrap found for TLS Client")
		}
		// TODO: do not resolve all names at bootstrap
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(h), dns.TypeA)
		res, err := bootstrap.Resolve(msg, false)
		if err != nil || res.Rcode != dns.RcodeSuccess || len(res.Answer) == 0 {
			res, err = bootstrap.Resolve(msg, true)
			if err != nil {
				continue
			}
		}
		addressList := make([]string, len(res.Answer))
		index := 0
		for _, answer := range res.Answer {
			if a, ok := answer.(*dns.A); ok {
				addressList[index] = fmt.Sprintf("%s:%d", a.A, port)
				index++
			}
		}
		hostnames = append(hostnames, &hostnameAddress{hostname: h, address: addressList})
	}

	if len(defaultAddress) > 0 {
		hostnames = append(hostnames, &hostnameAddress{hostname: hostname, address: defaultAddress})
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
		return reply, fmt.Errorf("Failed to resolve %s using %s", request.Question[0].Name, client.String())
	}
	return res, nil
}

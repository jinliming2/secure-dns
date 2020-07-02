package resolver

import "github.com/miekg/dns"

// DNSClient is a DNS client
type DNSClient interface {
	String() string
	Resolve(*dns.Msg, bool) *dns.Msg
}

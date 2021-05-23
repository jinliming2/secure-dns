package resolver

import (
	"math/rand"
	"regexp"
	"time"

	"github.com/miekg/dns"
)

var (
	regexDNSMsg = regexp.MustCompile("\\bapplication/dns-message\\b")

	mimeDNSMsg = "application/dns-message"

	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// DNSClient is a DNS client
type DNSClient interface {
	String() string
	FallbackNoECSEnabled() bool
	Resolve(*dns.Msg, bool, bool) (*dns.Msg, error)
}

type addressHostname struct {
	address  string
	hostname string
}

func getEmptyResponse(request *dns.Msg) *dns.Msg {
	return new(dns.Msg).SetReply(request)
}

func getEmptyErrorResponse(request *dns.Msg) *dns.Msg {
	return new(dns.Msg).SetRcode(request, dns.RcodeServerFailure)
}

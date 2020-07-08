package resolver

import (
	"regexp"

	"github.com/miekg/dns"
)

var (
	regexDNSMsg = regexp.MustCompile("\\bapplication/dns-message\\b")

	mimeDNSMsg = "application/dns-message"
)

// DNSClient is a DNS client
type DNSClient interface {
	String() string
	Resolve(*dns.Msg, bool) (*dns.Msg, error)
}

type hostnameAddress struct {
	hostname string
	address  []string
}

func getEmptyResponse(request *dns.Msg) *dns.Msg {
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:       request.Id,
			Response: true,
			Opcode:   request.Opcode,
			Rcode:    dns.RcodeSuccess,
		},
		Compress: true,
		Question: make([]dns.Question, len(request.Question)),
	}
	copy(msg.Question, request.Question)
	return msg
}

func getEmptyErrorResponse(request *dns.Msg) *dns.Msg {
	msg := getEmptyResponse(request)
	msg.Rcode = dns.RcodeServerFailure
	return msg
}

package resolver

import (
	"fmt"

	"github.com/miekg/dns"
)

// HostsDNSClient resolves DNS with Hosts
type HostsDNSClient struct {
	records map[string][]string
}

// NewHostsDNSClient returns a new hosts DNS client
func NewHostsDNSClient(records map[string][]string) *HostsDNSClient {
	return &HostsDNSClient{records: records}
}

func (client *HostsDNSClient) String() string {
	return "HOSTS resolver"
}

func (client *HostsDNSClient) FallbackNoECSEnabled() bool {
	return false
}

// Resolve DNS
func (client *HostsDNSClient) Resolve(request *dns.Msg, useTCP bool, forceNoECS bool) (reply *dns.Msg, _ error) {
	reply = getEmptyResponse(request)

	question := request.Question[0]

	var questionType string
	var ok bool
	if questionType, ok = dns.TypeToString[question.Qtype]; !ok {
		reply.Rcode = dns.RcodeFormatError
		return
	}

	var records []string
	if records, ok = client.records[questionType]; !ok {
		reply.Answer = make([]dns.RR, 0)
		return
	}

	reply.Answer = make([]dns.RR, len(records))

	for index, record := range records {
		zone := fmt.Sprintf("%s 0 IN %s %s", question.Name, questionType, record)
		if rr, err := dns.NewRR(zone); err == nil {
			reply.Answer[index] = rr
		} else {
			reply.Rcode = dns.RcodeServerFailure
			reply.Answer = nil
			return
		}
	}

	return
}

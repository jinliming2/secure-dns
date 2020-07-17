package resolver

import (
	"time"

	"github.com/miekg/dns"
)

// FixRecordTTL of reply dns msg
func FixRecordTTL(rr dns.RR, delta time.Duration) {
	header := rr.Header()
	if header.Rrtype == dns.TypeOPT {
		return
	}
	old := time.Duration(header.Ttl) * time.Second
	new := old - delta
	if new > 0 {
		header.Ttl = uint32(new / time.Second)
	} else {
		header.Ttl = 0
	}
}

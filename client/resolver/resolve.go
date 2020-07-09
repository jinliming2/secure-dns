package resolver

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/miekg/dns"
)

func resolveTLS(
	host []string,
	port uint16,
	hostname string,
	bootstrap DNSClient,
	errName string,
) (hostnames []*hostnameAddress, err error) {
	var defaultAddress []string
	for _, h := range host {
		tlsURL, _ := url.Parse("tls://" + h)
		tlsHostname := tlsURL.Hostname()
		if tlsHostname == "" {
			continue
		}
		ip := net.ParseIP(tlsHostname)
		if ip != nil {
			if ip.To4() != nil {
				defaultAddress = append(defaultAddress, fmt.Sprintf("%s:%d", ip, port))
			} else {
				defaultAddress = append(defaultAddress, fmt.Sprintf("[%s]:%d", ip, port))
			}
			continue
		}
		if bootstrap == nil {
			return nil, fmt.Errorf("No bootstrap found for %s: %s:%d", errName, tlsHostname, port)
		}
		// TODO: do not resolve all names at bootstrap
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(tlsHostname), dns.TypeA)
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
		if len(addressList) > 0 {
			hostnames = append(hostnames, &hostnameAddress{hostname: tlsHostname, address: addressList})
		}
	}

	if len(defaultAddress) > 0 {
		hostnames = append(hostnames, &hostnameAddress{hostname: hostname, address: defaultAddress})
	}

	return
}

func fixHTTPSRecordTTL(rr dns.RR, delta time.Duration) {
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

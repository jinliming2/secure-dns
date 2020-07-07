package client

import (
	"strings"

	"github.com/jinliming2/secure-dns/client/resolver"
)

type customResolver struct {
	matcher  func(string) bool
	resolver resolver.DNSClient
}

func newCustomResolver(resolver resolver.DNSClient, domain, suffix []string) *customResolver {
	domainList := make([]string, len(domain))
	for index, d := range domain {
		domainList[index] = strings.Trim(d, ".")
	}

	suffixList := make([]string, len(suffix))
	for index, s := range suffix {
		suffixList[index] = "." + strings.Trim(s, ".")
	}

	return &customResolver{
		matcher: func(domain string) bool {
			trimmedDomain := strings.Trim(domain, ".")

			for _, d := range domainList {
				if trimmedDomain == d {
					return true
				}
			}

			trimmedDomain = "." + trimmedDomain
			for _, s := range suffixList {
				if strings.HasSuffix(trimmedDomain, s) {
					return true
				}
			}

			return false
		},
		resolver: resolver,
	}
}

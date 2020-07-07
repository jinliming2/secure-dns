package client

import (
	"strings"

	"github.com/jinliming2/secure-dns/client/resolver"
	"github.com/jinliming2/secure-dns/config"
	"go.uber.org/zap"
)

// NewClient returns a client with dnsClients
func NewClient(logger *zap.SugaredLogger, conf *config.Config) (client *Client) {
	client = &Client{logger: logger, timeout: conf.Config.Timeout}

	logger.Info("creating clients...")

	for _, traditional := range conf.Traditional {
		dnsConfig := config.DNSSettings{
			CustomECS: append(traditional.CustomECS, conf.Config.CustomECS...),
			NoECS:     conf.Config.NoECS || traditional.NoECS,
		}
		c := resolver.NewTraditionalDNSClient(traditional.Host, traditional.Port, conf.Config.Timeout, dnsConfig)

		if traditional.Bootstrap {
			logger.Debugf("new traditional resolver: %s (for bootstrap)", c.String())
			if client.bootstrap != nil {
				logger.Warnf("only one bootstrap resolver allowed, ignoring %s...", c.String())
				continue
			}
			if len(traditional.Domain)+len(traditional.Suffix) > 0 {
				logger.Warn("domain and suffix doesn't support for bootstrap resolver")
			}
			client.bootstrap = c
		} else if len(traditional.Domain)+len(traditional.Suffix) > 0 {
			logger.Debugf("new traditional resolver: %s (for specified domain or suffix use)", c.String())
			cr := newCustomResolver(c, traditional.Domain, traditional.Suffix)
			client.custom = append(client.custom, cr)
		} else {
			logger.Debugf("new traditional resolver: %s", c.String())
			client.upstream = append(client.upstream, c)
		}
	}

	for _, tls := range conf.TLS {
		dnsConfig := config.DNSSettings{
			CustomECS: append(tls.CustomECS, conf.Config.CustomECS...),
			NoECS:     conf.Config.NoECS || tls.NoECS,
		}
		c, err := resolver.NewTLSDNSClient(tls.Host, tls.Port, tls.Hostname, conf.Config.Timeout, dnsConfig, client.bootstrap)
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(tls.Domain)+len(tls.Suffix) > 0 {
			logger.Debugf("new TLS resolver: %s (for specified domain or suffix use)", c.String())
			cr := newCustomResolver(c, tls.Domain, tls.Suffix)
			client.custom = append(client.custom, cr)
		} else {
			logger.Debugf("new TLS resolver: %s", c.String())
			client.upstream = append(client.upstream, c)
		}
	}

	for _, https := range conf.HTTPS {
		dnsConfig := config.DNSSettings{
			CustomECS: append(https.CustomECS, conf.Config.CustomECS...),
			NoECS:     conf.Config.NoECS || https.NoECS,
		}
		var c resolver.DNSClient
		if https.Google {
			c = resolver.NewHTTPSGoogleDNSClient(https.Host, https.Port, https.Hostname, https.Path, https.Cookie, conf.Config.Timeout, dnsConfig, client.bootstrap)
		} else {
			c = resolver.NewHTTPSDNSClient(https.Host, https.Port, https.Hostname, https.Path, https.Cookie, conf.Config.Timeout, dnsConfig, client.bootstrap)
		}

		if len(https.Domain)+len(https.Suffix) > 0 {
			logger.Debugf("new HTTPS resolver: %s (for specified domain or suffix use)", c.String())
			cr := newCustomResolver(c, https.Domain, https.Suffix)
			client.custom = append(client.custom, cr)
		} else {
			logger.Debugf("new HTTPS resolver: %s", c.String())
			client.upstream = append(client.upstream, c)
		}
	}

	for domain, b := range conf.Hosts {
		c := resolver.NewHostsDNSClient(b)
		if strings.HasPrefix(domain, "*.") {
			domain = strings.TrimLeft(domain, "*.")
			logger.Debugf("new HOSTS resolver: %s (for wildcard domain)", domain)
			cr := newCustomResolver(c, []string{}, []string{domain})
			client.custom = append(client.custom, cr)
		} else {
			logger.Debugf("new HOSTS resolver: %s", domain)
			cr := newCustomResolver(c, []string{domain}, []string{})
			client.custom = append(client.custom, cr)
		}
	}

	return
}

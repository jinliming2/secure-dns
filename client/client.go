package client

import (
	"context"
	"strings"

	"github.com/jinliming2/secure-dns/client/resolver"
	"github.com/jinliming2/secure-dns/config"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

type customResolver struct {
	matcher  func(string) bool
	resolver resolver.DNSClient
}

// Client handles DNS requests
type Client struct {
	logger  *zap.SugaredLogger
	timeout uint

	bootstrap resolver.DNSClient
	upstream  []resolver.DNSClient

	custom []*customResolver

	servers []*dns.Server
}

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

// ListenAndServe listen on addresses and serve DNS service
func (client *Client) ListenAndServe(addr []string) error {
	client.servers = make([]*dns.Server, 2*len(addr))

	results := make(chan error)

	client.logger.Info("creating server...")
	for _, address := range addr {
		client.logger.Debugf("new server: %s", address)
		udpServer := &dns.Server{
			Addr:    address,
			Net:     "udp",
			Handler: dns.HandlerFunc(client.udpHandlerFunc),
			UDPSize: dns.DefaultMsgSize,
		}
		tcpServer := &dns.Server{
			Addr:    address,
			Net:     "tcp",
			Handler: dns.HandlerFunc(client.tcpHandlerFunc),
		}
		go startDNSServer(udpServer, client.logger, results)
		go startDNSServer(tcpServer, client.logger, results)
		client.servers = append(client.servers, udpServer, tcpServer)
	}

	for i := 0; i < 2*len(addr); i++ {
		if err := <-results; err != nil {
			client.Shutdown()
			return err
		}
	}

	close(results)
	return nil
}

// Shutdown shuts down a server
func (client *Client) Shutdown() []error {
	return client.ShutdownContext(context.Background())
}

// ShutdownContext shuts down a server
func (client *Client) ShutdownContext(ctx context.Context) (errors []error) {
	client.logger.Info("shutting down servers")
	for _, server := range client.servers {
		if server != nil {
			client.logger.Debugf("shutting down server %s://%s", server.Net, server.Addr)
			if err := server.ShutdownContext(ctx); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return
}

func startDNSServer(server *dns.Server, logger *zap.SugaredLogger, results chan error) {
	err := server.ListenAndServe()
	if err != nil {
		logger.Errorf("server %s://%s exited with error: %s", server.Net, server.Addr, err.Error())
	}
	results <- err
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

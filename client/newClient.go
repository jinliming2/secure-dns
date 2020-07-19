package client

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/jinliming2/secure-dns/client/cache"
	"github.com/jinliming2/secure-dns/client/resolver"
	"github.com/jinliming2/secure-dns/config"
	"github.com/jinliming2/secure-dns/selector"
	"go.uber.org/zap"
)

// NewClient returns a client with dnsClients
func NewClient(logger *zap.SugaredLogger, conf *config.Config) (client *Client, err error) {
	timeout := time.Duration(conf.Config.Timeout) * time.Second
	client = &Client{logger: logger, timeout: timeout}

	switch conf.Config.RoundRobin {
	case config.SelectorClock:
		client.upstream = &selector.Clock{}
	case config.SelectorRandom:
		client.upstream = &selector.Random{}
	case config.SelectorSWRR:
		client.upstream = &selector.SWrr{}
	case config.SelectorWRandom:
		client.upstream = &selector.WRandom{}
	default:
		err = fmt.Errorf("No such round robin: %s", conf.Config.RoundRobin)
		return
	}

	logger.Info("creating clients...")

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

	var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

traditionalLoop:
	for _, traditional := range conf.Traditional {
		if traditional.Bootstrap {
			logger.Debugf("new traditional resolver: %s (for bootstrap)", fmt.Sprintf("dns://%s:%d", traditional.Host, traditional.Port))
			if client.bootstrap != nil {
				logger.Warnf("only one bootstrap resolver allowed, ignoring %s...", fmt.Sprintf("dns://%s:%d", traditional.Host, traditional.Port))
				continue
			}
			if len(traditional.Domain)+len(traditional.Suffix) > 0 {
				logger.Warn("domain and suffix doesn't support for bootstrap resolver")
			}
			count := len(traditional.Host)
			addresses := make([]string, count)
			for index, host := range traditional.Host {
				if addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, traditional.Port)); err == nil {
					addresses[index] = addr.String()
				} else if addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%d", host, traditional.Port)); err == nil {
					addresses[index] = addr.String()
				} else {
					logger.Warnf("parse bootstrap address failed: %s:%d, [%s]:%d", host, traditional.Port, host, traditional.Port)
					continue traditionalLoop
				}
			}
			client.bootstrap = &net.Resolver{
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, network, addresses[randomSource.Intn(count)])
				},
			}
			continue
		}

		dnsConfig := config.DNSSettings{
			CustomECS:        append(traditional.CustomECS, conf.Config.CustomECS...),
			NoECS:            conf.Config.NoECS || traditional.NoECS,
			NoSingleInflight: conf.Config.NoSingleInflight || traditional.NoSingleInflight,
		}
		c := resolver.NewTraditionalDNSClient(traditional.Host, traditional.Port, timeout, dnsConfig)

		if len(traditional.Domain)+len(traditional.Suffix) > 0 {
			logger.Debugf("new traditional resolver: %s (for specified domain or suffix use)", c.String())
			cr := newCustomResolver(c, traditional.Domain, traditional.Suffix)
			client.custom = append(client.custom, cr)
		} else {
			logger.Debugf("new traditional resolver: %s", c.String())
			client.upstream.Add(traditional.Weight, c)
		}
	}

	for _, tls := range conf.TLS {
		dnsConfig := config.DNSSettings{
			CustomECS:        append(tls.CustomECS, conf.Config.CustomECS...),
			NoECS:            conf.Config.NoECS || tls.NoECS,
			NoSingleInflight: conf.Config.NoSingleInflight || tls.NoSingleInflight,
		}
		c, err := resolver.NewTLSDNSClient(tls.Host, tls.Port, tls.Hostname, timeout, dnsConfig, client.bootstrap)
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
			client.upstream.Add(tls.Weight, c)
		}
	}

	for _, https := range conf.HTTPS {
		dnsConfig := config.DNSSettings{
			CustomECS:        append(https.CustomECS, conf.Config.CustomECS...),
			NoECS:            conf.Config.NoECS || https.NoECS,
			NoUserAgent:      conf.Config.NoUserAgent || https.NoUserAgent,
			NoSingleInflight: conf.Config.NoSingleInflight || https.NoSingleInflight,
		}
		if https.UserAgent != "" {
			dnsConfig.UserAgent = https.UserAgent
		} else if conf.Config.UserAgent != "" {
			dnsConfig.UserAgent = conf.Config.UserAgent
		}
		var c resolver.DNSClient
		var err error
		if https.Google {
			c, err = resolver.NewHTTPSGoogleDNSClient(
				https.Host,
				https.Port,
				https.Hostname,
				https.Path,
				https.Cookie,
				timeout,
				dnsConfig,
				client.bootstrap,
				logger,
			)
		} else {
			c, err = resolver.NewHTTPSDNSClient(
				https.Host,
				https.Port,
				https.Hostname,
				https.Path,
				https.Cookie,
				timeout,
				dnsConfig,
				client.bootstrap,
				logger,
			)
		}
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(https.Domain)+len(https.Suffix) > 0 {
			logger.Debugf("new HTTPS resolver: %s (for specified domain or suffix use)", c.String())
			cr := newCustomResolver(c, https.Domain, https.Suffix)
			client.custom = append(client.custom, cr)
		} else {
			logger.Debugf("new HTTPS resolver: %s", c.String())
			client.upstream.Add(https.Weight, c)
		}
	}

	client.upstream.Start()
	logger.Infof("using round robin: %s", client.upstream.Name())

	if !conf.Config.NoCache {
		client.cacher = cache.NewCache()
	}

	return
}

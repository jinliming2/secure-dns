package client

import (
	"context"
	"net"
	"time"

	"github.com/jinliming2/secure-dns/client/cache"
	"github.com/jinliming2/secure-dns/selector"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// Client handles DNS requests
type Client struct {
	logger  *zap.SugaredLogger
	timeout time.Duration

	bootstrap *net.Resolver
	upstream  selector.Selector

	custom []*customResolver

	servers []*dns.Server

	cacher        *cache.Cache
	cacheNoAnswer uint32
}

func startDNSServer(server *dns.Server, logger *zap.SugaredLogger, results chan error) {
	err := server.ListenAndServe()
	if err != nil {
		logger.Errorf("server %s://%s exited with error: %s", server.Net, server.Addr, err.Error())
	}
	results <- err
}

// ListenAndServe listen on addresses and serve DNS service
func (client *Client) ListenAndServe(addr []string) error {
	client.servers = make([]*dns.Server, 0, 2*len(addr))

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
	if client.cacher != nil {
		client.cacher.Destroy()
	}
	return
}

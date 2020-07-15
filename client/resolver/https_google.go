package resolver

import (
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/jinliming2/secure-dns/versions"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

// HTTPSGoogleDNSClient resolves DNS with DNS-over-HTTPS Google API
type HTTPSGoogleDNSClient struct {
	host      []string
	port      uint16
	addresses []addressHostname
	client    *http.Client
	path      string
	timeout   uint
	logger    *zap.SugaredLogger
	config.DNSSettings
}

// NewHTTPSGoogleDNSClient returns a new HTTPS DNS client using Google API
func NewHTTPSGoogleDNSClient(
	host []string,
	port uint16,
	hostname, path string,
	cookie bool,
	timeout uint,
	settings config.DNSSettings,
	bootstrap *net.Resolver,
	logger *zap.SugaredLogger,
) (*HTTPSGoogleDNSClient, error) {

	addresses := make([]addressHostname, len(host))
	for index, h := range host {
		if ip := net.ParseIP(h); ip != nil {
			if ip.To4() == nil {
				addresses[index] = addressHostname{address: fmt.Sprintf("[%s]:%d", h, port), hostname: hostname}
			} else {
				addresses[index] = addressHostname{address: fmt.Sprintf("%s:%d", h, port), hostname: hostname}
			}
		} else {
			addresses[index] = addressHostname{address: fmt.Sprintf("%s:%d", h, port)}
		}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{
		Resolver: bootstrap,
	}
	transport.DialContext = dialer.DialContext

	var jar http.CookieJar
	if cookie {
		jar, _ = cookiejar.New(nil)
	}

	return &HTTPSGoogleDNSClient{
		host:      host,
		port:      port,
		addresses: addresses,
		client: &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   time.Duration(timeout),
		},
		path:        path,
		timeout:     timeout,
		logger:      logger,
		DNSSettings: settings,
	}, nil
}

func (client *HTTPSGoogleDNSClient) String() string {
	return fmt.Sprintf("https+google://%s:%d%s", client.host, client.port, client.path)
}

// Resolve DNS
func (client *HTTPSGoogleDNSClient) Resolve(request *dns.Msg, useTCP bool) (*dns.Msg, error) {
	ecs.SetECS(request, client.NoECS, client.CustomECS)

	query := url.Values{}
	query.Set("name", request.Question[0].Name)
	query.Set("type", strconv.FormatUint(uint64(request.Question[0].Qtype), 10))
	if request.CheckingDisabled {
		query.Set("cd", "1")
	}
	query.Set("ct", mimeDNSMsg)
	if opt := request.IsEdns0(); opt != nil {
		if opt.Do() {
			query.Set("do", "1")
		}
		for _, option := range opt.Option {
			if option.Option() == dns.EDNS0SUBNET {
				eDNS0Subnet := option.(*dns.EDNS0_SUBNET)
				subnet := fmt.Sprintf("%s/%d", eDNS0Subnet.Address.String(), eDNS0Subnet.SourceNetmask)
				query.Set("edns_client_subnet", subnet)
			}
		}
	}
	// TODO: random padding
	// query.Set("random_padding", "")

	// TODO: use random address
	url := fmt.Sprintf("https://%s%s?%s", client.addresses[0].address, client.path, query.Encode())

	req, err := http.NewRequest(http.MethodGet, url, nil)
	client.logger.Debugf("[%d] GET %s", request.Id, url)
	if err != nil {
		return getEmptyErrorResponse(request), err
	}
	req.Header.Set("accept", mimeDNSMsg)
	req.Close = false
	if client.addresses[0].hostname != "" {
		req.Host = client.addresses[0].hostname
	}

	if client.NoUserAgent {
		req.Header.Set("user-agent", "")
	} else if client.UserAgent != "" {
		req.Header.Set("user-agent", client.UserAgent)
	} else {
		req.Header.Set("user-agent", versions.USERAGENT)
	}

	// TODO: use random address
	return httpsGetDNSMessage(request, req, client.client, client.addresses[0], client.path, client.logger)
}

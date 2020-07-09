package resolver

import (
	"fmt"
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
	hostnames []*hostnameAddress
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
	bootstrap DNSClient,
	logger *zap.SugaredLogger,
) (*HTTPSGoogleDNSClient, error) {
	hostnames, err := resolveTLS(host, port, hostname, bootstrap, "HTTPS Client")
	if err != nil {
		return nil, err
	}

	var jar http.CookieJar
	if cookie {
		jar, _ = cookiejar.New(nil)
	}

	return &HTTPSGoogleDNSClient{
		host:      host,
		port:      port,
		hostnames: hostnames,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Duration(timeout),
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
	// TODO: use random hostname
	address := client.hostnames[0]

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
	url := fmt.Sprintf("https://%s%s?%s", address.address[0], client.path, query.Encode())

	req, err := http.NewRequest(http.MethodGet, url, nil)
	client.logger.Debugf("[%d] GET %s", request.Id, url)
	if err != nil {
		return getEmptyErrorResponse(request), err
	}
	req.Header.Set("accept", mimeDNSMsg)
	req.Close = false
	req.Host = address.hostname

	if client.NoUserAgent {
		req.Header.Set("user-agent", "")
	} else if client.UserAgent != "" {
		req.Header.Set("user-agent", client.UserAgent)
	} else {
		req.Header.Set("user-agent", versions.USERAGENT)
	}

	return httpsGetDNSMessage(request, req, client.client, address.address[0], address.hostname, client.path, client.logger)
}

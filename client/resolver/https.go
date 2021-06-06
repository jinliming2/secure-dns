package resolver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/jinliming2/secure-dns/versions"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// HTTPSDNSClient resolves DNS with DNS-over-HTTPS
type HTTPSDNSClient struct {
	host           []string
	port           uint16
	addresses      []addressHostname
	client         *http.Client
	path           string
	timeout        time.Duration
	singleInflight *singleflight.Group
	logger         *zap.SugaredLogger
	config.DNSSettings
}

// NewHTTPSDNSClient returns a new HTTPS DNS client
func NewHTTPSDNSClient(
	host []string,
	port uint16,
	hostname, path string,
	cookie bool,
	timeout time.Duration,
	settings config.DNSSettings,
	bootstrap *net.Resolver,
	logger *zap.SugaredLogger,
) (*HTTPSDNSClient, error) {

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

	var sf *singleflight.Group
	if !settings.NoSingleInflight {
		sf = &singleflight.Group{}
	}

	return &HTTPSDNSClient{
		host:      host,
		port:      port,
		addresses: addresses,
		client: &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   timeout,
		},
		path:           path,
		timeout:        timeout,
		singleInflight: sf,
		logger:         logger,
		DNSSettings:    settings,
	}, nil
}

func (client *HTTPSDNSClient) String() string {
	return fmt.Sprintf("https://%s:%d%s", client.host, client.port, client.path)
}

func (client *HTTPSDNSClient) ECSDisabled() bool {
	return client.NoECS
}

func (client *HTTPSDNSClient) FallbackNoECSEnabled() bool {
	return client.FallbackNoECS
}

// Resolve DNS
func (client *HTTPSDNSClient) Resolve(request *dns.Msg, useTCP bool, forceNoECS bool) (*dns.Msg, error) {
	return httpsSingleInflightRequest(request, forceNoECS, client.singleInflight, client.resolve)
}

func (client *HTTPSDNSClient) resolve(request *dns.Msg, forceNoECS bool) (*dns.Msg, error) {
	ecs.SetECS(request, forceNoECS || client.NoECS, client.CustomECS)

	msg, err := request.Pack()
	if err != nil {
		reply := getEmptyErrorResponse(request)
		reply.Rcode = dns.RcodeFormatError
		return reply, err
	}

	data := base64.RawURLEncoding.EncodeToString(msg)

	address := client.addresses[randomSource.Intn(len(client.addresses))]

	url := fmt.Sprintf("https://%s%s?dns=%s", address.address, client.path, data)

	var req *http.Request
	if len(url) < 2048 {
		req, err = http.NewRequest(http.MethodGet, url, nil)
		client.logger.Debugf("[%d] GET %s", request.Id, url)
		if err != nil {
			return getEmptyErrorResponse(request), err
		}
	} else {
		req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s%s", address.address, client.path), bytes.NewReader(msg))
		if err != nil {
			return getEmptyErrorResponse(request), err
		}
		req.ContentLength = int64(len(msg))
		req.Header.Set("content-type", mimeDNSMsg)
		client.logger.Debugf("[%d] POST %s with %d bytes body", request.Id, req.URL, req.ContentLength)
	}
	req.Header.Set("accept", mimeDNSMsg)
	req.Close = false
	if address.hostname != "" {
		req.Host = address.hostname
	}

	if client.NoUserAgent {
		req.Header.Set("user-agent", "")
	} else if client.UserAgent != "" {
		req.Header.Set("user-agent", client.UserAgent)
	} else {
		req.Header.Set("user-agent", versions.USERAGENT)
	}

	return httpsGetDNSMessage(request, req, client.client, address, client.path, client.logger)
}

func httpsSingleInflightRequest(
	request *dns.Msg,
	forceNoECS bool,
	singleInflight *singleflight.Group,
	resolve func(request *dns.Msg, forceNoECS bool) (*dns.Msg, error),
) (*dns.Msg, error) {
	if singleInflight == nil {
		return resolve(request, forceNoECS)
	}

	question := request.Question[0]
	key := fmt.Sprintf("%s:%d:%d", question.Name, question.Qtype, question.Qclass)

	result := <-singleInflight.DoChan(key, func() (interface{}, error) {
		return resolve(request, forceNoECS)
	})

	if result.Err != nil || result.Val == nil {
		return getEmptyErrorResponse(request), result.Err
	}

	reply := result.Val.(*dns.Msg)
	if result.Shared {
		reply = reply.Copy()
	}
	reply.Id = request.Id

	return reply, nil
}

func httpsGetDNSMessage(
	request *dns.Msg,
	req *http.Request,
	client *http.Client,
	address addressHostname,
	path string,
	logger *zap.SugaredLogger,
) (*dns.Msg, error) {
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return getEmptyErrorResponse(request), err
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return getEmptyErrorResponse(request), fmt.Errorf("HTTP error from %s%s: %d %s", address, path, res.StatusCode, res.Status)
	}
	contentType := res.Header.Get("content-type")
	if !regexDNSMsg.MatchString(contentType) {
		return getEmptyErrorResponse(request), fmt.Errorf("HTTP unsupported MIME type: %s", contentType)
	}

	logger.Debugf("[%d] %s: %s", request.Id, res.Status, contentType)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return getEmptyErrorResponse(request), err
	}

	reply := new(dns.Msg)
	err = reply.Unpack(body)
	if err != nil {
		return getEmptyErrorResponse(request), err
	}
	reply.Id = request.Id

	headerLastModified := res.Header.Get("last-modified")
	if headerLastModified != "" {
		if modifiedTime, err := time.Parse(http.TimeFormat, headerLastModified); err == nil {
			now := time.Now().UTC()
			headerDate := res.Header.Get("date")
			if headerDate != "" {
				if date, err := time.Parse(http.TimeFormat, headerDate); err == nil {
					now = date
				}
			}
			delta := now.Sub(modifiedTime)
			if delta > 0 {
				for _, rr := range reply.Answer {
					FixRecordTTL(rr, delta)
				}
				for _, rr := range reply.Ns {
					FixRecordTTL(rr, delta)
				}
				for _, rr := range reply.Extra {
					FixRecordTTL(rr, delta)
				}
			}
		}
	}

	return reply, nil
}

package resolver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/jinliming2/secure-dns/client/ecs"
	"github.com/jinliming2/secure-dns/config"
	"github.com/jinliming2/secure-dns/versions"
	"github.com/miekg/dns"
)

// HTTPSDNSClient resolves DNS with DNS-over-HTTPS
type HTTPSDNSClient struct {
	host      []string
	port      uint16
	hostnames []*hostnameAddress
	client    *http.Client
	path      string
	timeout   uint
	config.DNSSettings
}

// NewHTTPSDNSClient returns a new HTTPS DNS client
func NewHTTPSDNSClient(host []string, port uint16, hostname string, path string, cookie bool, timeout uint, settings config.DNSSettings, bootstrap DNSClient) (*HTTPSDNSClient, error) {
	hostnames, err := resolveTLS(host, port, hostname, bootstrap, "HTTPS Client")
	if err != nil {
		return nil, err
	}

	var jar http.CookieJar
	if cookie {
		jar, _ = cookiejar.New(nil)
	}

	return &HTTPSDNSClient{
		host:      host,
		port:      port,
		hostnames: hostnames,
		client: &http.Client{
			Jar:     jar,
			Timeout: time.Duration(timeout),
		},
		path:        path,
		timeout:     timeout,
		DNSSettings: settings,
	}, nil
}

func (client *HTTPSDNSClient) String() string {
	return fmt.Sprintf("https://%s:%d%s", client.host, client.port, client.path)
}

// Resolve DNS
func (client *HTTPSDNSClient) Resolve(request *dns.Msg, useTCP bool) (*dns.Msg, error) {
	ecs.SetECS(request, client.NoECS, client.CustomECS)
	// TODO: use random hostname
	address := client.hostnames[0]

	msg, err := request.Pack()
	if err != nil {
		reply := getEmptyErrorResponse(request)
		reply.Rcode = dns.RcodeFormatError
		return reply, err
	}

	data := base64.RawURLEncoding.EncodeToString(msg)

	// TODO: use random address
	url := fmt.Sprintf("https://%s%s?dns=%s", address.address[0], client.path, data)

	var req *http.Request
	if len(url) < 2048 {
		req, err = http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return getEmptyErrorResponse(request), err
		}
	} else {
		req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s%s", address.address[0], client.path), bytes.NewReader(msg))
		if err != nil {
			return getEmptyErrorResponse(request), err
		}
		req.ContentLength = int64(len(msg))
		req.Header.Set("content-type", "application/dns-message")
	}
	req.Header.Set("accept", "application/dns-message")
	req.Close = false
	req.Host = address.hostname

	if client.NoUserAgent {
		req.Header.Set("user-agent", "")
	} else if client.UserAgent != "" {
		req.Header.Set("user-agent", client.UserAgent)
	} else {
		req.Header.Set("user-agent", versions.USERAGENT)
	}

	return httpsGetDNSMessage(request, req, client.client, address.address[0], address.hostname, client.path)
}

func httpsGetDNSMessage(request *dns.Msg, req *http.Request, client *http.Client, address, hostname, path string) (*dns.Msg, error) {
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return getEmptyErrorResponse(request), err
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return getEmptyErrorResponse(request), fmt.Errorf("HTTP error from %s%s (%s): %s", address, path, hostname, res.Status)
	}
	if !dnsMsgRegex.MatchString(res.Header.Get("content-type")) {
		return getEmptyErrorResponse(request), fmt.Errorf("HTTP unsupported MIME type: %s", res.Header.Get("content-type"))
	}

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
					fixHTTPSRecordTTL(rr, delta)
				}
				for _, rr := range reply.Ns {
					fixHTTPSRecordTTL(rr, delta)
				}
				for _, rr := range reply.Extra {
					fixHTTPSRecordTTL(rr, delta)
				}
			}
		}
	}

	return reply, nil
}

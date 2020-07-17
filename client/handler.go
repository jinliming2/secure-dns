package client

import (
	"fmt"

	"github.com/jinliming2/secure-dns/client/resolver"
	"github.com/miekg/dns"
)

func (client *Client) tcpHandlerFunc(w dns.ResponseWriter, r *dns.Msg) {
	client.handlerFunc(w, r, true)
}

func (client *Client) udpHandlerFunc(w dns.ResponseWriter, r *dns.Msg) {
	client.handlerFunc(w, r, false)
}

func (client *Client) handlerFunc(w dns.ResponseWriter, r *dns.Msg, useTCP bool) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(client.timeout)*time.Second)
	// defer cancel()

	if r.Response {
		client.logger.Warn("received a response packet")
		return
	}

	if len(r.Question) != 1 {
		client.logger.Warn("request packet contains more than 1 question is not allowed")
		reply := new(dns.Msg).SetReply(r).SetRcodeFormatError(r)
		w.WriteMsg(reply)
		return
	}

	question := &r.Question[0]
	qName := question.Name
	qClass := ""
	qType := ""

	if class, ok := dns.ClassToString[question.Qclass]; ok {
		qClass = class
	} else {
		qClass = fmt.Sprintf("%d", question.Qclass)
	}

	if t, ok := dns.TypeToString[question.Qtype]; ok {
		qType = t
	} else {
		qType = fmt.Sprintf("%d", question.Qtype)
	}

	client.logger.Infow(fmt.Sprintf("[%d] request", r.Id), "name", qName, "class", qClass, "type", qType)

	var c *resolver.DNSClient

	for _, custom := range client.custom {
		if custom.matcher(qName) {
			c = &custom.resolver
			client.logger.Debugf("[%d] using %s for %s [condition]", r.Id, (*c).String(), qName)
			break
		}
	}

	if c == nil {
		if client.upstream.Empty() {
			client.logger.Warnf("no upstream to use for querying %s", qName)
			reply := new(dns.Msg).SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(reply)
			return
		}

		c = client.upstream.Get().Client
		client.logger.Debugf("[%d] using %s for %s", r.Id, (*c).String(), qName)
	}

	response, err := (*c).Resolve(r, useTCP)
	if err != nil {
		client.logger.Warn(err.Error())
	}
	w.WriteMsg(response)
}

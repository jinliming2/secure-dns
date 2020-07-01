package client

import (
	"net"

	"github.com/miekg/dns"
)

func setECS(r *dns.Msg, noECS bool, ecs []net.IP) {
	opt := r.IsEdns0()
	var eDNS0Subnet *dns.EDNS0_SUBNET

	for index, option := range opt.Option {
		if option.Option() == dns.EDNS0SUBNET {
			if noECS {
				opt.Option[index] = opt.Option[len(opt.Option)-1]
				opt.Option = opt.Option[:len(opt.Option)-1]
				return
			}

			eDNS0Subnet = option.(*dns.EDNS0_SUBNET)
			break
		}
	}

	if noECS || len(ecs) == 0 {
		return
	}

	if eDNS0Subnet == nil {
		eDNS0Subnet = new(dns.EDNS0_SUBNET)
		eDNS0Subnet.Code = dns.EDNS0SUBNET
		eDNS0Subnet.SourceScope = 0
		opt.Option = append(opt.Option, eDNS0Subnet)
	}

	// TODO: select ecs randomally
	ip := ecs[0]
	ip4 := ip.To4()

	if ip4 != nil {
		eDNS0Subnet.Family = 1
		eDNS0Subnet.SourceNetmask = 24
		eDNS0Subnet.Address = ip4
	} else {
		eDNS0Subnet.Family = 2
		eDNS0Subnet.SourceNetmask = 56
		eDNS0Subnet.Address = ip
	}
}

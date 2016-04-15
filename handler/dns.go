package handler

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"net/url"

	log "code.google.com/p/log4go"
	"httpdns/misc"
)

func DnsDecoder(urlStr string) (string, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	IP := Dns(u.Host)

	return IP, nil
}

func Dns(host string) *net.IP {
	for _, server := range misc.Conf.Dnses {
		ipAddr := dnss(host, server+":53")
		if ipAddr != nil {
			return ipAddr
		}
	}
	return nil
}

/**
 * CNAME -> A
 */
func dnss(host, server string) *net.IP {
	addrs, err := Lookup("CNAME", host, dnsServer)
	if err != nil {
		log.Warn("dns cname fail with the host[%s]. error: [%s]", host, err.Error())
		return nil
	}

	for {
		if len(addrs.Answer) == 0 {
			break
		}
		host = addrs.Answer[0].(*dns.CNAME).Target
		addrs, err = Lookup("CNAME", host, dnsServer)
		if err != nil {
			log.Warn("dns cname fail with the host[%s]. error: [%s]", host, err.Error())
			return nil
		}
	}
	addrs, err = Lookup("A", host, dnsServer)
	if err != nil {
		log.Warn("dns a fail with the host[%s]. error: [%s]", host, err.Error())
		return nil
	}
	for _, a := range addrs.Answer {
		if a.(*dns.A).A != nil {
			return &a.(*dns.A).A
		}
	}
	return nil
}

func Lookup(ctype, host, dnsServer string) (*dns.Msg, error) {

	itype, ok := dns.StringToType[ctype]
	if !ok {
		return nil, fmt.Errorf("Invalid type %s", ctype)
	}

	host = dns.Fqdn(host)
	client := &dns.Client{}
	msg := &dns.Msg{}
	msg.SetQuestion(host, itype)
	response := &dns.Msg{}

	response, err := lookup(msg, client, dnsServer, false)
	if err != nil {
		return response, err
	}

	return response, nil
}

func lookup(msg *dns.Msg, client *dns.Client, server string, edns bool) (*dns.Msg, error) {
	if edns {
		opt := &dns.OPT{
			Hdr: dns.RR_Header{
				Name:   ".",
				Rrtype: dns.TypeOPT,
			},
		}
		opt.SetUDPSize(dns.DefaultMsgSize)
		msg.Extra = append(msg.Extra, opt)
	}
	response, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, err
	}

	if msg.Id != response.Id {
		return nil, fmt.Errorf("DNS ID mismatch, request: %d, response: %d!", msg.Id, response.Id)
	}

	if response.MsgHdr.Truncated {
		if client.Net == "tcp" {
			return nil, fmt.Errorf("Got truncated message on tcp")
		}
		if edns {
			client.Net = "tcp"
		}

		return lookup(msg, client, server, !edns)
	}
	return response, nil
}

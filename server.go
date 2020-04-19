package main

import (
	"crypto/tls"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"time"
)

// convenience function to generate an nxdomain response
func sendNXDomain(w dns.ResponseWriter, r *dns.Msg) {
	log.Printf("sending nxdomain")
	m := &dns.Msg{}
	m.SetRcode(r, dns.RcodeNameError)
	w.WriteMsg(m)
}

func (server *Server) processResults(r dns.Msg, domain string, rrtype uint16) (Response, error) {
	if r.Rcode != dns.RcodeSuccess {
		// wish we had an easier way of translating the rrtype into human, but i don't got one yet
		return Response{}, fmt.Errorf("got unsuccessful lookup code for domain [%s] rrtype [%d]: [%d]", domain, rrtype, r.Rcode)
	}
	return Response{
		Entry:        r,
		CreationTime: time.Now(),
		Key:          domain,
		Qtype:        rrtype,
	}, nil
}

func (server *Server) RecursiveQuery(domain string, rrtype uint16) (Response, error) {
	RecursiveQueryCounter.Inc()
	// lifted from example code https://github.com/miekg/dns/blob/master/example_test.go
	port := "853"
	// TODO error checking
	config := GetConfiguration()

	m := &dns.Msg{}
	m.SetQuestion(domain, rrtype)
	m.RecursionDesired = true
	// TODO cycle through all servers. cache connection
	var errorstring string
	for _, s := range config.Resolvers {
		r, _, err := server.dnsClient.Exchange(m, s+":"+port)
		if err != nil {
			ResolverErrorsCounter.Inc()
			errorstring = fmt.Sprintf("error looking up domain [%s] on server [%s:%s]: %s: %s", domain, s, port, err, errorstring)
			// build a huge error string of all errors from all servers
			continue
		}
		return server.processResults(*r, domain, rrtype)
	}
	// if we went through each server and couldn't find at least one result, bail with the full error string
	return Response{}, fmt.Errorf(errorstring)
}

// retrieves the record for that domain, either from cache or from
// a recursive query
func (server *Server) RetrieveRecords(domain string, rrtype uint16) (Response, error) {
	// First: check caches

	// Need to keep caches locked until the end of the function because
	// we may need to add records consistently
	server.Cache.Lock()
	defer server.Cache.Unlock()
	cached_response, ok := server.Cache.Get(domain, rrtype)
	if ok {
		CacheHitsCounter.Inc()
		return cached_response, nil
	}

	// Now check the hosted cache (stuff in our zone files that we're taking care of
	server.HostedCache.RLock()
	defer server.HostedCache.RUnlock()
	cached_response, ok = server.HostedCache.Get(domain, rrtype)
	if ok {
		HostedCacheHitsCounter.Inc()
		return cached_response, nil
	}

	// Second, query upstream if there's no cache
	// TODO only do if requested b/c thats what the spec says IIRC
	response, err := server.RecursiveQuery(domain, rrtype)
	if err != nil {
		return response, fmt.Errorf("error running recursive query on domain [%s]: %s\n", domain, err)
	}
	server.Cache.Add(response)
	return response, nil
}

func (server *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	TotalDnsQueriesCounter.Inc()

	msg := dns.Msg{}
	msg.SetReply(r)
	domain := msg.Question[0].Name
	// FIXME when should this be set to what
	msg.Authoritative = false
	msg.RecursionAvailable = true

	response, err := server.RetrieveRecords(domain, r.Question[0].Qtype)
	if err != nil {
		LocalServfailsCounter.Inc()
		log.Printf("error retrieving record for domain [%s]: %s", domain, err)
		// TODO SERVFAIL here (like, seriously bro, do this shit)
		sendNXDomain(w, r)
		return
	}

	msg.Answer = response.Entry.Answer

	w.WriteMsg(&msg)
}

func NewServer() (*Server, error) {
	ret := &Server{
		dnsClient: dns.Client{
			Net:       "tcp-tls",
			TLSConfig: &tls.Config{},
		},
	}
	newcache, err := NewCache()
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize lookup cache: %s\n", err)
	}
	newcache.Init()
	ret.Cache = newcache

	hostedcache, err := NewCache()
	// don't init, we don't clean this one
	ret.HostedCache = hostedcache
	return ret, nil
}

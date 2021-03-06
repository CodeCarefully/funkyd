package main

// General utilites
import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"time"
)

func UpstreamTestingDialer(upstream Upstream) func(addr string) (conn *dns.Conn, err error) {
	expectedAddress := upstream.GetAddress()
	return func(addr string) (conn *dns.Conn, err error) {
		if addr != expectedAddress {
			err = fmt.Errorf("got unexpected address [%s] when dialing upstream [%v], expecting address [%s] to be dialed", addr, upstream, expectedAddress)
		}
		server, client := net.Pipe()
		server.Close()
		return &dns.Conn{Conn: client}, err
	}
}

func WaitForCondition(x int, f func() bool) (result bool) {
	for i := 0; i < x; i++ {
		if result = f(); result {
			break
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	return
}

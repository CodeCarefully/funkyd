package main

// Generic functions and types for servers
import (
	"crypto/tls"
	"fmt"
	"github.com/miekg/dns"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"strings"
	"syscall"
	"time"
)

// making this to support dependency injection into the server
type Client interface {
	// Make a new connection
	Dial(address string) (conn *dns.Conn, err error)

	// Run DNS queries
	ExchangeWithConn(s *dns.Msg, conn *dns.Conn) (r *dns.Msg, rtt time.Duration, err error)
}

// this abstraction helps us test the entire servedns path
type ResponseWriter interface {
	WriteMsg(*dns.Msg) error
}

type Server interface {
	// Needs to handle DNS queries
	dns.Handler

	// Internal function to implement ServeDNS, this allows testing
	HandleDNS(w ResponseWriter, m *dns.Msg)

	// Retrieves a new connection to an upstream
	GetConnection() (*ConnEntry, error)

	// Runs a recursive query for a given record and record type
	RecursiveQuery(domain string, rrtype uint16) (Response, string, error)

	// Retrieves records from cache or an upstream
	RetrieveRecords(domain string, rrtype uint16) (Response, string, error)

	// Retrieve the server's outbound client
	GetDnsClient() Client

	// Retrieve the cache of locally hosted records
	GetHostedCache() *RecordCache

	// Add a upstream to the server's list
	AddUpstream(u *Upstream)

	// Get a copy of the connection pool for this server
	GetConnectionPool() ConnPool
}

func processResults(r dns.Msg, domain string, rrtype uint16) (Response, error) {
	return Response{
		Entry:        r,
		CreationTime: time.Now(),
		Key:          domain,
		Qtype:        rrtype,
	}, nil
}

func sendServfail(w ResponseWriter, duration time.Duration, r *dns.Msg) {
	LocalServfailsCounter.Inc()
	m := &dns.Msg{}
	m.SetRcode(r, dns.RcodeServerFailure)
	w.WriteMsg(m)
	logQuery("servfail", duration, m)
}

func logQuery(source string, duration time.Duration, response *dns.Msg) error {
	var queryContext LogContext
	for i, _ := range response.Question {
		for j, _ := range response.Answer {
			answerBits := strings.Split(response.Answer[j].String(), " ")
			queryContext = LogContext{
				"name":         response.Question[i].Name,
				"type":         dns.Type(response.Question[i].Qtype).String(),
				"opcode":       dns.OpcodeToString[response.Opcode],
				"answer":       answerBits[len(answerBits)-1],
				"answerSource": fmt.Sprintf("[%s]", source),
				"duration":     fmt.Sprintf("%s", duration),
			}
			QueryLogger.Log(LogMessage{
				Context: queryContext,
			})
		}
	}
	return nil
}

func sockoptSetter(network, address string, c syscall.RawConn) (err error) {
	config := GetConfiguration()
	err = c.Control(func(fd uintptr) {
		if config.UseTfo {
			if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_FASTOPEN_CONNECT, 1); err != nil {
				log.Printf("could not set TCP fast open to [%s]: %s", address, err.Error())
			}
		}
	})
	return
}

func buildDialer(timeout time.Duration) (dialer *net.Dialer) {
	return &net.Dialer{
		Control: sockoptSetter,
		Timeout: timeout,
	}
}
func BuildClient() (*dns.Client, error) {
	config := GetConfiguration()
	timeout := config.Timeout * time.Millisecond
	cl := &dns.Client{
		SingleInflight: true,
		Dialer:         buildDialer(timeout),
		Timeout:        timeout,
		Net:            "tcp-tls",
		TLSConfig: &tls.Config{
			InsecureSkipVerify: config.SkipUpstreamVerification,
		},
	}
	Logger.Log(LogMessage{
		Level: CRITICAL,
		Context: LogContext{
			"what": "instantiated new dns client in TLS mode",
			"next": "returning for use",
		},
	})
	return cl, nil
}

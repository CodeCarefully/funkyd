package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var (
	confFile   = flag.String("conf", "", "location of the funkyd configuration file")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
)

func validateConfFile() error {
	file := *confFile
	if _, err := os.Stat(file); err != nil {
		return err
	}

	if file == "" {
		return fmt.Errorf("no configuration file specified")
	}

	return nil
}

func validateFlags() error {
	if err := validateConfFile(); err != nil {
		return err
	}
	return nil
}

func runBlackholeServer(srv *dns.Server) error {
	config := GetConfiguration()
	switch config.ListenProtocol {
	case "tcp-tls":
		log.Printf("starting tls blackhole server")
		if (config.TlsConfig == tlsConfig{}) {
			log.Fatalf("attempted to listen for TLS connections, but no tls config was defined")
		}
		if config.TlsConfig.CertificateFile == "" {
			log.Fatalf("invalid certificate file in configuration")
		}

		if config.TlsConfig.PrivateKeyFile == "" {
			log.Fatalf("invalid private key in configuration")
		}

		cert, err := tls.LoadX509KeyPair(config.TlsConfig.CertificateFile, config.TlsConfig.PrivateKeyFile)
		if err != nil {
			log.Fatalf("could not load tls files")
		}

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		srv.Handler = &BlackholeServer{}
		return srv.ListenAndServe()
	default:
		return fmt.Errorf("unsupported protocol [%s]", config.ListenProtocol)
	}
}

func main() {
	flag.Parse()
	validateFlags()

	log.Printf("reading configuration from [%s]", *confFile)
	// read in configuration
	err := InitConfiguration(*confFile)
	if err != nil {
		log.Fatalf("could not open configuration: %s\n", err)
	}

	config := GetConfiguration()

	InitLoggers()
	InitApi()

	server, err := NewMutexServer(nil)
	if err != nil {
		log.Fatalf("could not initialize new server: %s\n", err)
	}

	// read in zone files, if configured to do so
	for _, file := range config.ZoneFiles {
		file, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("could not read zone file [%s]: %s\n", file, err)
		}
		responses, err := ParseZoneFile(string(file))
		if err != nil {
			log.Fatalf("could not parse zone file [%s]: %s\n", file, err)
		}
		for _, response := range responses {
			log.Printf("adding [%v]\n", response)
			// TODO one function to make the keys, please
			server.GetHostedCache().Add(response)
		}
	}

	// set up DNS server
	protocol := config.ListenProtocol
	if protocol == "" {
		protocol = "udp"
	}
	srv := &dns.Server{Addr: ":" + strconv.Itoa(config.DnsPort), Net: protocol, MaxTCPQueries: -1, ReusePort: true}
	srv.Handler = server
	if config.Blackhole {
		// PSYCH!
		err := runBlackholeServer(srv)
		if err != nil {
			log.Fatalf("Failed to run blackhole server: %s", err)
		}
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set %s listener %s\n", protocol, err.Error())
	}
}

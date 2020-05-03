package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type logConfig struct {
	// Whether or not this log should be on
	Enabled bool `json:"enabled"`
	// What log level to use
	Level LogLevel `json:"level"`
	// Where the log file should live
	Location string `json:"location"`
	// Whether to always use the minimal format for logs, which may be harder to parse
	TrimFormat bool `json:"trim_format"`
}

type Configuration struct {
	// Location of zone files with local dns configuration
	ZoneFiles []string `json:"zone_files"`

	// Port to listen for DNS traffic on
	DnsPort int `json:"dns_port"`

	// List of upstream resolvers, overrides resolv.conf
	Resolvers []string `json:"resolvers"`

	// Port to expose admin API on
	HttpPort int `json:"http_port"`

	// Sets the maximum connections to keep in the connection pool per upstream resolver
	MaxConnsPerHost int `json:"max_conns_per_host"`

	// Server logging
	ServerLog logConfig `json:"server_log"`

	// Query logging
	QueryLog logConfig `json:"query_log"`
}

var configuration Configuration

func InitConfiguration(configpath string) error {
	file, _ := os.Open(configpath)
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	configuration = Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		return fmt.Errorf("error while loading configuration from JSON: %s\n", err)
	}

	configJSON, err := json.MarshalIndent(configuration, "", "    ")
	if err != nil {
		return fmt.Errorf("could not render configuration [%v] as JSON", configuration)
	}
	fmt.Printf("running configuration: %s\n", string(configJSON))
	return nil
}

func GetConfiguration() Configuration {
	return configuration
}

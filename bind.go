package main

import (
  "time"
  "strings"
  "fmt"
  "github.com/miekg/dns"
)

// lifted and modified from cli53 code for this
// parses out a slice of Records from the contents of a zone file
// zone: the full text of a bind file
func ParseZoneFile(zone string) ([]Response, error) {
  tokensch := dns.ParseZone(strings.NewReader(zone), ".", "")
  responses := make([]Response, 0)

  for token := range tokensch {
    if token.Error != nil {
      return nil, fmt.Errorf("token error: %s\n", token.Error)
    }
    // TODO the stuff we host sholdn't follow the same rules, check out the spec
   cachedMsg := dns.Msg {
      Answer: []dns.RR{token.RR},
   }
   response := Response {
      Key: token.RR.Header().Name,
      Entry: cachedMsg,
      Ttl: time.Duration(token.RR.Header().Ttl) * time.Second,
      CreationTime: time.Now(),
      Qtype: token.RR.Header().Rrtype,
    }
    responses = append(responses, response)
  }

  return responses, nil
}


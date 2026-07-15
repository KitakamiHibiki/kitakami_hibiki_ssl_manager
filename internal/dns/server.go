package dns

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

// ChallengeStore is the interface for looking up challenge values.
type ChallengeStore interface {
	GetDomainByHash(hash string) (string, error)
	GetChallengeKeyAuth(domain string) string
}

// Server is an authoritative DNS server for challenge subdomains.
// It handles <hash>.challenge.<host> TXT queries.
type Server struct {
	mu       sync.RWMutex
	store    ChallengeStore
	host     string
	server   *dns.Server
}

// NewServer creates a DNS server that answers for *.challenge.<host>.
func NewServer(store ChallengeStore, host string, addr string) *Server {
	s := &Server{
		store: store,
		host:  strings.TrimSuffix(host, ".") + ".",
	}

	mux := dns.NewServeMux()
	mux.HandleFunc(s.host, s.handleChallenge)
	mux.HandleFunc(".", s.handleDefault)

	s.server = &dns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: mux,
	}
	return s
}

// Start begins listening for DNS queries.
func (s *Server) Start() error {
	log.Printf("[dns] starting on %s, zone: %s", s.server.Addr, s.host)
	return s.server.ListenAndServe()
}

// Shutdown stops the DNS server.
func (s *Server) Shutdown() error {
	return s.server.Shutdown()
}

// handleChallenge handles queries for *.challenge.<host>.
func (s *Server) handleChallenge(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	q := r.Question[0]
	if q.Qtype != dns.TypeTXT {
		w.WriteMsg(m)
		return
	}

	// Parse query name: <md5hash>.challenge.<host>
	// e.g. 5ab0fa2e.challenge.ssl.example.com
	name := strings.TrimSuffix(q.Name, ".")
	challengeZone := ".challenge." + strings.TrimSuffix(s.host, ".")
	hashStr := strings.TrimSuffix(name, challengeZone)
	if hashStr == name || hashStr == "" {
		w.WriteMsg(m)
		return
	}

	domain, err := s.store.GetDomainByHash(hashStr)
	if err != nil {
		log.Printf("[dns] hash %s not found: %v", hashStr, err)
		w.WriteMsg(m)
		return
	}

	val := s.store.GetChallengeKeyAuth(domain)
	if val == "" {
		log.Printf("[dns] no challenge for domain %s", domain)
		w.WriteMsg(m)
		return
	}

	txt := dns.TXT{
		Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
		Txt: []string{val},
	}
	m.Answer = append(m.Answer, &txt)
	w.WriteMsg(m)
}

// handleDefault returns NXDOMAIN for unknown queries.
func (s *Server) handleDefault(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.SetRcode(r, dns.RcodeNameError)
	w.WriteMsg(m)
}

// ResolveTXT queries the local DNS server for a TXT record.
func ResolveTXT(addr, domain string) ([]string, error) {
	c := new(dns.Client)
	c.Net = "udp"

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	m.RecursionDesired = true

	// Ensure addr has port
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "53")
	}

	resp, _, err := c.Exchange(m, addr)
	if err != nil {
		return nil, fmt.Errorf("dns query: %w", err)
	}

	var results []string
	for _, ans := range resp.Answer {
		if txt, ok := ans.(*dns.TXT); ok {
			results = append(results, txt.Txt...)
		}
	}
	return results, nil
}

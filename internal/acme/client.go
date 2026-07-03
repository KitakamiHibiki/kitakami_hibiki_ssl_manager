package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-acme/lego/v4/certcrypto"
)

var ManualDNS = NewManualProvider()

type Client struct {
	legoClient *lego.Client
	certDir    string
}

type user struct {
	email      string
	privateKey *rsa.PrivateKey
}

func (u *user) GetEmail() string                        { return u.email }
func (u *user) GetRegistration() *registration.Resource { return nil }
func (u *user) GetPrivateKey() crypto.PrivateKey        { return u.privateKey }

func NewClient(email, directory, certDir string) (*Client, error) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	config := lego.NewConfig(&user{email: email, privateKey: privateKey})
	config.CADirURL = directory
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	_, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	return &Client{legoClient: client, certDir: certDir}, nil
}

func (c *Client) SetDNSProvider(nameservers []string) {
	c.legoClient.Challenge.SetDNS01Provider(
		ManualDNS,
		dns01.AddDNSTimeout(10*time.Minute),
		dns01.AddRecursiveNameservers(nameservers),
	)
}

// Obtain requests a certificate for the given domain.
func (c *Client) Obtain(domain string) (*certificate.Resource, error) {
	return c.legoClient.Certificate.Obtain(certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	})
}

// ParseExpiry extracts the NotAfter time from a PEM-encoded certificate.
func ParseExpiry(certPEM []byte) (time.Time, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to decode PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

// CleanError extracts the human-readable portion from an ACME error string.
// Lego errors follow the pattern "context: acme: error: <code> :: ... :: <detail>".
func CleanError(err error) string {
	msg := err.Error()
	if idx := strings.LastIndex(msg, ":: "); idx != -1 {
		return strings.TrimSpace(msg[idx+3:])
	}
	return msg
}

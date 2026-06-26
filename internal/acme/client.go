package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type Client struct {
	user    *acmeUser
	lego    *lego.Client
	certDir string
}

type acmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *acmeUser) GetEmail() string                       { return u.Email }
func (u *acmeUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *acmeUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

func NewClient(email, dirURL, certDir string) (*Client, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	user := &acmeUser{Email: email, key: privateKey}

	cfg := lego.NewConfig(user)
	cfg.CADirURL = dirURL
	cfg.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{user: user, lego: client, certDir: certDir}, nil
}

func (c *Client) SetHTTPProvider(provider challenge.Provider) error {
	return c.lego.Challenge.SetHTTP01Provider(provider)
}

func (c *Client) SetDNSProvider(provider challenge.Provider) error {
	return c.lego.Challenge.SetDNS01Provider(provider)
}

func (c *Client) Obtain(domain string) (*certificate.Resource, error) {
	reg, err := c.lego.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	c.user.Registration = reg

	req := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	cert, err := c.lego.Certificate.Obtain(req)
	if err != nil {
		return nil, err
	}

	return cert, c.saveCert(domain, cert)
}

func (c *Client) Renew(domain string) (*certificate.Resource, error) {
	cert, err := c.lego.Certificate.Renew(certificate.Resource{
		Domain: domain,
	}, true, false, "")
	if err != nil {
		return nil, err
	}
	return cert, c.saveCert(domain, cert)
}

func (c *Client) saveCert(domain string, cert *certificate.Resource) error {
	if err := os.MkdirAll(filepath.Join(c.certDir, domain), 0755); err != nil {
		return err
	}

	files := map[string]string{
		"fullchain.pem": string(cert.Certificate),
		"privkey.pem":   string(cert.PrivateKey),
	}

	if len(cert.IssuerCertificate) > 0 {
		files["issuer.pem"] = string(cert.IssuerCertificate)
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(c.certDir, domain, name), []byte(content), 0600); err != nil {
			return err
		}
	}
	return nil
}

func ParseExpiry(certPEM []byte) (time.Time, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/platform"
)

type NginxDeployer struct {
	plat platform.Platform
}

func NewNginxDeployer(plat platform.Platform) *NginxDeployer {
	return &NginxDeployer{plat: plat}
}

func (n *NginxDeployer) Deploy(domain string, certPath string) error {
	confDir := n.plat.NginxConfDir()
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return err
	}

	fullchainPath := filepath.Join(certPath, "fullchain.pem")
	privkeyPath := filepath.Join(certPath, "privkey.pem")

	certData, err := os.ReadFile(fullchainPath)
	if err != nil {
		return err
	}
	keyData, err := os.ReadFile(privkeyPath)
	if err != nil {
		return err
	}

	targetCrt := filepath.Join(confDir, domain+".crt")
	targetKey := filepath.Join(confDir, domain+".key")

	if err := os.WriteFile(targetCrt, certData, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(targetKey, keyData, 0600); err != nil {
		return err
	}

	if err := n.plat.ReloadNginx(); err != nil {
		return fmt.Errorf("nginx reload failed: %w", err)
	}

	return nil
}

package deploy

import (
	"os"
	"path/filepath"
)

type LocalDeployer struct {
	targetDir string
}

func NewLocalDeployer(targetDir string) *LocalDeployer {
	return &LocalDeployer{targetDir: targetDir}
}

func (l *LocalDeployer) Deploy(domain string, certPath string) error {
	target := filepath.Join(l.targetDir, domain)
	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(certPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(certPath, entry.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(target, entry.Name()), data, 0644); err != nil {
			return err
		}
	}

	return nil
}

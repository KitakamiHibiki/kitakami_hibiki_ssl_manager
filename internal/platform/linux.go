package platform

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Linux struct{}

func (l *Linux) CertDir() string {
	return "/etc/ssl/kitakami"
}

func (l *Linux) NginxConfDir() string {
	return "/etc/nginx/conf.d"
}

func (l *Linux) ReloadNginx() error {
	return exec.Command("systemctl", "reload", "nginx").Run()
}

func (l *Linux) DataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kitakami")
}

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Windows struct{}

func (w *Windows) CertDir() string {
	return filepath.Join(os.Getenv("ProgramData"), "kitakami", "certs")
}

func (w *Windows) NginxConfDir() string {
	return filepath.Join("C:", "nginx", "conf", "conf.d")
}

func (w *Windows) ReloadNginx() error {
	return exec.Command("nginx", "-s", "reload").Run()
}

func (w *Windows) DataDir() string {
	return filepath.Join(os.Getenv("APPDATA"), "kitakami")
}

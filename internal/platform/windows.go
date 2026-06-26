package platform

import (
	"os/exec"
	"path/filepath"
)

type Windows struct{}

func (w *Windows) NginxConfDir() string {
	return filepath.Join("C:", "nginx", "conf", "conf.d")
}

func (w *Windows) ReloadNginx() error {
	return exec.Command("nginx", "-s", "reload").Run()
}

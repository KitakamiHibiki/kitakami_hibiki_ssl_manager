package platform

import "os/exec"

type Linux struct{}

func (l *Linux) NginxConfDir() string {
	return "/etc/nginx/conf.d"
}

func (l *Linux) ReloadNginx() error {
	return exec.Command("systemctl", "reload", "nginx").Run()
}

package platform

import "runtime"

type Platform interface {
	NginxConfDir() string
	ReloadNginx() error
}

func Detect() Platform {
	switch runtime.GOOS {
	case "windows":
		return &Windows{}
	default:
		return &Linux{}
	}
}

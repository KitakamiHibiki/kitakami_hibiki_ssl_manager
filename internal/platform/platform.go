package platform

import "runtime"

type Platform interface {
	CertDir() string
	NginxConfDir() string
	ReloadNginx() error
	DataDir() string
}

func Detect() Platform {
	switch runtime.GOOS {
	case "windows":
		return &Windows{}
	default:
		return &Linux{}
	}
}

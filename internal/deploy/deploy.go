package deploy

type Deployer interface {
	Deploy(domain string, certPath string) error
}

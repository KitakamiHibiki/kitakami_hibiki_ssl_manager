package deploy

import (
	"log"
	"time"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type Deployer struct {
	db      *store.DB
	certDir string
}

func NewDeployer(db *store.DB, certDir string) *Deployer {
	return &Deployer{db: db, certDir: certDir}
}

func (d *Deployer) DeployCert(certID uint) {
	cert, err := d.db.GetCertificateByID(certID)
	if err != nil {
		log.Printf("[deploy] cert %d not found: %v", certID, err)
		return
	}

	dom, err := d.db.GetDomain(cert.DomainID)
	if err != nil {
		log.Printf("[deploy] domain %d not found: %v", cert.DomainID, err)
		return
	}

	node, err := d.db.ListNodeByID(dom.DeployNodeID)
	if err != nil {
		log.Printf("[deploy] node %d not found: %v", dom.DeployNodeID, err)
		return
	}

	dl := &store.DeployLog{
		CertID:    certID,
		DomainID:  cert.DomainID,
		NodeID:    node.ID,
		NodeName:  node.Name,
		Status:    "pending",
		StartedAt: time.Now(),
	}
	d.db.CreateDeployLog(dl)

	detail, err := Deploy(node, d.certDir, dom)
	dl.FinishedAt = time.Now()
	if err != nil {
		dl.Status = "failed"
		dl.ErrorMsg = err.Error()
		log.Printf("[deploy] cert %d to node %s failed: %v", certID, node.Name, err)
	} else {
		dl.Status = "success"
		dl.Detail = detail
		log.Printf("[deploy] cert %d deployed to %s", certID, node.Name)
	}
	d.db.UpdateDeployLog(dl)
}

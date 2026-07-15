package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

func parseKey(raw []byte) (ssh.Signer, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty key")
	}
	signer, err := ssh.ParsePrivateKey(raw)
	if err == nil {
		return signer, nil
	}
	key, keyErr := ssh.ParseRawPrivateKey(raw)
	if keyErr != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return ssh.NewSignerFromKey(key)
}

func connect(node *store.DeployNode) (*ssh.Client, error) {
	var authMethod ssh.AuthMethod
	if node.AuthType == "key" {
		signer, err := parseKey([]byte(node.SSHKey))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(node.Password)
	}

	config := &ssh.ClientConfig{
		User:            node.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", node.Host, node.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func uploadFile(client *ssh.Client, localPath, remotePath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read local file: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	if err := session.Start("cat > '" + remotePath + "'"); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	if _, err := stdin.Write(data); err != nil {
		return fmt.Errorf("write stdin: %w", err)
	}
	stdin.Close()

	if err := session.Wait(); err != nil {
		return fmt.Errorf("cat: %w", err)
	}
	return nil
}

func runCmd(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("run command: %w", err)
	}
	return string(output), nil
}

// scpUpload uses the system scp command to upload a file.
func scpUpload(localPath, user, host string, port int, keyPath, remotePath string) error {
	args := []string{
		"-P", strconv.Itoa(port),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}
	if keyPath != "" {
		args = append(args, "-i", keyPath)
	}
	args = append(args, localPath, user+"@"+host+":"+remotePath)

	cmd := exec.Command("scp", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Deploy uploads certificate files to the remote node and runs the reload command.
func Deploy(node *store.DeployNode, certDir string, cert *store.Certificate) (string, error) {
	fullchain := filepath.Join(certDir, cert.Domain, "fullchain.pem")
	privkey := filepath.Join(certDir, cert.Domain, "privkey.pem")
	remoteCert := path.Join(cert.CertPath, cert.CertName)
	remoteKey := path.Join(cert.KeyPath, cert.KeyName)

	// mkdir first
	client, err := connect(node)
	if err != nil {
		return "", fmt.Errorf("ssh connect: %w", err)
	}
	if _, err := runCmd(client, "mkdir -p "+cert.CertPath+" "+cert.KeyPath); err != nil {
		client.Close()
		return "", fmt.Errorf("mkdir: %w", err)
	}
	client.Close()

	if node.AuthType == "key" && node.SSHKey != "" {
		// Use system scp for key auth
		tmp, err := os.CreateTemp("", "deploy-key-*")
		if err != nil {
			return "", fmt.Errorf("temp key: %w", err)
		}
		keyPath := tmp.Name()
		defer os.Remove(keyPath)

		if _, err := tmp.WriteString(node.SSHKey); err != nil {
			tmp.Close()
			return "", fmt.Errorf("write key: %w", err)
		}
		tmp.Chmod(0600)
		tmp.Close()

		if err := scpUpload(fullchain, node.Username, node.Host, node.Port, keyPath, remoteCert); err != nil {
			return "", fmt.Errorf("upload cert: %w", err)
		}
		if err := scpUpload(privkey, node.Username, node.Host, node.Port, keyPath, remoteKey); err != nil {
			return "", fmt.Errorf("upload key: %w", err)
		}
	} else {
		// Password auth fallback: upload via SSH session
		client2, err := connect(node)
		if err != nil {
			return "", fmt.Errorf("ssh connect: %w", err)
		}
		defer client2.Close()

		if err := uploadFile(client2, fullchain, remoteCert); err != nil {
			return "", fmt.Errorf("upload cert: %w", err)
		}
		if err := uploadFile(client2, privkey, remoteKey); err != nil {
			return "", fmt.Errorf("upload key: %w", err)
		}
	}

	// nginx reload
	client3, err := connect(node)
	if err != nil {
		// files uploaded but reload failed — non-critical
		detail := fmt.Sprintf("上传证书 → %s:%s\n上传私钥 → %s:%s",
			node.Host, remoteCert, node.Host, remoteKey)
		detail += "\nnginx reload: 连接失败"
		return detail, nil
	}
	defer client3.Close()

	reloadOut, _ := runCmd(client3, "sh -c 'nginx -s reload 2>&1 || true'")

	detail := fmt.Sprintf("上传证书 → %s:%s\n上传私钥 → %s:%s",
		node.Host, remoteCert, node.Host, remoteKey)
	if strings.TrimSpace(reloadOut) != "" {
		detail += "\nnginx reload: " + reloadOut
	}
	return detail, nil
}

// TestConnection tries to connect and authenticate, returns nil on success.
func TestConnection(node *store.DeployNode) error {
	client, err := connect(node)
	if err != nil {
		return err
	}
	client.Close()
	return nil
}

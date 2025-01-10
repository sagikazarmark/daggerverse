// OpenSSH server module for testing SSH connections.
package main

import (
	"dagger/openssh-server/internal/dagger"
	"fmt"
	"strings"
)

type OpensshServer struct {
	Port int

	// +private
	Container *dagger.Container
}

func New(
	// Custom container to use as a base container. OpenSSH server MUST be installed.
	//
	// +optional
	container *dagger.Container,
) *OpensshServer {
	if container == nil {
		container = dag.Apko().
			Wolfi().
			WithPackages([]string{"openssh-server"}).
			Container()
	}

	// Cleanup
	container = container.
		WithWorkdir("/").
		WithUser("root").
		WithDirectory("/var/empty", dag.Directory()). // Prevent "Missing privilege separation directory" error
		WithDirectory(
			"/etc/ssh",
			dag.Directory().
				WithFile("", dag.CurrentModule().Source().File("etc/sshd_config")).
				WithDirectory("sshd_config.d", dag.Directory()),
		).                                     // Clear existing SSH configuration
		WithExec([]string{"ssh-keygen", "-A"}) // Generate host keys

	return &OpensshServer{
		Port:      22,
		Container: container,
	}
}

// Mount a custom SSH configuration file (with .conf extension).
func (m *OpensshServer) WithConfig(name string, file *dagger.File) *OpensshServer {
	return &OpensshServer{
		Port:      m.Port,
		Container: m.Container.WithFile(fmt.Sprintf("/etc/ssh/sshd_config.d/%s.conf", name), file),
	}
}

// Returns the SSH host keys.
func (m *OpensshServer) HostKeys() *dagger.Directory {
	sshDir := m.Container.Directory("/etc/ssh")

	return dag.Directory().WithFiles("", []*dagger.File{sshDir.File("ssh_host_ecdsa_key.pub"), sshDir.File("ssh_host_ed25519_key.pub"), sshDir.File("ssh_host_rsa_key.pub")})
}

// Return a formatted SSH known_hosts file.
func (m *OpensshServer) KnownHosts(host string) *dagger.File {
	return m.Container.
		WithWorkdir("/etc/ssh").
		WithExec([]string{"sh", "-c", fmt.Sprintf("echo %s $(cat ssh_host_ecdsa_key.pub) > /known_hosts", host)}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("echo %s $(cat ssh_host_ed25519_key.pub) >> /known_hosts", host)}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("echo %s $(cat ssh_host_rsa_key.pub) >> /known_hosts", host)}).
		File("/known_hosts")
}

// Authorize a public key.
// By default, the key is authorized for the root user.
func (m *OpensshServer) WithAuthorizedKey(
	publicKey *dagger.File,

	// Authorize the key for this user.
	//
	// +optional
	user string,
) *OpensshServer {
	user = strings.ToLower(user)

	return &OpensshServer{
		Port: m.Port,
		Container: m.Container.
			With(func(c *dagger.Container) *dagger.Container {
				if user != "" && user != "root" {
					c = c.WithUser(user)
				}

				return c
			}).
			WithExec([]string{"sh", "-c", "mkdir -p ~/.ssh"}).
			WithMountedFile("/tmp/public-key", publicKey).
			WithExec([]string{"sh", "-c", "cat /tmp/public-key >> ~/.ssh/authorized_keys"}).
			WithoutFile("/tmp/public-key").
			WithExec([]string{"sh", "-c", "chmod 600 ~/.ssh/authorized_keys"}).
			With(func(c *dagger.Container) *dagger.Container {
				if user != "" && user != "root" {
					c = c.WithUser("root")
				}

				return c
			}),
	}
}

// Set the port number for the OpenSSH server.
func (m *OpensshServer) WithPort(port int) (*OpensshServer, error) {
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port number \"%d\": port number must be between 1 and 65535", port)
	}

	return &OpensshServer{
		Port:      port,
		Container: m.Container,
	}, nil
}

// Return a service that runs the OpenSSH server.
func (m *OpensshServer) Service() *dagger.Service {
	return m.Container.
		WithDefaultArgs([]string{"/usr/sbin/sshd", "-D", "-e", "-p", fmt.Sprintf("%d", m.Port)}).
		WithExposedPort(m.Port).
		AsService()
}

// OpenSSH server module for testing SSH connections.
package main

import (
	"fmt"
	"strings"
)

type OpensshServer struct {
	Container *Container
}

func New(
	// Custom container to use as a base container. OpenSSH server MUST be installed.
	//
	// +optional
	container *Container,
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
		Container: container,
	}
}

// Mount a custom SSH configuration file (with .conf extension).
func (m *OpensshServer) WithConfig(name string, file *File) *OpensshServer {
	return &OpensshServer{
		Container: m.Container.WithFile(fmt.Sprintf("/etc/ssh/sshd_config.d/%s.conf", name), file),
	}
}

// Returns the SSH host keys.
func (m *OpensshServer) HostKeys() *Directory {
	sshDir := m.Container.Directory("/etc/ssh")

	return dag.Directory().WithFiles("", []*File{sshDir.File("ssh_host_ecdsa_key.pub"), sshDir.File("ssh_host_ed25519_key.pub"), sshDir.File("ssh_host_rsa_key.pub")})
}

// Return a formatted SSH known_hosts file.
func (m *OpensshServer) KnownHosts(host string) *File {
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
	publicKey *File,

	// Authorize the key for this user.
	//
	// +optional
	user string,
) *OpensshServer {
	user = strings.ToLower(user)

	return &OpensshServer{
		Container: m.Container.
			With(func(c *Container) *Container {
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
			With(func(c *Container) *Container {
				if user != "" && user != "root" {
					c = c.WithUser("root")
				}

				return c
			}),
	}
}

// Return a service that runs the OpenSSH server.
func (m *OpensshServer) Service(
	// +optional
	// +default=22
	port int,
) *Service {
	return m.Container.
		WithExec([]string{"/usr/sbin/sshd", "-D", "-e", "-p", fmt.Sprintf("%d", port)}).
		WithExposedPort(port).
		AsService()
}

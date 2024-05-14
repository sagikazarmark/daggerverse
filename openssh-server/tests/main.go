package main

import (
	"context"
	"time"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Basic)
	p.Go(m.CustomPort)
	p.Go(m.User)
	p.Go(m.Config)

	return p.Wait()
}

func (m *Tests) Basic(ctx context.Context) error {
	publicKey, privateKey := keys()

	server := dag.OpensshServer().WithAuthorizedKey(publicKey)

	_, err := client(server, privateKey).
		WithExec([]string{"ssh", "-vvv", "-T", "root@server"}).
		Sync(ctx)

	return err
}

func (m *Tests) CustomPort(ctx context.Context) error {
	publicKey, privateKey := keys()

	server := dag.OpensshServer().WithAuthorizedKey(publicKey)

	_, err := clientWithPort(server, 2222, privateKey).
		WithExec([]string{"ssh", "-vvv", "-p", "2222", "-T", "root@server"}).
		Sync(ctx)

	return err
}

func (m *Tests) User(ctx context.Context) error {
	publicKey, privateKey := keys()

	server := dag.OpensshServer(OpensshServerOpts{
		Container: dag.Apko().Config(dag.CurrentModule().Source().File("testdata/git.apko.yaml")).Container(),
	}).WithAuthorizedKey(publicKey, OpensshServerWithAuthorizedKeyOpts{User: "git"})

	_, err := client(server, privateKey).
		WithExec([]string{"ssh", "-vvv", "-T", "git@server"}).
		Sync(ctx)

	return err
}

func (m *Tests) Config(ctx context.Context) error {
	publicKey, privateKey := keys()

	config := dag.CurrentModule().Source().File("testdata/custom.conf")

	server := dag.OpensshServer().WithAuthorizedKey(publicKey).WithConfig("custom", config)

	_, err := client(server, privateKey).
		WithExec([]string{"ssh", "-vvv", "-T", "root@server"}).
		Sync(ctx)

	return err
}

func keys() (*File, *File) {
	keygen := dag.Apko().Wolfi().WithPackage("openssh-keygen").Container().
		WithExec([]string{"mkdir", "-p", "/ssh"}).
		WithExec([]string{"ssh-keygen", "-q", "-N", "", "-t", "ed25519", "-f", "/ssh/id_ed25519"})

	publicKey := keygen.File("/ssh/id_ed25519.pub")
	privateKey := keygen.File("/ssh/id_ed25519")

	return publicKey, privateKey
}

func client(server *OpensshServer, privateKey *File) *Container {
	return clientWithPort(server, 0, privateKey)
}

func clientWithPort(server *OpensshServer, port int, privateKey *File) *Container {
	var serviceOpts OpensshServerServiceOpts

	if port > 0 {
		serviceOpts.Port = port
	}

	return dag.Apko().Wolfi().
		WithPackages([]string{"openssh-client"}).
		Container().
		WithServiceBinding("server", server.Service(serviceOpts)).
		WithMountedFile("/root/.ssh/known_hosts", server.KnownHosts("server")).
		WithMountedFile("/root/.ssh/id_ed25519", privateKey).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}

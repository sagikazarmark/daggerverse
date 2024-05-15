package main

import (
	"context"
	"time"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct {
	// +private
	PublicKey *File

	// +private
	PrivateKey *Secret
}

func New() *Tests {
	keyPair := dag.SSHKeygen().Ed25519().Generate()

	return &Tests{
		PublicKey:  keyPair.PublicKey(),
		PrivateKey: keyPair.PrivateKey(),
	}
}

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
	server := dag.OpensshServer().WithAuthorizedKey(m.PublicKey)

	_, err := client(server, m.PrivateKey).
		WithExec([]string{"ssh", "-vvv", "-T", "root@server"}).
		Sync(ctx)

	return err
}

func (m *Tests) CustomPort(ctx context.Context) error {
	server := dag.OpensshServer().WithAuthorizedKey(m.PublicKey).WithPort(2222)

	_, err := client(server, m.PrivateKey).
		WithExec([]string{"ssh", "-vvv", "-p", "2222", "-T", "root@server"}).
		Sync(ctx)

	return err
}

func (m *Tests) User(ctx context.Context) error {
	server := dag.OpensshServer(OpensshServerOpts{
		Container: dag.Apko().Config(dag.CurrentModule().Source().File("testdata/git.apko.yaml")).Container(),
	}).WithAuthorizedKey(m.PublicKey, OpensshServerWithAuthorizedKeyOpts{User: "git"})

	_, err := client(server, m.PrivateKey).
		WithExec([]string{"ssh", "-vvv", "-T", "git@server"}).
		Sync(ctx)

	return err
}

func (m *Tests) Config(ctx context.Context) error {
	config := dag.CurrentModule().Source().File("testdata/custom.conf")

	server := dag.OpensshServer().WithAuthorizedKey(m.PublicKey).WithConfig("custom", config)

	_, err := client(server, m.PrivateKey).
		WithExec([]string{"ssh", "-vvv", "-T", "root@server"}).
		Sync(ctx)

	return err
}

func client(server *OpensshServer, privateKey *Secret) *Container {
	return dag.Apko().Wolfi().
		WithPackages([]string{"openssh-client"}).
		Container().
		WithServiceBinding("server", server.Service()).
		WithMountedFile("/root/.ssh/known_hosts", server.KnownHosts("server")).
		WithMountedSecret("/root/.ssh/id_ed25519", privateKey).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}

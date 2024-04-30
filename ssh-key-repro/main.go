package main

import "context"

type SshKeyRepro struct{}

func (m *SshKeyRepro) TestOk() *Container {
	return dag.
		Wolfi().
		Container(WolfiContainerOpts{
			Packages: []string{"git", "openssh"},
		}).
		WithMountedFile("/ssh-key", dag.CurrentModule().Source().File("id_ed25519")).
		WithExec([]string{"ssh-keygen", "-y", "-f", "/ssh-key"})
}

func (m *SshKeyRepro) TestOkToo(ctx context.Context) *Container {
	sshKeyContents, _ := dag.CurrentModule().Source().File("id_ed25519").Contents(ctx)

	sshKey := dag.SetSecret("ssh-key", sshKeyContents)

	return dag.
		Wolfi().
		Container(WolfiContainerOpts{
			Packages: []string{"git", "openssh"},
		}).
		WithMountedSecret("/ssh-key", sshKey).
		WithExec([]string{"ssh-keygen", "-y", "-f", "/ssh-key"})
}

func (m *SshKeyRepro) TestFail(ctx context.Context, sshKey *Secret) *Container {
	// sshKeyContents, _ := dag.CurrentModule().Source().File("id_ed25519").Contents(ctx)
	//
	// sshKey := dag.SetSecret("ssh-key", sshKeyContents)

	return dag.
		Wolfi().
		Container(WolfiContainerOpts{
			Packages: []string{"git", "openssh"},
		}).
		WithMountedSecret("/ssh-key", sshKey).
		WithExec([]string{"ssh-keygen", "-y", "-f", "/ssh-key"})
}

func (m *SshKeyRepro) TestMaybeOk(ctx context.Context, sshKey *Secret) *Container {
	// sshKeyContents, _ := dag.CurrentModule().Source().File("id_ed25519").Contents(ctx)
	//
	// sshKey := dag.SetSecret("ssh-key", sshKeyContents)

	return dag.
		Wolfi().
		Container(WolfiContainerOpts{
			Packages: []string{"git", "openssh"},
		}).
		WithMountedSecret("/ssh-key", sshKey).
		WithExec([]string{"cp", "/ssh-key", "/ssh-key2"}).
		WithExec([]string{"sh", "-c", "echo '' >> /ssh-key2"}).
		WithExec([]string{"ssh-keygen", "-y", "-f", "/ssh-key2"})
}

package main

// Switch to unsafe mode to access the container directly.
func (m *Go) Unsafe() *Unsafe {
	return &Unsafe{m.Ctr}
}

type Unsafe struct {
	// +private
	Ctr *Container
}

func (m *Unsafe) Container() *Container {
	return m.Ctr
}

// Switch to back to safe mode to use the module's API.
func (m *Unsafe) Safe() *Go {
	return &Go{m.Ctr}
}

// Retrieves this container after executing the specified command inside it.
func (m *Unsafe) WithExec(
	args []string,

	// If the container has an entrypoint, ignore it for args rather than using it to wrap them.
	//
	// +optional
	skipEntrypoint bool,
	// Content to write to the command's standard input before closing (e.g., "Hello world").
	//
	// +optional
	stdin string,
	// Redirect the command's standard output to a file in the container (e.g., "/tmp/stdout").
	//
	// +optional
	redirectStdout string,
	// Redirect the command's standard error to a file in the container (e.g., "/tmp/stderr").
	//
	// +optional
	redirectStderr string,
	// Provides Dagger access to the executed command.
	//
	// Do not use this option unless you trust the command being executed; the command being executed WILL BE GRANTED FULL ACCESS TO YOUR HOST FILESYSTEM.
	//
	// +optional
	experimentalPrivilegedNesting bool,
	// Execute the command with all root capabilities. This is similar to running a command with "sudo" or executing "docker run" with the "--privileged" flag. Containerization does not provide any security guarantees when using this option. It should only be used when absolutely necessary and only with trusted commands.
	//
	// +optional
	insecureRootCapabilities bool,
) *Unsafe {
	return &Unsafe{
		m.Ctr.WithExec(args, ContainerWithExecOpts{
			SkipEntrypoint:                skipEntrypoint,
			Stdin:                         stdin,
			RedirectStdout:                redirectStdout,
			RedirectStderr:                redirectStderr,
			ExperimentalPrivilegedNesting: experimentalPrivilegedNesting,
			InsecureRootCapabilities:      insecureRootCapabilities,
		}),
	}
}

// Retrieves this container plus a directory mounted at the given path.
func (m *Unsafe) WithMountedDirectory(
	path string,
	source *Directory,

	// A user:group to set for the mounted directory and its contents.
	//
	// The user and group can either be an ID (1000:1000) or a name (foo:bar).
	//
	// If the group is omitted, it defaults to the same as the user.
	//
	// +optional
	owner string,
) *Unsafe {
	return &Unsafe{
		m.Ctr.WithMountedDirectory(
			path,
			source,
			ContainerWithMountedDirectoryOpts{
				Owner: owner,
			},
		),
	}
}

// Retrieves this container plus a file mounted at the given path.
func (m *Unsafe) WithMountedFile(
	path string,
	source *File,

	// A user or user:group to set for the mounted file.
	//
	// The user and group can either be an ID (1000:1000) or a name (foo:bar).
	//
	// If the group is omitted, it defaults to the same as the user.
	//
	// +optional
	owner string,
) *Unsafe {
	return &Unsafe{
		m.Ctr.WithMountedFile(
			path,
			source,
			ContainerWithMountedFileOpts{
				Owner: owner,
			},
		),
	}
}

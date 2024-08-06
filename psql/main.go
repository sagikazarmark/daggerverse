// psql is a terminal-based front-end to Postgres.
//
// It enables you to type in queries interactively, issue them to Postgres, and see the query results.
// Alternatively, input can be from a file.
// In addition, it provides a number of meta-commands and various shell-like features to facilitate writing scripts and automating a wide variety of tasks.

package main

import (
	"context"
	"dagger/psql/internal/dagger"
	"fmt"
	"time"

	"github.com/jszwec/csvutil"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "postgres"

type Psql struct {
	Container *dagger.Container
}

// This option determines whether or with what priority a secure SSL TCP/IP connection will be negotiated with the server.
type SSLMode string

const (
	// Only try a non-SSL connection.
	Disable SSLMode = "disable"

	// First try a non-SSL connection; if that fails, try an SSL connection.
	Allow SSLMode = "allow"

	// First try an SSL connection; if that fails, try a non-SSL connection.
	Prefer SSLMode = "prefer"

	// Only try an SSL connection. If a root CA file is present, verify the certificate in the same way as if verify-ca was specified.
	Require SSLMode = "require"

	// Only try an SSL connection, and verify that the server certificate is issued by a trusted certificate authority (CA).
	VerifyCA SSLMode = "verifyca"

	// Only try an SSL connection, verify that the server certificate is issued by a trusted CA and that the requested server host name matches that in the certificate.
	VerifyFull SSLMode = "verifyfull"
)

func New(
	// Name of host to connect to.
	//
	// +optional
	host string,

	// Service to connect to. Port needs to match the exposed port of the service.
	//
	// +optional
	service *dagger.Service,

	// Port number to connect to at the server host.
	//
	// +optional
	// +default=5432
	port int,

	// PostgreSQL user name to connect as.
	//
	// +optional
	// +default="postgres"
	user string,

	// Password to be used if the server demands password authentication.
	//
	// +optional
	password *dagger.Secret,

	// The database name.
	//
	// +optional
	database string,

	// This option determines whether or with what priority a secure SSL TCP/IP connection will be negotiated with the server.
	//
	// +optional
	sslmode string,

	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container. Takes precedence over version.
	//
	// +optional
	container *dagger.Container,
) (*Psql, error) {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	if host == "" && service == nil {
		return nil, fmt.Errorf("either host or service must be provided")
	}

	container = container.
		With(func(c *dagger.Container) *dagger.Container {
			if service != nil {
				if host == "" {
					host = "postgres"
				}

				c = c.WithServiceBinding(host, service)
			}

			return c
		}).
		WithEnvVariable("PGHOST", host).
		WithEnvVariable("PGPORT", fmt.Sprintf("%d", port)).
		WithEnvVariable("PGDATABASE", "postgres").
		With(func(c *dagger.Container) *dagger.Container {
			if user != "" {
				c = c.WithEnvVariable("PGUSER", user)
			}

			if password != nil {
				c = c.WithSecretVariable("PGPASSWORD", password)
			}

			if database != "" {
				c = c.WithEnvVariable("PGDATABASE", database)
			}

			if sslmode != "" {
				c = c.WithEnvVariable("PGSSLMODE", sslmode)
			}

			return c
		})

	return &Psql{
		Container: container,
	}, nil
}

// Open a psql terminal.
func (m *Psql) Terminal() *dagger.Container {
	return m.Container.
		Terminal(dagger.ContainerTerminalOpts{
			Cmd: []string{"psql"},
		})
}

type DatabaseListEntry struct {
	Name             string
	Owner            string
	Encoding         string
	LocaleProvider   string `csv:"Locale provider"`
	Collate          string
	Ctype            string
	ICULocale        string `csv:"ICU Locale"`
	ICURules         string `csv:"ICU Rules"`
	AccessPrivileges string `csv:"Access privileges"`
}

// List all available databases.
func (m *Psql) List(ctx context.Context) ([]DatabaseListEntry, error) {
	output, err := m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec([]string{"psql", "-l", "--csv"}).
		Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var databases []DatabaseListEntry

	if err := csvutil.Unmarshal([]byte(output), &databases); err != nil {
		return nil, err
	}

	return databases, nil
}

// Run a single command.
func (m *Psql) RunCommand(ctx context.Context, command string) (string, error) {
	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec([]string{"psql", "-c", command}).
		Stdout(ctx)
}

// Run a single command from a file.
func (m *Psql) RunFile(ctx context.Context, file *dagger.File) (string, error) {
	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithMountedFile("/work/command", file).
		WithExec([]string{"psql", "-f", "/work/command"}).
		Stdout(ctx)
}

// Run a series of commands.
func (m *Psql) Run() *Run {
	return nil
}

type Run struct {
	// +private
	Commands []Command

	// +private
	Container *dagger.Container
}

type Command struct {
	Kind bool

	Command string
	File    *dagger.File
}

// Add a command to the list of commands.
func (m *Run) WithCommand(command string) *Run {
	m.Commands = append(m.Commands, Command{
		Kind:    false,
		Command: command,
	})

	return m
}

// Add a command to the list of commands from a file.
func (m *Run) WithFile(file *dagger.File) *Run {
	m.Commands = append(m.Commands, Command{
		Kind: true,
		File: file,
	})

	return m
}

// Add a command to the list of commands from a file.
func (m *Run) Execute(ctx context.Context) (string, error) {
	container := m.Container

	args := []string{"psql"}

	var fileCount int

	for _, command := range m.Commands {
		if command.Kind {
			path := fmt.Sprintf("/work/commands/command-%d", fileCount)

			container = container.WithMountedFile(path, command.File)
			args = append(args, "-f", path)

			fileCount++
		} else {
			args = append(args, "-c", command.Command)
		}
	}

	return container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(args).
		Stdout(ctx)
}

// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"io"
	"path"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	Command           = "db-lifecycle"
	Migrate           = "migrate"
	Initialize        = "init"
	Replicate         = "replicate"
	Postgres          = cryptoutilSharedMagic.DockerServicePostgres
	SQLite            = "sqlite"
	InputName         = "lifecycle.json"
	SecretDirectory   = "/run/cryptoutil/secrets/"
	maximumInput      = 1 << 20
	commandTimeout    = 10 * time.Minute
	connectionTimeout = 10 * time.Second
)

// Connection contains protected file references, never inline credentials.
// DSN is a postgres:// URL containing only host/port/database, or an explicit
// SQLite DSN. PostgreSQL credentials and TLS material are separate files.
type Connection struct {
	DSNFile         string `json:"dsnFile"`
	UsernameFile    string `json:"usernameFile,omitempty"`
	PasswordFile    string `json:"passwordFile,omitempty"`
	CAFile          string `json:"caFile,omitempty"`
	CertificateFile string `json:"certificateFile,omitempty"`
	KeyFile         string `json:"keyFile,omitempty"`
}

// SubscriberTLS identifies paths accessible to the PostgreSQL subscriber
// SERVER, not paths in the Job. PostgreSQL stores connection information in its
// protected subscription catalog; this inherent persistence is not a temp file.
type SubscriberTLS struct {
	CA          string `json:"ca"`
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}

type Target struct {
	PSID          string        `json:"psid"`
	Database      string        `json:"database"`
	Role          string        `json:"role"`
	Publication   string        `json:"publication"`
	Subscription  string        `json:"subscription"`
	Slot          string        `json:"slot"`
	Leader        Connection    `json:"leader"`
	Follower      Connection    `json:"follower,omitempty"`
	SubscriberTLS SubscriberTLS `json:"subscriberTLS,omitempty"`
}

type Input struct {
	Backend       string     `json:"backend"`
	Targets       []Target   `json:"targets"`
	LeaderAdmin   Connection `json:"leaderAdmin,omitempty"`
	FollowerAdmin Connection `json:"followerAdmin,omitempty"`
}

func Decode(raw []byte) (Input, error) {
	var input Input
	if len(raw) == 0 || len(raw) > maximumInput {
		return input, errors.New("database input size rejected")
	}

	var unique map[string]any
	if err := yaml.Unmarshal(raw, &unique); err != nil {
		return input, errors.New("database input duplicate or malformed properties")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return Input{}, errors.New("database input rejected")
	}

	var tail any
	if err := decoder.Decode(&tail); !errors.Is(err, io.EOF) {
		return Input{}, errors.New("database input trailing content")
	}

	return input, nil
}

func protectedPath(value string) bool {
	if !strings.HasPrefix(value, SecretDirectory) || path.Clean(value) != value || strings.Contains(value, `\`) {
		return false
	}

	parts := strings.Split(strings.TrimPrefix(value, SecretDirectory), "/")

	return len(parts) == 2 && identifier(parts[0], true) && regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(parts[1]) && parts[1] != "." && parts[1] != ".."
}

func identifier(value string, dashes bool) bool {
	pattern := `^[a-z][a-z0-9_]{0,62}$`
	if dashes {
		pattern = `^[a-z][a-z0-9-]{0,62}$`
	}

	return regexp.MustCompile(pattern).MatchString(value)
}

func (c Connection) Validate(backend string) error {
	if !protectedPath(c.DSNFile) {
		return errors.New("database DSN requires protected file")
	}

	if backend == SQLite {
		if c.UsernameFile != "" || c.PasswordFile != "" || c.CAFile != "" || c.CertificateFile != "" || c.KeyFile != "" {
			return errors.New("SQLite forbids PostgreSQL material")
		}

		return nil
	}

	for _, file := range []string{c.UsernameFile, c.PasswordFile, c.CAFile, c.CertificateFile, c.KeyFile} {
		if !protectedPath(file) {
			return errors.New("PostgreSQL requires explicit protected credential and TLS files")
		}
	}

	return nil
}

// Validate is executable before any protected file or network reads.
func (input Input) Validate(operation string) error {
	if (operation != Migrate && operation != Initialize && operation != Replicate) ||
		(input.Backend != SQLite && input.Backend != Postgres) || len(input.Targets) < 1 || len(input.Targets) > 128 {
		return errors.New("database lifecycle operation/backend/targets rejected")
	}

	if input.Backend == SQLite && operation != Migrate {
		return errors.New("SQLite has no shared initialization or logical replication")
	}

	if operation == Migrate && len(input.Targets) != 1 {
		return errors.New("migration requires one isolated service")
	}

	seen := map[string]bool{}

	for _, target := range input.Targets {
		if _, err := Migrations(target.PSID); err != nil {
			return err
		}

		if !identifier(target.Database, false) || !identifier(target.Role, false) || seen[target.Database] {
			return errors.New("database target identity invalid or duplicate")
		}

		seen[target.Database] = true
		if err := target.Leader.Validate(input.Backend); err != nil {
			return err
		}

		if operation == Migrate {
			if target.Follower != (Connection{}) || target.SubscriberTLS != (SubscriberTLS{}) {
				return errors.New("migration cannot consume follower inputs")
			}

			continue
		}

		for _, name := range []string{target.Publication, target.Subscription, target.Slot} {
			if !identifier(name, false) {
				return errors.New("replication object identity rejected")
			}
		}

		if err := target.Follower.Validate(Postgres); err != nil {
			return err
		}

		if operation == Replicate {
			for _, file := range []string{target.SubscriberTLS.CA, target.SubscriberTLS.Certificate, target.SubscriberTLS.Key} {
				if !path.IsAbs(file) || path.Clean(file) != file || file == "/" || strings.ContainsAny(file, "\\\x00\r\n") {
					return errors.New("subscriber server TLS paths must be explicit absolute paths")
				}
			}
		} else if target.SubscriberTLS != (SubscriberTLS{}) {
			return errors.New("initialization does not consume subscriber transport")
		}
	}

	if operation == Initialize {
		if err := input.LeaderAdmin.Validate(Postgres); err != nil {
			return err
		}

		return input.FollowerAdmin.Validate(Postgres)
	}

	if input.LeaderAdmin != (Connection{}) || input.FollowerAdmin != (Connection{}) {
		return errors.New("non-initialization operation cannot consume administrator inputs")
	}

	return nil
}

// Files returns every explicitly requested Job-side protected input.
func (input Input) Files() []string {
	result := []string{}

	connections := make([]Connection, 0, 2+2*len(input.Targets))

	connections = append(connections, input.LeaderAdmin, input.FollowerAdmin)
	for _, target := range input.Targets {
		connections = append(connections, target.Leader, target.Follower)
	}

	for _, connection := range connections {
		for _, file := range []string{connection.DSNFile, connection.UsernameFile, connection.PasswordFile, connection.CAFile, connection.CertificateFile, connection.KeyFile} {
			if file != "" {
				result = append(result, file)
			}
		}
	}

	return result
}

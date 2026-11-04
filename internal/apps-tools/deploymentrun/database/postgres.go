// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite" // Production pure-Go SQLite driver.
)

type SQLClient struct {
	DB        *sql.DB
	sourceURL *url.URL
	backend   string
}

// OpenProtected never uses a credential environment variable or command argv.
// Job files are Kubernetes read-only projections of protected host imports.
func OpenProtected(read func(string) ([]byte, error)) Open {
	return openProtectedWith(read, sql.Open, func(config *pgx.ConnConfig) (*sql.DB, error) {
		return stdlib.OpenDB(*config), nil
	})
}

func openProtectedWith(read func(string) ([]byte, error), openSQLite func(string, string) (*sql.DB, error),
	openPostgres func(*pgx.ConnConfig) (*sql.DB, error),
) Open {
	return func(ctx context.Context, connection Connection, backend string) (Client, error) {
		if err := connection.Validate(backend); err != nil {
			return nil, err
		}

		data := map[string][]byte{}

		for _, file := range []string{connection.DSNFile, connection.UsernameFile, connection.PasswordFile, connection.CAFile, connection.CertificateFile, connection.KeyFile} {
			if file == "" {
				continue
			}

			raw, err := read(file)
			if err != nil || len(raw) == 0 || len(raw) > maximumInput {
				return nil, errors.New("protected database material read rejected")
			}

			data[file] = raw
		}

		dsn := strings.TrimSpace(string(data[connection.DSNFile]))
		if backend == SQLite {
			if !strings.HasPrefix(dsn, "file:") || strings.ContainsAny(dsn, "\x00\r\n") {
				return nil, errors.New("explicit SQLite file DSN required")
			}

			db, err := openSQLite(SQLite, dsn)
			if err != nil {
				return nil, errors.New("SQLite connection rejected")
			}

			db.SetMaxOpenConns(1)

			if err := db.PingContext(ctx); err != nil {
				_ = db.Close()

				return nil, errors.New("SQLite connection unavailable")
			}

			return &SQLClient{DB: db, backend: SQLite}, nil
		}

		config, source, err := postgresConfiguration(dsn, data[connection.UsernameFile], data[connection.PasswordFile],
			data[connection.CAFile], data[connection.CertificateFile], data[connection.KeyFile])
		if err != nil {
			return nil, err
		}

		db, err := openPostgres(config)
		if err != nil {
			return nil, errors.New("PostgreSQL connection rejected")
		}

		db.SetMaxOpenConns(1)

		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()

			return nil, errors.New("PostgreSQL connection unavailable")
		}

		return &SQLClient{DB: db, sourceURL: source, backend: Postgres}, nil
	}
}

func postgresConfiguration(dsn string, user, password, ca, cert, key []byte) (*pgx.ConnConfig, *url.URL, error) {
	return postgresConfigurationWithEnvironment(dsn, user, password, ca, cert, key, os.Getenv)
}

func postgresConfigurationWithEnvironment(dsn string, user, password, ca, cert, key []byte, lookup func(string) string) (*pgx.ConnConfig, *url.URL, error) {
	if err := explicitPostgresEnvironment(lookup); err != nil {
		return nil, nil, err
	}

	endpoint, err := url.Parse(dsn)
	if err != nil || endpoint.Scheme != Postgres || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" ||
		endpoint.Hostname() == "" || endpoint.Port() == "" || !identifier(strings.TrimPrefix(endpoint.Path, "/"), false) {
		return nil, nil, errors.New("PostgreSQL DSN must explicitly contain only host port and database")
	}

	port, err := strconv.ParseUint(endpoint.Port(), 10, 16)
	if err != nil || port == 0 {
		return nil, nil, errors.New("PostgreSQL port rejected")
	}

	username := strings.TrimSpace(string(user))

	pass := strings.TrimSuffix(strings.TrimSuffix(string(password), "\n"), "\r")
	if !identifier(username, false) || pass == "" || strings.ContainsAny(pass, "\x00\r\n") {
		return nil, nil, errors.New("PostgreSQL credential rejected")
	}

	roots, err := postgresTrustRoots(ca, time.Now().UTC())
	if err != nil {
		return nil, nil, errors.New("PostgreSQL server trust rejected")
	}

	certificates, err := postgresCertificates(cert)
	if err != nil || !singlePostgresPrivateKey(key) {
		return nil, nil, errors.New("PostgreSQL client encoding rejected")
	}

	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, nil, errors.New("PostgreSQL client keypair rejected")
	}

	leaf := certificates[0]
	if !postgresClientCertificate(leaf, time.Now().UTC()) {
		return nil, nil, errors.New("PostgreSQL client leaf rejected")
	}

	// Parse a fully explicit bootstrap URL to establish pgx's internal config
	// invariant. Replace every effective connection/runtime/TLS setting; ambient
	// PGHOST/PGUSER/PGPASSWORD/PGSSLMODE are never effective configuration.
	config, err := pgx.ParseConfig("postgres://explicit:explicit@127.0.0.1:5432/postgres?sslmode=disable&connect_timeout=10&target_session_attrs=any")
	if err != nil {
		return nil, nil, errors.New("PostgreSQL driver configuration rejected")
	}

	config.Host = endpoint.Hostname()
	config.Port = uint16(port)
	config.Database = strings.TrimPrefix(endpoint.Path, "/")
	config.User = username
	config.Password = pass
	config.RuntimeParams = map[string]string{"statement_timeout": "60000", "lock_timeout": "10000", "standard_conforming_strings": "on"}
	config.Fallbacks = nil
	config.ConnectTimeout = connectionTimeout
	config.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots, Certificates: []tls.Certificate{pair}, ServerName: endpoint.Hostname()}
	config.AfterConnect = nil
	config.OnNotice = nil
	config.OnNotification = nil
	config.ValidateConnect = nil
	config.Tracer = nil
	config.DialFunc = (&net.Dialer{Timeout: connectionTimeout}).DialContext
	config.LookupFunc = net.DefaultResolver.LookupHost
	source := *endpoint
	source.User = url.UserPassword(username, pass)

	return config, &source, nil
}

func (client *SQLClient) Close() error {
	if err := client.DB.Close(); err != nil {
		return errors.New("database closure rejected")
	}

	return nil
}

func (client *SQLClient) RequireDatabase(ctx context.Context, name, owner string) error {
	if client.backend == SQLite {
		return contextFailure(ctx)
	}

	var current, actualOwner string
	if err := client.DB.QueryRowContext(ctx, "SELECT current_database(),pg_get_userbyid(datdba) FROM pg_database WHERE datname=current_database()").Scan(&current, &actualOwner); err != nil || current != name || actualOwner != owner {
		return errors.New("database identity mismatch")
	}

	return nil
}

func (client *SQLClient) EnsureDatabase(ctx context.Context, name, role string) error {
	if !identifier(name, false) || !identifier(role, false) {
		return errors.New("database identity rejected")
	}

	var exists bool
	if err := client.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname=$1)", role).Scan(&exists); err != nil || !exists {
		return errors.New("explicit database owner role must already be provisioned")
	}

	var owner string

	err := client.DB.QueryRowContext(ctx, "SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname=$1", name).Scan(&owner)
	if err == nil {
		if owner != role {
			return errors.New("existing database owner collision")
		}

		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return errors.New("database ownership lookup rejected")
	}

	_, err = client.DB.ExecContext(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()+" OWNER "+pgx.Identifier{role}.Sanitize())
	if err != nil {
		return errors.New("database creation rejected")
	}

	return nil
}

func (client *SQLClient) publication(ctx context.Context, name string) (bool, error) {
	var all, owned bool

	err := client.DB.QueryRowContext(ctx, `SELECT (puballtables AND pubinsert AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot),
 pubowner=(SELECT oid FROM pg_roles WHERE rolname=current_user) FROM pg_publication WHERE pubname=$1`, name).Scan(&all, &owned)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil || !all || !owned {
		return false, errors.New("existing publication contract collision")
	}

	return true, nil
}

func (client *SQLClient) EnsurePublication(ctx context.Context, name string) error {
	if !identifier(name, false) {
		return errors.New("publication identity rejected")
	}

	exists, err := client.publication(ctx, name)
	if err != nil || exists {
		return err
	}

	if _, err := client.DB.ExecContext(ctx, "CREATE PUBLICATION "+pgx.Identifier{name}.Sanitize()+" FOR ALL TABLES"); err != nil {
		return errors.New("publication creation rejected")
	}

	return nil
}

func (client *SQLClient) RequirePublication(ctx context.Context, name string) error {
	exists, err := client.publication(ctx, name)
	if err != nil || !exists {
		return errors.New("source publication not ready")
	}

	return nil
}

func (client *SQLClient) Migrate(ctx context.Context, psid, backend string) error {
	return Apply(ctx, client.DB, psid, backend)
}

// Schema proves migration cursor and complete public table/column/constraint
// parity, rather than assuming legacy identity client tables exist in every PSID.
func (client *SQLClient) Schema(ctx context.Context) (string, error) {
	return replicaSchema(ctx, client.DB, true)
}

func replicaSchema(ctx context.Context, query replicaQueries, requireCursor bool) (string, error) {
	var schema string

	err := query.QueryRowContext(ctx, `SELECT jsonb_build_object(
 'migration',(SELECT jsonb_agg(to_jsonb(m)) FROM schema_migrations m WHERE NOT dirty),
 'columns',(SELECT jsonb_agg(to_jsonb(c) ORDER BY table_name,ordinal_position) FROM
 (SELECT table_name,column_name,ordinal_position,data_type,is_nullable,column_default
 FROM information_schema.columns WHERE table_schema='public') c),
 'constraints',(SELECT jsonb_agg(to_jsonb(c) ORDER BY table_name,definition) FROM
 (SELECT r.relname AS table_name,pg_get_constraintdef(c.oid) AS definition
 FROM pg_constraint c JOIN pg_class r ON r.oid=c.conrelid JOIN pg_namespace n ON n.oid=r.relnamespace
 WHERE n.nspname='public') c)
 )::text`).Scan(&schema)
	if err != nil {
		return "", errors.New("migrated schema inspection rejected")
	}

	if !requireCursor {
		return schema, nil
	}

	var dirty bool
	if err := query.QueryRowContext(ctx, "SELECT dirty FROM (SELECT count(*)<>1 OR COALESCE(bool_or(dirty),true) AS dirty FROM schema_migrations) cursor").Scan(&dirty); err != nil || dirty {
		return "", errors.New("migration cursor not clean")
	}

	return schema, nil
}

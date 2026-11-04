// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ProvisionRoles only creates ordinary, run-bound login roles. It does not
// grant replication, superuser, database creation, membership or publication
// authority. Existing owned roles are never password-reset or privilege-edited.
const ProvisionRoles = "provision"

const roleServiceTier = "s"

type RoleProvisionInput struct {
	Run      string `json:"run"`
	Tier     string `json:"tier"`
	Scope    string `json:"scope"`
	Variant  string `json:"variant"`
	Database Input  `json:"database"`
}

func (input RoleProvisionInput) Validate() error {
	if !identifier(input.Run, true) || !identifier(input.Scope, true) ||
		(input.Tier != roleServiceTier && input.Tier != "p" && input.Tier != "u") ||
		(input.Variant != "pg" && input.Variant != "mixed-mem" && input.Variant != "mixed-disk") {
		return errors.New("explicit run-bound PostgreSQL role scope required")
	}

	if err := input.Database.Validate(Initialize); err != nil {
		return errors.New("protected PostgreSQL role inputs rejected")
	}

	for _, target := range input.Database.Targets {
		// Exact owning graph.go sqlID derivation; parity tests must compare this
		// independently to graph release database/role identities.
		stem := strings.ReplaceAll(strings.Join([]string{input.Run, input.Tier, input.Scope, input.Variant, target.PSID}, "_"), "-", "_")
		if target.Database != "d_"+stem || target.Role != "u_"+stem || target.Publication != "pub_"+stem || target.Subscription != "sub_"+stem {
			return errors.New("role provisioning graph identity mismatch")
		}

		if target.Slot != "slot_"+stem || (input.Tier == roleServiceTier && input.Scope != target.PSID) {
			return errors.New("role provisioning graph scope mismatch")
		}
	}

	return nil
}

type ordinaryRoleClient interface {
	EnsureOrdinaryRole(context.Context, string, string, []byte) error
}

func ProvisionOrdinaryRoles(ctx context.Context, input RoleProvisionInput, open Open, read func(string) ([]byte, error)) (result error) {
	if ctx == nil || open == nil || read == nil {
		return errors.New("protected role provisioning dependencies required")
	}

	if err := input.Validate(); err != nil {
		return err
	}

	if err := contextFailure(ctx); err != nil {
		return err
	}

	passwords := map[string][]byte{}

	for _, target := range input.Database.Targets {
		for _, connection := range []Connection{target.Leader, target.Follower} {
			if err := contextFailure(ctx); err != nil {
				return err
			}

			username, err := read(connection.UsernameFile)
			if err != nil || len(username) > maximumInput || strings.TrimSpace(string(username)) != target.Role {
				return errors.New("protected role username binding rejected")
			}

			password, err := read(connection.PasswordFile)

			normalized := strings.TrimSuffix(strings.TrimSuffix(string(password), "\n"), "\r")
			if err != nil || len(password) > maximumInput || normalized == "" || strings.ContainsAny(normalized, "\x00\r\n") {
				return errors.New("protected role password unavailable")
			}

			passwords[connection.PasswordFile] = append([]byte(nil), password...)
		}
	}

	if err := contextFailure(ctx); err != nil {
		return err
	}

	clients := []Client{}
	defer func() {
		for _, client := range clients {
			if err := client.Close(); err != nil {
				result = errors.Join(result, errors.New("protected role connection closure rejected"))
			}
		}

		result = errors.Join(result, contextFailure(ctx))
	}()

	for _, side := range []struct {
		admin  Connection
		leader bool
	}{{input.Database.LeaderAdmin, true}, {input.Database.FollowerAdmin, false}} {
		client, err := open(ctx, side.admin, Postgres)
		if err != nil || client == nil {
			return errors.New("protected role administrator unavailable")
		}

		clients = append(clients, client)

		provisioner, ok := client.(ordinaryRoleClient)
		if !ok {
			return errors.New("ordinary role provisioner unavailable")
		}

		for _, target := range input.Database.Targets {
			connection := target.Follower
			if side.leader {
				connection = target.Leader
			}

			if err := provisioner.EnsureOrdinaryRole(ctx, input.Run, target.Role, passwords[connection.PasswordFile]); err != nil {
				return errors.New("ordinary role verify-or-create rejected")
			}
			// Authenticate through the existing administrator database, whose
			// existence does not depend on the later initialization Job. Keep
			// the application's explicit CA/certificate/key and credentials.
			connection.DSNFile = side.admin.DSNFile

			check, err := open(ctx, connection, Postgres)
			if err != nil || check == nil {
				return errors.New("protected provisioned role authentication rejected")
			}

			clients = append(clients, check)
		}
	}

	return nil
}

func (client *SQLClient) EnsureOrdinaryRole(ctx context.Context, run, role string, rawPassword []byte) (result error) {
	password := strings.TrimSuffix(strings.TrimSuffix(string(rawPassword), "\n"), "\r")
	if ctx == nil || client == nil || client.DB == nil || client.backend != Postgres ||
		!identifier(run, true) || !identifier(role, false) || !strings.HasPrefix(role, "u_"+strings.ReplaceAll(run, "-", "_")+"_") ||
		password == "" || len(password) > maximumInput || strings.ContainsAny(password, "\x00\r\n") {
		return errors.New("ordinary role protected identity rejected")
	}

	tx, err := client.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return errors.New("ordinary role transaction unavailable")
	}

	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", "cryptoutil-role/"+run+"/"+role); err != nil {
		return errors.New("ordinary role ownership lease rejected")
	}

	comment := "cu.io/run=" + run + ";cu.io/role=" + role
	query := `SELECT rolcanlogin AND NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole AND NOT rolreplication AND NOT rolbypassrls
 AND NOT rolinherit AND rolconnlimit=-1 AND rolvaliduntil IS NULL
 AND NOT EXISTS(SELECT 1 FROM pg_auth_members WHERE member=pg_roles.oid OR roleid=pg_roles.oid),
 COALESCE(shobj_description(oid,'pg_authid'),'') FROM pg_roles WHERE rolname=$1`

	var (
		ordinary  bool
		ownership string
	)

	err = tx.QueryRowContext(ctx, query, role).Scan(&ordinary, &ownership)
	if errors.Is(err, sql.ErrNoRows) {
		quoted := "'" + strings.ReplaceAll(password, "'", "''") + "'"
		if _, err := tx.ExecContext(ctx, "CREATE ROLE "+pgx.Identifier{role}.Sanitize()+
			" WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD "+quoted); err != nil {
			return errors.New("ordinary role creation rejected")
		}

		if _, err := tx.ExecContext(ctx, "COMMENT ON ROLE "+pgx.Identifier{role}.Sanitize()+" IS '"+comment+"'"); err != nil {
			return errors.New("ordinary role ownership publication rejected")
		}

		err = tx.QueryRowContext(ctx, query, role).Scan(&ordinary, &ownership)
	}

	if err != nil || !ordinary || ownership != comment {
		return errors.New("existing role ownership or privilege collision")
	}

	if err := tx.Commit(); err != nil {
		return errors.New("ordinary role transaction completion rejected")
	}

	return nil
}

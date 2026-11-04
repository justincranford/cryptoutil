// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"database/sql"
	json "encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func sqlLiteral(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }

func (client *SQLClient) EnsureSubscription(ctx context.Context, target Target, source Client) error {
	leader, ok := source.(*SQLClient)
	if !ok || leader.sourceURL == nil {
		return errors.New("explicit publisher transport required")
	}

	connection := *leader.sourceURL
	parameters := url.Values{
		"sslmode": {"verify-full"}, "sslrootcert": {target.SubscriberTLS.CA},
		"sslcert": {target.SubscriberTLS.Certificate}, "sslkey": {target.SubscriberTLS.Key},
		"connect_timeout":          {"10"},
		"ssl_min_protocol_version": {"TLSv1.3"},
	}
	connection.RawQuery = parameters.Encode()
	conninfo := connection.String()

	var (
		enabled, owned  bool
		existing, slot  string
		publicationsRaw string
		publications    []string
	)

	err := client.DB.QueryRowContext(ctx, `SELECT subenabled,subowner=(SELECT oid FROM pg_roles WHERE rolname=current_user),
 subconninfo,subslotname,array_to_json(subpublications)::text FROM pg_subscription
 WHERE subname=$1 AND subdbid=(SELECT oid FROM pg_database WHERE datname=current_database())`, target.Subscription).
		Scan(&enabled, &owned, &existing, &slot, &publicationsRaw)
	if err == nil {
		if err := json.Unmarshal([]byte(publicationsRaw), &publications); err != nil {
			return errors.New("subscription publication identity rejected")
		}

		if !enabled || !owned || existing != conninfo || slot != target.Slot || !slices.Equal(publications, []string{target.Publication}) {
			return errors.New("existing subscription contract collision")
		}
	} else {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.New("subscription lookup rejected")
		}

		if err := client.requireEmptyReplica(ctx); err != nil {
			return err
		}

		//nolint:gosec // G202: DDL cannot parameterize these fields. Identifiers use pgx quoting; literals double quotes with standard_conforming_strings=on.
		statement := "CREATE SUBSCRIPTION " + pgx.Identifier{target.Subscription}.Sanitize() +
			" CONNECTION " + sqlLiteral(conninfo) + " PUBLICATION " + pgx.Identifier{target.Publication}.Sanitize() +
			" WITH (copy_data=true, create_slot=true, enabled=true, slot_name=" + sqlLiteral(target.Slot) + ")"
		if _, err := client.DB.ExecContext(ctx, statement); err != nil {
			return errors.New("subscription creation rejected")
		}
	}

	for {
		var ready bool

		err := client.DB.QueryRowContext(ctx, `SELECT
 EXISTS(SELECT 1 FROM pg_stat_subscription st JOIN pg_subscription s ON s.oid=st.subid
 WHERE s.subname=$1 AND s.subdbid=(SELECT oid FROM pg_database WHERE datname=current_database()) AND st.pid IS NOT NULL)
 AND EXISTS(SELECT 1 FROM pg_subscription_rel r JOIN pg_subscription s ON s.oid=r.srsubid
 WHERE s.subname=$1 AND s.subdbid=(SELECT oid FROM pg_database WHERE datname=current_database()))
 AND NOT EXISTS(SELECT 1 FROM pg_subscription_rel r JOIN pg_subscription s ON s.oid=r.srsubid
 WHERE s.subname=$1 AND s.subdbid=(SELECT oid FROM pg_database WHERE datname=current_database()) AND r.srsubstate<>'r')`, target.Subscription).Scan(&ready)
		if err != nil {
			return errors.New("subscription readiness lookup rejected")
		}

		if ready {
			return nil
		}

		select {
		case <-ctx.Done():
			return contextFailure(ctx)
		case <-time.After(time.Second):
		}
	}
}

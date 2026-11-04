// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func ordinaryRoleFixture() RoleProvisionInput {
	input := RoleProvisionInput{Run: testOrdinaryRun, Tier: roleServiceTier, Scope: psidKMS, Variant: "pg", Database: inputFixture(Initialize)}
	stem := "p2sm1103_s_sm_kms_pg_sm_kms"
	target := &input.Database.Targets[0]
	target.Database, target.Role = "d_"+stem, "u_"+stem
	target.Publication, target.Subscription, target.Slot = "pub_"+stem, "sub_"+stem, "slot_"+stem

	return input
}

func TestSupportedOrdinaryRoleCommandRejectsForgedAndUnprotectedInputs(t *testing.T) {
	t.Parallel()

	for mode := range testRoleCase8 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			args := []string{"--input=" + SecretDirectory + "postgres-bootstrap/" + RoleProvisionInputName}
			input := ordinaryRoleFixture()

			if mode == 1 {
				args = []string{testOrdinaryHelp}
			}

			if mode == 2 {
				args = []string{"--input=unprotected"}
			}

			if mode == testRoleCase3 {
				input.Run = "another-run"
			}

			if mode == testRoleCase7 {
				input.Database.Targets[0].Slot = "another_slot"
			}

			raw, err := json.Marshal(input)
			require.NoError(t, err)

			if mode == testRoleCase4 {
				raw = append(raw, []byte(` {}`)...)
			}

			read := func(string) ([]byte, error) {
				if mode == testRoleCase5 {
					return nil, errors.New("PRIVATE")
				}

				return raw, nil
			}
			calls := 0
			provision := func(ctx context.Context, input RoleProvisionInput, open Open, read func(string) ([]byte, error)) error {
				calls++
				_, bounded := ctx.Deadline()
				require.True(t, bounded)
				require.NotNil(t, open)
				require.NotNil(t, read)
				require.NoError(t, input.Validate())

				if mode == testRoleCase6 {
					return errors.New("PRIVATE")
				}

				return nil
			}

			var stdout, stderr bytes.Buffer

			code := roleProvisionMain(args, &stdout, &stderr, read, provision)
			if mode < 2 {
				require.Zero(t, code)
			} else {
				require.Equal(t, 1, code)
			}

			if mode == 0 || mode == testRoleCase6 {
				require.Equal(t, 1, calls)
			} else {
				require.Zero(t, calls)
			}

			require.NotContains(t, stdout.String()+stderr.String(), "PRIVATE")
		})
	}
}

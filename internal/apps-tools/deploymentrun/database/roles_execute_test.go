// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roleExecutionClient struct {
	*fakeClient
	t *testing.T
}

func (client *roleExecutionClient) EnsureOrdinaryRole(_ context.Context, run, role string, password []byte) error {
	require.Equal(client.t, testOrdinaryRun, run)
	require.True(client.t, strings.HasPrefix(role, "u_p2sm1103_"))
	require.Equal(client.t, testOrdinaryPassword, string(password))

	return client.record(ProvisionRoles)
}

func TestProtectedOrdinaryRoleProvisioningOrdersBothAdministratorsAndAuthentication(t *testing.T) {
	t.Parallel()

	for mode := range 16 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			input := ordinaryRoleFixture()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			calls := []string{}
			opens := 0
			open := func(_ context.Context, connection Connection, backend string) (Client, error) {
				opens++

				require.Equal(t, Postgres, backend)

				if opens == 2 {
					require.Equal(t, input.Database.LeaderAdmin.DSNFile, connection.DSNFile)
					require.Equal(t, input.Database.Targets[0].Leader.CertificateFile, connection.CertificateFile)
				}

				if opens == testRoleCase4 {
					require.Equal(t, input.Database.FollowerAdmin.DSNFile, connection.DSNFile)
					require.Equal(t, input.Database.Targets[0].Follower.CertificateFile, connection.CertificateFile)
				}

				if mode == testRoleCase4 || (mode == testRoleCase5 && opens == 2) {
					return nil, errors.New("PRIVATE")
				}

				if mode == testRoleCase6 {
					return nil, nil
				}

				base := &fakeClient{name: strconv.Itoa(opens), calls: &calls}
				if mode == testRoleCase7 {
					return base, nil
				}

				if mode == testRoleCase10 {
					base.fail = "1:" + ProvisionRoles
				}

				if mode == testRoleCase11 {
					base.fail = "1:" + testClose
				}

				return &roleExecutionClient{fakeClient: base, t: t}, nil
			}
			read := func(file string) ([]byte, error) {
				if mode == testRoleCase8 {
					return nil, errors.New("PRIVATE")
				}

				if strings.HasSuffix(file, "/username") {
					if mode == testRoleCase9 {
						return []byte("not-the-graph-role"), nil
					}

					return []byte(input.Database.Targets[0].Role), nil
				}

				if mode == testRoleCase12 {
					return nil, errors.New("PRIVATE")
				}

				if mode == testRoleCase14 || (mode == testRoleCase15 && file == input.Database.Targets[0].Follower.PasswordFile) {
					cancel()
				}

				return []byte(testOrdinaryPassword), nil
			}

			switch mode {
			case 1:
				ctx = nil
			case 2:
				open = nil
			case testRoleCase3:
				read = nil
			case testRoleCase13:
				cancel()
			}

			err := ProvisionOrdinaryRoles(ctx, input, open, read)
			if mode == 0 {
				require.NoError(t, err)
				require.Equal(t, testRoleCase4, opens)
				require.Equal(t, []string{"1:" + ProvisionRoles, "3:" + ProvisionRoles, "1:" + testClose, "2:" + testClose, "3:" + testClose, "4:" + testClose}, calls)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")

				if mode == testRoleCase8 || mode == testRoleCase9 || mode == testRoleCase12 || mode == testRoleCase14 || mode == testRoleCase15 {
					require.Zero(t, opens, "invalid protected credentials must be rejected before any administrator connection or role mutation")
				}
			}
		})
	}
}

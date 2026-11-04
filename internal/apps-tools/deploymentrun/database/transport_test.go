// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func transportFixture(t *testing.T) ([]byte, []byte, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	now := time.Now().UTC()
	root := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "explicit-root"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	ca, err := x509.CreateCertificate(crand.Reader, root, root, &key.PublicKey, key)
	require.NoError(t, err)

	leaf := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "explicit-client"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	cert, err := x509.CreateCertificate(crand.Reader, leaf, root, &key.PublicKey, key)
	require.NoError(t, err)
	private, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: testCertificatePEMType, Bytes: ca}), pem.EncodeToMemory(&pem.Block{Type: testCertificatePEMType, Bytes: cert}),
		pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: private})
}

func TestExplicitPostgresTransport(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testValid, testBadURL, testHost, testInlineUser, testQuery, testZeroPort, testLargePort, testBadUser, testBadPassword, testBadCA, testBadPair, testBadLeaf, testWrongEKU} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			ca, cert, key := transportFixture(t)
			dsn, user, password := "postgres://leader.example:5432/d_sm_kms", []byte(testMigrationOwner), []byte("DO-NOT-LOG")

			switch test {
			case testBadURL:
				dsn = "://bad"
			case testHost:
				dsn = "postgres:///d_sm_kms"
			case testInlineUser:
				dsn = "postgres://inline@leader.example:5432/d_sm_kms"
			case testQuery:
				dsn += "?sslmode=disable"
			case testZeroPort:
				dsn = "postgres://leader.example:0/d_sm_kms"
			case testLargePort:
				dsn = "postgres://leader.example:65536/d_sm_kms"
			case testBadUser:
				user = []byte(testSQLInjection)
			case testBadPassword:
				password = []byte("secret\ninjection")
			case testBadCA:
				ca = []byte(testInvalid)
			case testBadPair:
				key = []byte(testInvalid)
			case testBadLeaf:
				cert = ca
			case testWrongEKU:
				privateBlock, _ := pem.Decode(key)
				private, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
				require.NoError(t, err)

				ecdsaKey, ok := private.(*ecdsa.PrivateKey)
				require.True(t, ok)

				now := time.Now().UTC()
				leaf := &x509.Certificate{SerialNumber: big.NewInt(3), NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
				der, err := x509.CreateCertificate(crand.Reader, leaf, leaf, &ecdsaKey.PublicKey, ecdsaKey)
				require.NoError(t, err)

				cert = pem.EncodeToMemory(&pem.Block{Type: testCertificatePEMType, Bytes: der})
			}

			config, source, err := postgresConfiguration(dsn, user, password, ca, cert, key)
			if test != testValid {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")

				return
			}

			require.NoError(t, err)
			require.Equal(t, "leader.example", config.Host)
			require.Equal(t, testMigrationOwner, config.User)
			require.Equal(t, testDatabase, config.Database)
			require.Equal(t, "leader.example", config.TLSConfig.ServerName)
			require.Equal(t, uint16(tls.VersionTLS13), config.TLSConfig.MinVersion)
			require.False(t, config.TLSConfig.InsecureSkipVerify)
			require.Empty(t, config.Fallbacks)
			require.Nil(t, config.Tracer)
			require.Equal(t, "on", config.RuntimeParams["standard_conforming_strings"])
			require.NotNil(t, source.User)
		})
	}
}

func TestProtectedDatabaseOpenerBoundaries(t *testing.T) {
	t.Parallel()

	for _, test := range []string{Postgres, SQLite, testInvalidFiles, testReadError, testEmptyFile, testOversizedFile, testSQLiteDSN, testSQLiteOpen, testSQLitePing, testPGConfigFailure, testPGOpen, testPGPing} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			ca, cert, key := transportFixture(t)
			connection := connectionFixture("owned")

			backend := Postgres
			if test == SQLite || test == testSQLiteDSN || test == testSQLiteOpen || test == testSQLitePing {
				backend = SQLite
				connection = Connection{DSNFile: SecretDirectory + "owned/dsn"}
			}

			if test == testInvalidFiles {
				connection.DSNFile = ""
			}

			read := func(file string) ([]byte, error) {
				if test == testReadError {
					return nil, errors.New("DO-NOT-LOG")
				}

				if test == testEmptyFile {
					return nil, nil
				}

				if test == testOversizedFile {
					return make([]byte, maximumInput+1), nil
				}

				switch file {
				case connection.DSNFile:
					if test == testPGConfigFailure || test == testSQLiteDSN {
						return []byte(testInvalid), nil
					}

					if backend == SQLite {
						return []byte("file:explicit?mode=memory&cache=shared"), nil
					}

					return []byte("postgres://leader.example:5432/d_sm_kms"), nil
				case connection.UsernameFile:
					return []byte(testMigrationOwner), nil
				case connection.PasswordFile:
					return []byte("DO-NOT-LOG"), nil
				case connection.CAFile:
					return ca, nil
				case connection.CertificateFile:
					return cert, nil
				default:
					return key, nil
				}
			}
			client, script := sqlFixture(t, nil)
			open := openProtectedWith(read, func(string, string) (*sql.DB, error) {
				if test == testSQLiteOpen {
					return nil, errors.New("DO-NOT-LOG")
				}

				if test == testSQLitePing {
					require.NoError(t, client.DB.Close())
				}

				return client.DB, nil
			}, func(config *pgx.ConnConfig) (*sql.DB, error) {
				require.NotNil(t, config.TLSConfig)

				if test == testPGOpen {
					return nil, errors.New("DO-NOT-LOG")
				}

				if test == testPGPing {
					require.NoError(t, client.DB.Close())
				}

				return client.DB, nil
			})

			got, err := open(context.Background(), connection, backend)
			if test == Postgres || test == SQLite {
				require.NoError(t, err)
				require.NoError(t, got.Close())
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
			}

			require.Empty(t, script.steps)
			require.NoError(t, client.DB.Close())
		})
	}

	require.NotNil(t, OpenProtected(func(string) ([]byte, error) { return nil, errors.New("injected") }))
}

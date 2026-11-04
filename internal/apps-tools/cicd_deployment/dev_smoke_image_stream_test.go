// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"archive/tar"
	"bytes"
	"context"
	sha256 "crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type developmentStreamFixtureProcess struct {
	startErr bool
	waitErr  bool
	input    io.Reader
	output   io.Writer
	archive  []byte
	done     chan error
	onWait   func()
}

func (process *developmentStreamFixtureProcess) Start() error {
	if process.startErr {
		return errors.New("PRIVATE")
	}

	process.done = make(chan error, 1)

	go func() {
		var err error
		if process.input != nil {
			_, err = io.Copy(io.Discard, process.input)
		} else {
			_, err = io.Copy(process.output, bytes.NewReader(process.archive))
		}

		process.done <- err
	}()

	return nil
}

func (process *developmentStreamFixtureProcess) Wait() error {
	err := <-process.done
	if process.onWait != nil {
		process.onWait()
	}

	if process.waitErr {
		return errors.New("PRIVATE")
	}

	return err
}

func developmentStreamArchive(t *testing.T, large bool) ([]byte, string) {
	t.Helper()

	index := []byte(`{"schemaVersion":2,"manifests":[]}`)
	hash := sha256.Sum256(index)

	var buffer bytes.Buffer

	writer := tar.NewWriter(&buffer)

	if large {
		layer := make([]byte, 9<<20)
		require.NoError(t, writer.WriteHeader(&tar.Header{Name: "blobs/sha256/test-public-layer", Typeflag: tar.TypeReg, Size: int64(len(layer))}))
		_, err := writer.Write(layer)
		require.NoError(t, err)
	}

	require.NoError(t, writer.WriteHeader(&tar.Header{Name: "blobs/sha256/" + hex.EncodeToString(hash[:]), Typeflag: tar.TypeReg, Size: int64(len(index))}))
	_, err := writer.Write(index)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return buffer.Bytes(), "test/owned@sha256:" + hex.EncodeToString(hash[:])
}

func TestDevelopmentImageStreamsBeyondJSONBudgetWithoutRetainingLayers(t *testing.T) {
	t.Parallel()

	for mode := range 6 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()
			archive, reference := developmentStreamArchive(t, mode == 0)

			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			factory := func(_ context.Context, _ string, args []string, input io.Reader, output io.Writer) developmentImageProcess {
				save := args[0] == developmentImage
				if !save {
					require.Contains(t, args, "--platform")
					require.Contains(t, args, "linux/amd64")
				}

				return &developmentStreamFixtureProcess{
					input: input, output: output, archive: archive,
					startErr: (mode == 1 && save) || (mode == 2 && !save),
					waitErr:  (mode == testDevThirdCase && save) || (mode == 4 && !save),
				}
			}

			ids := []string{strings.Repeat("a", sha256.Size*2), strings.Repeat("b", sha256.Size*2)}
			if mode == 5 {
				ids = nil
			}

			err := streamDevelopmentImage(ctx, "explicit-test-docker", reference, ids, factory)
			if mode == 0 || mode == 5 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
			}
		})
	}
}

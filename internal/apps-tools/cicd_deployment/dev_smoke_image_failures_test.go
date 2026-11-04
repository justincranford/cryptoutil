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
	"os/exec"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"

	"github.com/stretchr/testify/require"
)

type developmentImageFailWriter struct{}

const (
	testImageDuplicate      = "image-duplicate"
	testImageOtherBinding   = "image-other-binding"
	testImageCancelled      = "image-cancelled"
	testImageAbsent         = "image-absent"
	testImageUID            = "image-uid"
	testImageTransport      = "image-transport"
	testImageForged         = "image-forged"
	testImageEmpty          = "image-empty"
	testImageDuplicateUID   = "image-duplicate-uid"
	testImageCount          = "image-count"
	testImageValid          = "valid"
	testImageOversize       = "oversize"
	testImageLink           = "link"
	testImageMissing        = "missing"
	testImageMalformed      = "malformed"
	testImageWriter         = "writer"
	testImageInvalidHex     = "invalid-hex"
	testImageUppercase      = "uppercase"
	testImageOption         = "option"
	testImageNilContext     = "nil-context"
	testImageFactory        = "factory"
	testImageReference      = "reference"
	testImageMissingProcess = "missing-process"
	testImageNativeStart    = "native-start"
	testImageSecondStart    = "second-import-start"
	testImageQualified      = "qualified"
	testImageShort          = "short"
	testImageNamespace      = "namespace"
	testImageIncomplete     = "incomplete"
	testImageStream         = "stream"
)

func (developmentImageFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("PRIVATE")
}

func TestDevelopmentImageStreamRejectsForgedAndDuplicateBindings(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{testImageValid, testImageDuplicate, testImageOversize, testImageLink, testImageForged, testImageMissing, testImageMalformed, testImageWriter, testImageInvalidHex, testImageUppercase, testImageOption, testImageEmpty} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			index := []byte("public test index")
			hash := sha256.Sum256(index)
			digest := hex.EncodeToString(hash[:])
			reference := "test/owned@sha256:" + digest

			var buffer bytes.Buffer

			writer := tar.NewWriter(&buffer)

			count := 1
			if mode == testImageDuplicate {
				count++
			}

			for range count {
				header := &tar.Header{Name: "blobs/sha256/" + digest, Typeflag: tar.TypeReg, Size: int64(len(index))}
				if mode == testImageMissing {
					header.Name = testImageOtherBinding
				}

				if mode == testImageOversize {
					header.Size = developmentImageIndexLimit + 1
				}

				if mode == testImageLink {
					header.Typeflag, header.Linkname, header.Size = tar.TypeSymlink, testImageOtherBinding, 0
				}

				require.NoError(t, writer.WriteHeader(header))

				if mode != testImageLink {
					data := index
					if mode == testImageForged {
						data = bytes.Repeat([]byte("x"), len(index))
					}

					if mode == testImageOversize {
						data = make([]byte, developmentImageIndexLimit+1)
					}

					_, err := writer.Write(data)
					require.NoError(t, err)
				}
			}

			require.NoError(t, writer.Close())

			raw := buffer.Bytes()
			destination := io.Writer(io.Discard)

			switch mode {
			case testImageMalformed:
				raw = []byte("malformed")
			case testImageWriter:
				destination = developmentImageFailWriter{}
			case testImageInvalidHex:
				reference = "test/owned@sha256:" + strings.Repeat("z", sha256.Size*2)
			case testImageUppercase:
				reference = "test/owned@sha256:" + strings.ToUpper(digest)
			case testImageOption:
				reference = "--output=unauthorized@sha256:" + digest
			case testImageEmpty:
				reference = "test/owned:latest"
			}

			err := verifyDevelopmentImageStream(bytes.NewReader(raw), reference, destination)
			if mode == testImageValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
			}
		})
	}
}

func TestDevelopmentImageStreamDependencyAndProcessGuards(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{testImageNilContext, testImageCancelled, testImageFactory, testImageReference, testImageUID, testImageDuplicateUID, testImageMissingProcess, testImageNativeStart, testImageSecondStart} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			raw, reference := developmentStreamArchive(t, false)

			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			ids := []string{strings.Repeat("a", sha256.Size*2), strings.Repeat("b", sha256.Size*2)}
			calls := 0
			factory := developmentImageFactory(func(_ context.Context, _ string, _ []string, input io.Reader, output io.Writer) developmentImageProcess {
				calls++

				if mode == testImageMissingProcess {
					return nil
				}

				return &developmentStreamFixtureProcess{input: input, output: output, archive: raw, startErr: mode == testImageSecondStart && calls == 2}
			})

			switch mode {
			case testImageNilContext:
				ctx = nil
			case testImageCancelled:
				cancel()
			case testImageFactory:
				factory = nil
			case testImageReference:
				reference = "missing"
			case testImageUID:
				ids[0] = "mutable-container-name"
			case testImageDuplicateUID:
				ids[1] = ids[0]
			case testImageNativeStart:
				factory, ids = nativeDevelopmentImageProcess, nil
			}

			err := streamDevelopmentImage(ctx, "absent-development-image-executable", reference, ids, factory)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "PRIVATE")

			if mode == testImageCancelled {
				require.ErrorIs(t, err, context.Canceled)
			}
		})
	}
}

type developmentImageAvailabilityRunner struct {
	t        *testing.T
	expected string
	ids      []string
	fail     bool
}

func (runner developmentImageAvailabilityRunner) Run(_ context.Context, _ string, args []string, _ []byte) ([]byte, error) {
	require.Contains(runner.t, runner.ids, args[1])
	require.Equal(runner.t, developmentExec, args[0])
	require.Equal(runner.t, "--quiet", args[len(args)-1])

	if runner.fail {
		return nil, errors.New("PRIVATE")
	}

	return []byte(runner.expected), nil
}

func TestDevelopmentImageImportRequiresCapturedUIDsAndActualCRIName(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{testImageQualified, testImageShort, testImageNamespace, testImageIncomplete, testImageCount, testImageMissing, testImageTransport, testImageAbsent, testImageStream} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			raw, reference := developmentStreamArchive(t, false)
			cluster := gen.LocalK3sPlan{Server: gen.LocalK3sNode{Container: "owned-server"}}
			id := strings.Repeat("a", sha256.Size*2)
			input := developmentSmokeInput{docker: "explicit-test-docker", compute: developmentComputeSnapshot{
				Complete: true, Nodes: map[string]string{cluster.Server.Container: id},
			}}

			if mode == testImageShort {
				reference = strings.Replace(reference, "test/owned", "owned", 1)
			}

			if mode == testImageQualified {
				reference = "registry.test/" + reference
			}

			expected := reference
			if mode == testImageShort {
				expected = "docker.io/library/" + reference
			} else if mode != testImageQualified {
				expected = "docker.io/" + reference
			}

			switch mode {
			case testImageIncomplete:
				input.compute.Complete = false
			case testImageCount:
				input.compute.Nodes = nil
			case testImageMissing:
				input.compute.Nodes = map[string]string{"unrelated": id}
			case testImageAbsent:
				expected = "unrelated-immutable-name"
			}

			runner := developmentImageAvailabilityRunner{t: t, ids: []string{id}, expected: expected, fail: mode == testImageTransport}
			factory := func(_ context.Context, _ string, _ []string, reader io.Reader, writer io.Writer) developmentImageProcess {
				return &developmentStreamFixtureProcess{input: reader, output: writer, archive: raw, startErr: mode == testImageStream}
			}

			err := importDevelopmentImageWithFactory(t.Context(), input, cluster, reference, runner, factory)
			if mode == testImageQualified || mode == testImageShort || mode == testImageNamespace {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
			}
		})
	}
}

func TestDevelopmentImageNativeFactoryDoesNotOverrideEnvironment(t *testing.T) {
	t.Parallel()
	command := nativeDevelopmentImageProcess(t.Context(), "explicit-docker", []string{developmentImage, developmentImageSave, "immutable-reference"}, nil, io.Discard)
	native, ok := command.(*exec.Cmd)
	require.True(t, ok)
	require.Nil(t, native.Env)
	require.Equal(t, io.Discard, native.Stderr)
	require.Positive(t, native.WaitDelay)
}

var _ execution.Runner = developmentImageAvailabilityRunner{}

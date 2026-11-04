// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"archive/tar"
	"context"
	sha256 "crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const (
	developmentImageStreamLimit    = int64(8 << 30)
	developmentImageIndexLimit     = int64(1 << 20)
	developmentImageWaitDelay      = 5 * time.Second
	developmentImageSave           = "save"
	developmentContainerd          = "ctr"
	developmentContainerdImages    = "images"
	developmentContainerdNamespace = "k8s.io"
)

type developmentImageProcess interface {
	Start() error
	Wait() error
}

type developmentImageFactory func(context.Context, string, []string, io.Reader, io.Writer) developmentImageProcess

type developmentImageWait struct {
	done chan struct{}
	err  error
}

func nativeDevelopmentImageProcess(ctx context.Context, binary string, args []string, input io.Reader, output io.Writer) developmentImageProcess {
	command := exec.CommandContext(ctx, binary, args...)
	command.Stdin, command.Stdout, command.Stderr = input, output, io.Discard
	command.WaitDelay = developmentImageWaitDelay

	return command
}

func developmentImageDigest(reference string) (string, error) {
	index := strings.LastIndex(reference, "@sha256:")
	if index <= 0 || strings.HasPrefix(reference, "-") || strings.ContainsAny(reference, " \t\r\n\x00\\") {
		return "", errors.New("owned image immutable reference required")
	}

	digest := reference[index+len("@sha256:"):]

	raw, err := hex.DecodeString(digest)
	if err != nil || len(raw) != sha256.Size || strings.ToLower(digest) != digest {
		return "", errors.New("owned image digest encoding rejected")
	}

	return digest, nil
}

// verifyDevelopmentImageStream forwards the original archive without retaining
// layers. Only its inspected immutable index blob is buffered (at most 1 MiB).
// This does not relax the lifecycle ExecRunner's separate 8 MiB JSON limit.
func verifyDevelopmentImageStream(input io.Reader, reference string, output io.Writer) error {
	digest, err := developmentImageDigest(reference)
	if err != nil {
		return err
	}

	limited := &io.LimitedReader{R: input, N: developmentImageStreamLimit + 1}
	stream := io.TeeReader(limited, output)
	reader := tar.NewReader(stream)
	found := false

	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return errors.New("owned image archive stream rejected")
		}

		if header.Name != "blobs/sha256/"+digest {
			continue
		}

		if found || header.Typeflag != tar.TypeReg || header.Size <= 0 || header.Size > developmentImageIndexLimit {
			return errors.New("owned image index stream binding rejected")
		}

		raw, err := io.ReadAll(io.LimitReader(reader, developmentImageIndexLimit+1))

		actual := sha256.Sum256(raw)
		if err != nil || hex.EncodeToString(actual[:]) != digest {
			return errors.New("owned image index stream digest rejected")
		}

		found = true
	}

	if _, err := io.Copy(io.Discard, stream); err != nil || limited.N <= 0 {
		return errors.New("owned image archive stream bound rejected")
	}

	if !found {
		return errors.New("owned image inspected index absent")
	}

	return nil
}

// A zero-node call verifies an archive read-only. With nodes, each target must
// be a previously captured exact owned Docker UID, never a mutable name.
func streamDevelopmentImage(ctx context.Context, docker, reference string, nodeIDs []string, factory developmentImageFactory) (result error) {
	if ctx == nil || factory == nil {
		return errors.New("owned image streaming dependencies required")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("owned image streaming cancelled: %w", err)
	}

	if _, err := developmentImageDigest(reference); err != nil {
		return err
	}

	seen := map[string]bool{}

	for _, id := range nodeIDs {
		raw, err := hex.DecodeString(id)
		if err != nil || len(raw) != sha256.Size || strings.ToLower(id) != id || seen[id] {
			return errors.New("owned image target UID rejected")
		}

		seen[id] = true
	}

	processCtx, cancel := context.WithCancel(ctx)
	readers := []*io.PipeReader{}
	writers := []*io.PipeWriter{}
	waiters := []*developmentImageWait{}

	defer func() {
		cancel()

		for _, reader := range readers {
			_ = reader.Close()
		}

		for _, writer := range writers {
			_ = writer.Close()
		}

		for _, waiter := range waiters {
			<-waiter.done
		}

		if err := ctx.Err(); err != nil {
			result = errors.Join(result, err)
		}
	}()

	start := func(command developmentImageProcess, reader *io.PipeReader, writer *io.PipeWriter) error {
		if command == nil || command.Start() != nil {
			return errors.New("owned image streaming process unavailable")
		}

		waiter := &developmentImageWait{done: make(chan struct{})}
		waiters = append(waiters, waiter)

		go func() {
			err := command.Wait()
			if reader != nil {
				_ = reader.CloseWithError(err)
			}

			if writer != nil {
				_ = writer.CloseWithError(err)
			}

			waiter.err = err
			close(waiter.done)
		}()

		return nil
	}
	destinations := []io.Writer{}

	for _, id := range nodeIDs {
		reader, writer := io.Pipe()
		readers, writers = append(readers, reader), append(writers, writer)

		command := factory(processCtx, docker, []string{
			developmentExec, "-i", id, developmentK3s, developmentContainerd, developmentNamespaceFlag,
			developmentContainerdNamespace, developmentContainerdImages, "import", "--platform", "linux/amd64", "--digests", developmentStdin,
		}, reader, io.Discard)
		if err := start(command, reader, nil); err != nil {
			return err
		}

		destinations = append(destinations, writer)
	}

	reader, writer := io.Pipe()
	readers, writers = append(readers, reader), append(writers, writer)

	save := factory(processCtx, docker, []string{developmentImage, developmentImageSave, reference}, nil, writer)
	if err := start(save, nil, writer); err != nil {
		return err
	}

	output := io.Writer(io.Discard)
	if len(destinations) != 0 {
		output = io.MultiWriter(destinations...)
	}

	if err := verifyDevelopmentImageStream(reader, reference, output); err != nil {
		return err
	}

	for _, destination := range destinations {
		if closer, ok := destination.(io.Closer); ok {
			_ = closer.Close()
		}
	}
	// Collect successful exit results before canceling attached commands.
	for _, waiter := range waiters {
		<-waiter.done

		if waiter.err != nil {
			return errors.New("owned image streaming process rejected")
		}
	}

	return nil
}

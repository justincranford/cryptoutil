// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

func Main(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 0 && args[0] == ProvisionRoles {
		return roleProvisionMain(args[1:], stdout, stderr, readProtected, ProvisionOrdinaryRoles)
	}

	return MainWith(args, stdout, stderr, readProtected,
		func(ctx context.Context, operation string, input Input) error {
			return Execute(ctx, operation, input, OpenProtected(readProtected))
		})
}

func MainWith(args []string, stdout, stderr io.Writer, read func(string) ([]byte, error), execute func(context.Context, string, Input) error) int {
	flags := flag.NewFlagSet(Command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	operation := flags.String("operation", "", "explicit lifecycle operation")

	file := flags.String("input", "", "protected lifecycle manifest")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(stdout, "cryptoutil db-lifecycle --operation=init|migrate|replicate --input=/run/cryptoutil/secrets/OWNED-SECRET/lifecycle.json")
			if err == nil {
				return 0
			}
		}

		return failure(stderr)
	}

	if flags.NArg() != 0 || !protectedPath(*file) || path.Base(*file) != InputName {
		return failure(stderr)
	}

	raw, err := read(*file)
	if err != nil {
		return failure(stderr)
	}

	input, err := Decode(raw)
	if err != nil {
		return failure(stderr)
	}

	if err := input.Validate(*operation); err != nil {
		return failure(stderr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	if err := execute(ctx, *operation, input); err != nil {
		return failure(stderr)
	}

	if _, err := fmt.Fprintln(stdout, "PASS: database lifecycle operation verified"); err != nil {
		return 1
	}

	return 0
}

func readProtected(file string) ([]byte, error) {
	return readProtectedWith(file, func(name string) (io.ReadCloser, error) { return os.Open(name) })
}

func readProtectedWith(file string, open func(string) (io.ReadCloser, error)) ([]byte, error) {
	if !protectedPath(file) {
		return nil, errors.New("protected database source rejected")
	}

	handle, err := open(file)
	if err != nil {
		return nil, errors.New("protected database source unavailable")
	}

	raw, err := io.ReadAll(io.LimitReader(handle, maximumInput+1))

	closeErr := handle.Close()
	if err != nil || closeErr != nil || len(raw) == 0 || len(raw) > maximumInput {
		return nil, errors.New("protected database source read rejected")
	}

	return raw, nil
}

func failure(stderr io.Writer) int {
	_, _ = fmt.Fprintln(stderr, "database lifecycle input or operation rejected")

	return 1
}

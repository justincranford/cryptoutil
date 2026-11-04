// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path"

	"gopkg.in/yaml.v3"
)

const RoleProvisionInputName = "roles.json"

func DecodeRoleProvision(raw []byte) (RoleProvisionInput, error) {
	var input RoleProvisionInput
	if len(raw) == 0 || len(raw) > maximumInput {
		return input, errors.New("protected role manifest size rejected")
	}

	var unique map[string]any
	if err := yaml.Unmarshal(raw, &unique); err != nil {
		return input, errors.New("protected role manifest properties rejected")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return RoleProvisionInput{}, errors.New("protected role manifest encoding rejected")
	}

	var tail any
	if err := decoder.Decode(&tail); !errors.Is(err, io.EOF) {
		return RoleProvisionInput{}, errors.New("protected role manifest trailing content")
	}

	if err := input.Validate(); err != nil {
		return RoleProvisionInput{}, err
	}

	return input, nil
}

func roleProvisionMain(args []string, stdout, stderr io.Writer, read func(string) ([]byte, error),
	provision func(context.Context, RoleProvisionInput, Open, func(string) ([]byte, error)) error,
) int {
	flags := flag.NewFlagSet(ProvisionRoles, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	file := flags.String("input", "", "protected run-bound ordinary PostgreSQL role manifest")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(stdout, "cryptoutil db-lifecycle provision --input=/run/cryptoutil/secrets/OWNED-SECRET/roles.json; ordinary roles only")
			if err == nil {
				return 0
			}
		}

		return failure(stderr)
	}

	if flags.NArg() != 0 || !protectedPath(*file) || path.Base(*file) != RoleProvisionInputName {
		return failure(stderr)
	}

	raw, err := read(*file)
	if err != nil {
		return failure(stderr)
	}

	input, err := DecodeRoleProvision(raw)
	if err != nil {
		return failure(stderr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	if err := provision(ctx, input, OpenProtected(read), read); err != nil {
		return failure(stderr)
	}

	if _, err := fmt.Fprintln(stdout, "ordinary role ownership and protected authentication verified; replication/publication privileges not granted"); err != nil {
		return 1
	}

	return 0
}

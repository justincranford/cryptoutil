# identity-authz

## Overview

This directory contains the identity-authz application entrypoint package.

## Canonical Root Files

- authz.go
- authz_usage.go
- authz_cli_test.go
- authz_port_conflict_test.go
- testmain_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

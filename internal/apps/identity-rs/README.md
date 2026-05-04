# identity-rs

## Overview

This directory contains the identity-rs application entrypoint package.

## Canonical Root Files

- rs.go
- rs_usage.go
- rs_cli_test.go
- rs_lifecycle_test.go
- rs_port_conflict_test.go
- testmain_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

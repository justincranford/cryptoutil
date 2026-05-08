# identity-rp

## Overview

This directory contains the identity-rp application entrypoint package.

## Canonical Root Files

- rp.go
- rp_cli_test.go
- rp_port_conflict_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

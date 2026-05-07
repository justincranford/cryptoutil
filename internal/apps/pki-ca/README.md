# pki-ca

## Overview

This directory contains the pki-ca application entrypoint package.

## Canonical Root Files

- ca.go
- ca_cli_test.go
- ca_port_conflict_test.go
- testmain_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

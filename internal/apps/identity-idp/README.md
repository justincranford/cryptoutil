# identity-idp

## Overview

This directory contains the identity-idp application entrypoint package.

## Canonical Root Files

- idp.go
- idp_cli_test.go
- idp_port_conflict_test.go
- testmain_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

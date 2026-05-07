# sm-kms

## Overview

This directory contains the sm-kms application entrypoint package.

## Canonical Root Files

- kms.go
- kms_cli_test.go
- kms_port_conflict_test.go
- testmain_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

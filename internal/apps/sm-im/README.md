# sm-im

## Overview

This directory contains the sm-im application entrypoint package.

## Canonical Root Files

- im.go
- im_cli_test.go
- im_port_conflict_test.go
- testmain_test.go

## Canonical Subdirectories

- client/
- server/
- e2e/

## Notes

- This file is enforced by the `apps-ps-id-template` fitness check.
- Service-specific implementation details belong under `server/` and may differ by service.

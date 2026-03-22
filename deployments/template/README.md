# Template Directory

Scaffolding templates for creating **new** product-service deployments. NOT a deployable service.

## Purpose

Contains boilerplate compose files, configs, and secrets with placeholder names (underscores, `PRODUCT-SERVICE` tokens) that are copied and customized when adding a new service to the project.

## Files

- `compose.yml` — Base service deployment template
- `compose-cryptoutil-PRODUCT-SERVICE.yml` — Service-tier template
- `compose-cryptoutil-PRODUCT.yml` — Product-tier template
- `compose-cryptoutil.yml` — Suite-tier template
- `config/` — Template config files with placeholder settings
- `secrets/` — Template secret files with placeholder values (underscores)

## Usage

See the `/new-service` skill for step-by-step instructions on creating a new service from these templates.

## Not to Be Confused With

**`deployments/skeleton-template/`** is the actual deployed skeleton-template service — a running reference implementation. This `template/` directory is the inert scaffolding source.

# CLAUDE.md

Guidance for Claude Code (and other agents) working in this repository.

## Project Purpose

Starlight is a self-service golden-path platform that lets developers deploy a new
microservice to Kubernetes — with security, observability, and policy enforcement
applied automatically — in minutes instead of days. It is built entirely with free
and open-source tooling; no paid licenses or third-party SaaS are used.

## Conventions

- **Kubernetes resource names**: kebab-case (e.g. `payments-api`, `order-events-consumer`).
- **Git commit messages**: [Conventional Commits](https://www.conventionalcommits.org/)
  (e.g. `feat:`, `fix:`, `chore:`, `docs:`).
- **Secrets**: never commit secrets or `.env` files. Use local-only config and
  `.gitignore` to keep them out of version control.

## Current Phase

Phase 1: building the golden-path CLI and local k3d environment.

## Tooling

All tooling must be free and open-source. Current stack includes k3d, kubectl, Helm,
Argo CD, Prometheus, and Grafana — no paid SaaS.

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

Phase 1: building the golden-path CLI and GitOps core (Argo CD) on a local cluster.

## Tooling

All tooling must be free and open-source. Current stack includes kubectl, Helm,
Argo CD, Prometheus, and Grafana — no paid SaaS.

## Environment Assumption

Local development targets Docker Desktop Kubernetes (single-node), currently
running as context "docker-desktop". k3d may be introduced later specifically for
multi-node failure-scenario testing (Phase 4/5 reliability work), documented
separately when that increment happens. Manifests and charts should stay
Kubernetes-distro-agnostic where possible; AWS-specific concerns (IRSA, EKS Pod
Identity, VPC networking) are deferred to a later cloud-parity phase and should not
be assumed in local-only work.

## Safety

Never run destructive commands (terraform destroy, kubectl delete namespace,
git push --force) without explicit confirmation from the user in the current session.
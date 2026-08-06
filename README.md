# Starlight

## 1. Project Summary

Starlight is a self-service golden-path platform that lets developers deploy a new microservice — with security, observability, and policy enforcement applied automatically — in minutes instead of days.

## 2. Business Problem

## 3. Target Users

## 4. Key Capabilities

- **Golden-path CLI**: `starlight create-service` ([cli/](cli/)) scaffolds a new microservice — a Helm chart under `platform/charts/<name>/` and an Argo CD `Application` manifest under `platform/gitops/` — from the sample-service pattern. It only generates files; committing, pushing, and `kubectl apply` to register with Argo CD stay explicit, human-run steps. See [cli/README.md](cli/README.md).

## 5. Architecture Diagram

## 6. Deployment Flow

## 7. Technology Choices

- **GitOps engine**: [Argo CD](https://argo-cd.readthedocs.io/) v3.5.0 (Helm chart `argo/argo-cd` v10.2.3) was installed via Helm into the `argocd` namespace on the local `docker-desktop` Kubernetes cluster. See [platform/gitops/install-argocd.md](platform/gitops/install-argocd.md) for the exact install commands.

## 8. Repository Structure

## 9. Prerequisites

## 10. Local Quick Start

## 11. Cloud Deployment

## 12. Security Model

## 13. Observability

## 14. Testing Strategy

## 15. Reliability Model

## 16. Cost Estimate

## 17. Known Limitations

## 18. Architecture Decisions

## 19. Roadmap

## 20. Demo

## 21. Lessons Learned

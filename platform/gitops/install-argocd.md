# Installing Argo CD via Helm (local: docker-desktop)

This documents the exact commands used to install Argo CD into the local
`docker-desktop` Kubernetes cluster using the official Helm chart, pinned to
a specific version (Phase 1: golden-path CLI and GitOps core).

## Prerequisites

- `kubectl` context set to `docker-desktop`
- `helm` v3+ available on PATH

## Commands

```sh
# 1. Add the official Argo Helm chart repository and refresh the index
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

# 2. (Optional) confirm which chart versions are available before pinning
helm search repo argo/argo-cd --versions

# 3. Install Argo CD into a new "argocd" namespace, pinned to chart 10.2.3
#    (app version v3.5.0) — not "latest" — and wait for all resources to
#    become ready before returning.
helm install argocd argo/argo-cd \
  --version 10.2.3 \
  --namespace argocd \
  --create-namespace \
  --timeout 15m0s \
  --wait
```

## Chart / app version installed

- Helm chart: `argo/argo-cd` @ **10.2.3**
- Argo CD app version: **v3.5.0**

## Verifying the install

```sh
kubectl get pods -n argocd
```

All pods (application-controller, applicationset-controller, dex-server,
notifications-controller, redis, repo-server, server) should show
`STATUS: Running`.

## Retrieving the initial admin password

The initial admin password is stored in a Kubernetes Secret. Retrieve it
on demand — do **not** copy it into any file in this repo:

```sh
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d
```

Per the Argo CD getting-started guide, delete this secret after the first
login and switch to a normal admin password / SSO:

```sh
kubectl -n argocd delete secret argocd-initial-admin-secret
```

## Accessing the UI locally

```sh
kubectl port-forward service/argocd-server -n argocd 8080:443
```

Then open https://localhost:8080 (accept the self-signed certificate) and
log in as `admin` with the password retrieved above.

## Registering the sample-service Application

`platform/gitops/sample-service-app.yaml` defines an Argo CD `Application`
that points back at this repo (path `platform/charts/sample-service`) with
automated sync and self-heal enabled, to prove the GitOps loop end-to-end.
It is **not** applied automatically — register it explicitly when you're
ready to test the loop:

```sh
kubectl apply -f platform/gitops/sample-service-app.yaml
```

Argo CD will then create the `sample-service` namespace and deploy the
chart from Git. Watch progress with:

```sh
kubectl get application sample-service -n argocd
kubectl get pods -n sample-service
```

## Notes

- Installed via the official `argo-helm` chart repository
  (https://argoproj.github.io/argo-helm) — free and open-source, no paid
  SaaS involved, consistent with this project's tooling constraints.
- First install on a fresh Docker Desktop cluster can be slow because the
  `argocd-redis-secret-init` pre-install hook Job has to pull the
  `quay.io/argoproj/argocd` image (~210MB); the default 5-minute Helm wait
  timed out once here purely on image-pull time, so an explicit
  `--timeout 15m0s` was used on the successful run.
- The `argocd` Custom Resource Definitions (`applications.argoproj.io`,
  `applicationsets.argoproj.io`, `appprojects.argoproj.io`) are retained by
  Helm's resource policy even on `helm uninstall`, by design, to avoid
  data loss.

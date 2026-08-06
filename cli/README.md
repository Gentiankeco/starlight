# starlight CLI

The golden-path CLI for scaffolding a new microservice into this repo's
GitOps layout. It generates a Helm chart under `platform/charts/<name>/`
and an Argo CD `Application` manifest under `platform/gitops/`, both
modeled on `platform/charts/sample-service` and
`platform/gitops/sample-service-app.yaml`.

The CLI **only generates files**. It never runs `git` or `kubectl` for
you — per the Safety section in the root `CLAUDE.md`, registering the
service with Argo CD and pushing it to Git stays a deliberate, human step.

## Build

Run from the `cli/` directory (module `starlight-cli`, no external
dependencies):

```sh
go build ./...
```

This produces a `starlight` (or `starlight.exe` on Windows) binary in the
current directory. You can also build to a specific path:

```sh
go build -o starlight .
```

## Run

From the **repository root** (generated paths are relative to the current
working directory):

```sh
./cli/starlight create-service
```

You'll be prompted for a service name:

```
Service name: Payments Api
Invalid service name "Payments Api": use lowercase letters, numbers, and hyphens only (kebab-case), e.g. "order-events-consumer".
Service name: payments-api
```

Names must be kebab-case: lowercase letters, numbers, and hyphens only
(matching the Kubernetes resource naming convention in `CLAUDE.md`).
Invalid input is rejected and you're re-prompted.

You can also skip the interactive prompt with `-name`:

```sh
./cli/starlight create-service -name payments-api
```

## Output

For a service named `payments-api`, this generates:

- `platform/charts/payments-api/` — a copy of the sample-service Helm
  chart (Chart.yaml, values.yaml, deployment + service templates) with
  names substituted
- `platform/gitops/payments-api-app.yaml` — an Argo CD `Application`
  targeting namespace `payments-api`, with automated sync and self-heal
  enabled

The CLI prints the next steps after generating the files:

```sh
git add platform/charts/payments-api platform/gitops/payments-api-app.yaml
git commit -m "feat: add payments-api service"
git push
kubectl apply -f platform/gitops/payments-api-app.yaml   # registers with Argo CD
```

## Test

```sh
go test ./...
```

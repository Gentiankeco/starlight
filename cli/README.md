# starlight CLI

The golden-path CLI for scaffolding a new microservice into this repo's
GitOps layout. It generates a Helm chart under `platform/charts/<name>/`,
an Argo CD `Application` manifest under `platform/gitops/`, and a GitHub
Actions CI workflow template under `platform/ci-templates/`. The chart
and Application manifest are modeled on `platform/charts/sample-service`
and `platform/gitops/sample-service-app.yaml`.

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

You'll be prompted for a service name, then a container image, then the
port your container listens on, then the GitHub username or org the
service's image is published under:

```
Service name: Payments Api
Invalid service name "Payments Api": use lowercase letters, numbers, and hyphens only (kebab-case), e.g. "order-events-consumer".
Service name: payments-api
Container image (e.g. nginx:latest or ghcr.io/you/your-app:v1): ghcr.io/gentiankeco/payments-api:v1
Container port (default 8080, press Enter to accept): 8080
GitHub username or org (for ghcr.io image path, e.g. Gentiankeco): Gentiankeco
```

Names must be kebab-case: lowercase letters, numbers, and hyphens only
(matching the Kubernetes resource naming convention in `CLAUDE.md`).
Invalid input is rejected and you're re-prompted.

The container image can be any valid image reference — a bare name like
`nginx:latest` or a full registry path like `ghcr.io/you/your-app:v1`. Blank
input is rejected and you're re-prompted. The reference is split into
repository and tag (on the last colon after the last slash, so a registry
host with a port such as `localhost:5000/app` isn't mistaken for a tag) and
used to populate `image.repository` and `image.tag` in the generated
chart's `values.yaml`. If no tag is given, `latest` is used.

The container port is the port your application listens on inside the
container. Press Enter to accept the default of `8080`, or enter any
port number from 1-65535. Invalid input (non-numeric, `0`, or above
`65535`) is rejected and you're re-prompted. The value populates
`containerPort` in the generated chart's `values.yaml`, which drives the
container's `containerPort`, the liveness/readiness probe ports, and the
Service's `targetPort` — keeping them in sync so the pod actually
becomes ready and reachable.

The GitHub username or org must be non-empty and contain no spaces.
Invalid input is rejected and you're re-prompted. It's used as the owner
segment of the `ghcr.io/<username>/<service-name>` image path in the
generated CI workflow template.

You can also skip the interactive service name prompt with `-name` (the
image, container port, and GitHub username prompts still run):

```sh
./cli/starlight create-service -name payments-api
```

## Output

For a service named `payments-api` with image `ghcr.io/you/payments-api:v1`
and GitHub owner `Gentiankeco`, this generates:

- `platform/charts/payments-api/` — a copy of the sample-service Helm
  chart (Chart.yaml, values.yaml, deployment + service templates) with
  names substituted and `values.yaml`'s `image.repository` /
  `image.tag` / `containerPort` set from what you entered
- `platform/gitops/payments-api-app.yaml` — an Argo CD `Application`
  targeting namespace `payments-api`, with automated sync and self-heal
  enabled
- `platform/ci-templates/payments-api-ci.yml` — a ready-to-use GitHub
  Actions workflow that builds and pushes the service's image to
  `ghcr.io/Gentiankeco/payments-api` (tagged with both `github.sha` and
  `latest`) on every push to `main`, using the target repo's own
  `secrets.GITHUB_TOKEN` — no PAT needed there. This file lives in the
  `starlight` repo; it isn't automatically copied into the service's own
  repo (see the next-steps output below, or
  [docs/bootstrap-workflow-guide.md](../docs/bootstrap-workflow-guide.md)
  for the automated path).

The CLI prints the next steps after generating the files:

```sh
git add platform/charts/payments-api platform/gitops/payments-api-app.yaml platform/ci-templates/payments-api-ci.yml
git commit -m "feat: add payments-api service"
git push
kubectl apply -f platform/gitops/payments-api-app.yaml   # registers with Argo CD
# Copy platform/ci-templates/payments-api-ci.yml into your app repo at
# .github/workflows/build-and-push.yml — or trigger the Starlight
# bootstrap workflow to deliver it automatically.
```

## Test

```sh
go test ./...
```

# Bootstrap workflow guide

## What this workflow does

`bootstrap-service-ci` ([.github/workflows/bootstrap-service-ci.yml](../.github/workflows/bootstrap-service-ci.yml))
delivers CI directly into a developer's service repository. Given a
service name and a target repo, it:

1. Reads the golden-path CI template for that service from
   `platform/ci-templates/<service_name>-ci.yml` in this repo.
2. Clones the target repo and creates (or resets) a `devops-ci` branch.
3. Copies the template into the target repo as
   `.github/workflows/build-and-push.yml`.
4. Commits the change as `Starlight Platform <platform@starlight>` and
   pushes the `devops-ci` branch.

Nothing is merged automatically — the developer reviews and merges the
branch themselves, same as the `create-service` CLI's own generated
files stay unapplied until a human commits and pushes them (see
[cli/README.md](../cli/README.md)).

## Prerequisites

- A `BOOTSTRAP_TOKEN` secret in the `starlight` repo (see below).
- A CI template already present at
  `platform/ci-templates/<service_name>-ci.yml`. If it's missing, the
  workflow fails with:
  > No CI template found for service '\<name\>'. Run starlight
  > create-service first.

## Generating the PAT

The workflow authenticates cross-repo writes with a classic GitHub
personal access token — see
[docs/adr/0002-ci-bootstrap-auth.md](adr/0002-ci-bootstrap-auth.md) for
why a PAT was chosen over the default `GITHUB_TOKEN` or a GitHub App.

1. GitHub → Settings → Developer settings → Personal access tokens →
   Tokens (classic) → Generate new token (classic).
2. Scope: `repo` (full control of private repositories — needed to
   clone and push to the target repo).
3. Set an expiration and generate. Copy the token now; GitHub won't
   show it again.

## Adding it as a repo secret

1. In the `starlight` repo: Settings → Secrets and variables → Actions
   → New repository secret.
2. Name: `BOOTSTRAP_TOKEN`.
3. Value: the PAT you just generated.
4. Save.

## Triggering the workflow

1. GitHub → Actions → `bootstrap-service-ci` → Run workflow.
2. Fill in:
   - **service_name**: the service's name, matching
     `platform/ci-templates/<service_name>-ci.yml`.
   - **target_repo**: `owner/repo`, e.g.
     `Gentiankeco/starlight-demo-app`.
3. Run workflow.

## What the developer sees

A `devops-ci` branch appears in their repo, containing
`.github/workflows/build-and-push.yml`, committed by `Starlight
Platform <platform@starlight>` with the message:

```
ci: add Starlight golden-path CI workflow [automated]
```

The workflow run's log prints the branch URL directly:
`https://github.com/<target_repo>/tree/devops-ci`.

## What the developer should do

Review the workflow file on the `devops-ci` branch, open a PR (or merge
it directly) once they're satisfied it's correct for their service, and
delete the branch afterward if their workflow doesn't do that
automatically. Re-running `bootstrap-service-ci` at any point resets
`devops-ci` back to `main` and force-pushes a fresh copy of the
template — it does not preserve manual edits made on that branch.

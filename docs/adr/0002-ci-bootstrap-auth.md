# 2. Authentication for the CI bootstrap workflow

## Status

Accepted

## Context

The `bootstrap-service-ci` workflow ([.github/workflows/bootstrap-service-ci.yml](../../.github/workflows/bootstrap-service-ci.yml))
runs in the `starlight` repo but needs write access to a *different*
repository: it clones a developer's service repo, pushes a `devops-ci`
branch containing a generated CI workflow, and leaves it for the
developer to review and merge.

GitHub Actions' automatically-injected `secrets.GITHUB_TOKEN` cannot do
this. It is scoped to the repository the workflow runs in and has no
permissions on any other repository — by design, to stop a workflow in
one repo from silently reaching into another. Some credential with
explicit access to the target repo is required.

## Decision

Use a classic GitHub personal access token (PAT) with `repo` scope,
stored as the `BOOTSTRAP_TOKEN` secret in the `starlight` repo. The
workflow reads it via `secrets.BOOTSTRAP_TOKEN` and uses it both to
authenticate the `git clone`/`git push` against the target repo and as
the `GH_TOKEN`/`GITHUB_TOKEN` environment for any `gh` CLI use.

## Alternatives Considered

**`secrets.GITHUB_TOKEN` (the default).** Rejected — it's repo-scoped by
GitHub and cannot be granted access to another repository. Not viable
for this workflow's cross-repo write, regardless of configuration.

**GitHub App.** The correct answer for production: a GitHub App
installed on the target repos gets narrowly-scoped, auditable,
per-installation permissions, isn't tied to any one person's account,
and can be revoked per repo without affecting anything else. Rejected
*for now* because it requires standing up and registering an app,
managing its private key, and handling installation-token exchange —
meaningfully more setup than this project's current phase (Phase 1,
per [CLAUDE.md](../../CLAUDE.md)) warrants for a single-developer
bootstrap flow.

## Advantages of a PAT

- Simple: one secret, no extra infrastructure to stand up or operate.
- Works today, with tooling (`git`, `gh`) that already expects this
  auth style.

## Disadvantages of a PAT

- Long-lived: unlike an installation token, it doesn't expire on its
  own and has to be manually rotated.
- Broadly scoped: `repo` scope grants access to every repository the
  token owner can reach, not just the ones this workflow targets.
- Tied to a person: the token is minted under one individual's GitHub
  account, not the organization or the workflow itself.
- Revocation is all-or-nothing: revoking it (e.g. because the owner
  left) breaks every automation using it, not just this one.

## Risks

- If the person who generated `BOOTSTRAP_TOKEN` leaves the org or their
  access is revoked, `bootstrap-service-ci` breaks until someone
  generates and re-adds a replacement token.
- Because the scope is broad, a leaked `BOOTSTRAP_TOKEN` exposes more
  than this workflow strictly needs.

## Conditions to Revisit

Move to a GitHub App when any of the following becomes true:

- More than one platform-team member relies on this workflow (a PAT
  tied to a single person's identity stops being an acceptable single
  point of failure).
- The bootstrap flow is used against production or otherwise
  business-critical target repos, not just local/demo services.
- Target repos span multiple owners/orgs, where a single PAT's blanket
  `repo` scope becomes both impractical to grant and a disproportionate
  security exposure.

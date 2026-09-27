# GoreeCloud Code

GoreeCloud Code is the first-party GoreeCloud developer and source-control platform for repositories, collaboration, CI/CD, packages, security, AI-assisted development, and GoreeCloud platform integrations.

Forgejo is the preferred initial, replaceable infrastructure foundation. It is not the permanent GoreeCloud product boundary. GoreeCloud Code owns the user experience, service contracts, governance model, integration architecture, and long-term developer platform.

## Status

**Milestone 1 — Forgejo Connectivity (in progress)**

Milestone 0 established the product boundary, provider abstraction, Forgejo adapter, web application shell, shared contracts, CI foundation, and deployment scaffolding. M1 now includes the runnable GoreeCloud-owned API service, live Glaze UI repository dashboard, provider-neutral repository activity reads, and an end-to-end Forgejo validation stack.

Current M1 capabilities include provider health/version reporting, repository discovery and detail retrieval, branches, commits, issues, and pull-request reads. A real-instance validation run is still required before M1 is complete.

## Architecture

```text
GoreeCloud Code
├── apps/web                 Glaze UI developer experience
├── packages/contracts       Provider-neutral domain contracts
├── integrations/forgejo     Replaceable Forgejo provider
├── services/api             GoreeCloud-owned product API
├── deploy/forgejo           M1 Forgejo validation deployment
├── scripts                  End-to-end validation tools
└── docs                     Architecture, ADRs, security, migration
```

The runtime boundary is intentionally replaceable:

```text
Glaze UI / GoreeCloud clients
            │
            ▼
     GoreeCloud Code API
            │
            ▼
       ForgeProvider
            │
            ├── ForgejoProvider      initial implementation
            ├── GitHubProvider       migration/interoperability
            └── NativeProvider       future GoreeCloud implementation
```

The browser must not receive Forgejo credentials or depend directly on Forgejo-specific APIs.


## Repository visibility model

GoreeCloud Code is intended to be a public-facing source-control service while supporting both public and access-controlled private repositories. Service visibility and repository visibility are separate boundaries: making GoreeCloud Code publicly reachable must not make every hosted repository public.

Repository visibility must follow the information and collaboration boundary of each project. Public repositories remain appropriate for intentionally published source, documentation, and ecosystem collaboration. Private repositories must remain access-controlled and must not expose source, metadata, artifacts, issues, pull requests, CI evidence, or repository inventory to unauthenticated users.

Wardveil Security and GoreeCloud Privacy Shield are planned to be **private repositories in GoreeCloud Code**. Their current GitHub visibility during the transition does not define their eventual GoreeCloud Code visibility.

The migration path must preserve Git history and exact revision identity while explicitly recreating repository visibility, access control, branch protection, review requirements, secrets separation, CI/CD boundaries, recovery, and any metadata outside Git before authority is transferred.

## Local connectivity

Configure the GoreeCloud Code API using `.env.example`. At minimum, provide `FORGEJO_BASE_URL`. Private repository discovery requires a narrowly scoped `FORGEJO_TOKEN`; anonymous public discovery can use `FORGEJO_USERNAME` when supported by the instance.

The provider-neutral API includes:

- `GET /health`
- `GET /api/v1/provider`
- `GET /api/v1/repositories`
- `GET /api/v1/repositories/:owner/:name`
- repository branches, commits, issues, and pull-request read endpoints

For an isolated M1 test environment, use `deploy/forgejo/compose.yml` with `deploy/forgejo/.env.example`. This brings up Forgejo and PostgreSQL for local validation; it is not the production deployment architecture.

Once Forgejo and the GoreeCloud Code API are running, validate the provider boundary with:

```sh
FORGEJO_BASE_URL=http://localhost:3000 \
GOREECLOUD_CODE_API_URL=http://localhost:8787 \
VALIDATE_REPOSITORY=owner/repository \
pnpm validate:forgejo
```

See `docs/architecture/m1-forgejo-connectivity.md` for the full validation gate. M1 must not be marked complete until a real test run is recorded.

## Platform integrations

GoreeCloud Code is designed to integrate with substantive GoreeCloud platform systems as implementations become available:

- **Glaze UI** — developer experience and interaction system.
- **GoreeCloud AI** — governed AI-assisted software development.
- **Wardveil Security** — evidence-backed repository, runner, dependency, artifact, and deployment security state.
- **Privacy Shield** — privacy-control contracts, data minimization, retention, and telemetry governance.
- **Everkeep** — backup, recovery, preservation, portability, and succession.
- **GoreeCloud Mesh** — coordination and governance between first-party applications and services.

Public claims about these systems must remain tied to implemented capabilities and available evidence.

## Development principles

1. Standard Git remains the repository interoperability foundation.
2. Forgejo is replaceable infrastructure, not the permanent product boundary.
3. GoreeCloud-owned APIs and contracts must not expose unnecessary Forgejo-specific assumptions.
4. CI logic should remain portable between GitHub Actions, Forgejo Actions, and future GoreeCloud Pipelines.
5. Security-sensitive capabilities require least privilege, explicit authorization, and evidence-backed controls.
6. Repository migration must preserve Git history and deliberately account for metadata outside Git itself.

## Initial roadmap

### M0 — Bootstrap

Complete foundation: monorepo, provider contracts, Forgejo adapter, web shell, API boundary, CI, and architecture decisions.

### M1 — Forgejo connectivity

In progress: real-instance authentication and validation, repository discovery/detail, branches, commits, issues, pull requests, provider health, live web dashboard, and end-to-end validation tooling.

### M2 — Governed write operations

Planned: branch creation, issue mutation, pull-request creation, review workflows, and bounded GoreeCloud AI operations.

### M3 — Pipelines, packages, and migration

Planned: portable workflow execution, GoreeCloud-controlled runners, package/OCI registry integration, GitHub import tooling, and external repository mirroring.

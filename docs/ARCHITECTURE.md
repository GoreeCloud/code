# Architecture

**Current status:** planned target with a small read-only source foundation; not production accepted.

```text
Glaze Code web/desktop/mobile clients
        |
GoreeCloud Code API and capability broker
        |-- GoreeCloud Identity + Code-owned authorization
        |-- Wardveil Security / Policy / Privacy Shield
        |-- audit, idempotency and reconciliation
        |-- provider-neutral ForgeProvider
               |-- Forgejo adapter (initial)
               |-- future native or external adapters
        |-- GoreeCloud Containers / pipelines
        |-- GoreeCloud AI (approved scoped operations)
        |-- Everkeep / Observability / Manager / Mesh
```

The current API only implements a fixed repository read route and a minimal health endpoint with server-held secret files. It performs no writes. It is not Identity authentication, a production capability broker or accepted Wardveil security integration.

## Required migration boundaries

GoreeCloud Code owns developer-facing contracts, policy, workflow UI and trusted execution boundaries. Forgejo's internal database, template UI and release schedule must not silently become GoreeCloud-owned interface contracts. Git repositories remain standard. Migrate Git and non-Git collaboration metadata separately, with snapshots and verified restore rehearsals.

## Recommended implementation order

1. Import pinned upstream source with exact SHA, license, notices and deterministic build verification.
2. Establish provider-neutral contracts, per-resource authorization and scoped server-side provider credentials.
3. Add durable pre-execution audit, idempotency, uncertainty reconciliation and policy enforcement before writes.
4. Adopt current Stable Glaze components in controlled UI surfaces, with accessibility and responsive acceptance.
5. Implement isolation for build runners, CI/CD, packages and AI tools; verify recovery and rollout/rollback.

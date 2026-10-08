# GoreeCloud Code

**Lifecycle: Development — not Stable or production deployed.**

GoreeCloud Code will be the first-party GoreeCloud developer and source-control platform. The designated implementation repository is **GoreeCloud/code**. The development model is **Fork-to-Native**, using Forgejo as a bounded initial Git-hosting foundation while GoreeCloud owns developer-facing contracts, workflows, identity and security boundaries, and product experience.

## Working foundation

- `services/api/` — loopback-only, authenticated and read-only Code API prototype with explicit repository allowlisting, four sanitized repository/collaboration read routes, and server-held Forgejo credentials.
- `apps/web/` — static accessible Glaze-semantic interface preview, **not** Glaze V1.7 consumer acceptance.
- `vendor/forgejo/` — verified imported Forgejo v15.0.9 LTS tracked-source snapshot and provenance manifest. The branch-bound importer is in `scripts/import-forgejo.sh`.
- `.github/workflows/` — Code CI and source-import workflow; the initial import completed at [run 37716860233](https://github.com/GoreeCloud/code/actions/runs/37716860233).
- `docs/` — requirements, architecture, security, privacy, validation, planned and implemented feature status.

## Local developer commands

Requires Node.js 22+; the initial service has no third-party runtime dependencies.

```sh
node --check services/api/server.mjs
node --test services/api/*.test.mjs
bash -n scripts/import-forgejo.sh
node services/api/server.mjs
```

The service binds to `127.0.0.1:8731` and offers a minimal `/healthz`. Repository reads require protected server-side secrets and an explicit, trusted Forgejo HTTPS origin; see [Development manual](docs/USER-MANUAL.md). **No Code write API exists** in this seed.

## Governance and source origin

The approved source baseline is `v15.0.9` LTS. GitHub readback confirmed the imported `vendor/forgejo` tracked source, `LICENSE` and `UPSTREAM-SNAPSHOT.md`; upstream SHA: `19b9b9d216bbfb501c18514bd1a8c980246ca3f7`. Build, security and deployment acceptance remain outstanding. Retain Forgejo GPL-3.0-or-later licensing, notices and attribution. Source snapshots are not full-history mirrors.

Glaze current-release authority: [GoreeCloud/glaze](https://github.com/GoreeCloud/glaze). Product integration requires exact-Stable consumer acceptance. Validate the applicable GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze, Mesh, Identity, Policy, and Observability contracts; no unimplemented integration should be described as functioning.

**Read next:** [specifications](docs/SPECIFICATIONS.md) · [architecture](docs/ARCHITECTURE.md) · [upstream source](docs/UPSTREAM.md) · [planned features](docs/PLANNED-FEATURES.md) · [implemented features](docs/IMPLEMENTED-FEATURES.md) · [validation](docs/VALIDATION.md) · [security](.github/SECURITY.md).

Do not merge, deploy or declare Stable until the required exact-head build, test, security, privacy, source-provenance, Glaze, migration and recovery gates have been verified.

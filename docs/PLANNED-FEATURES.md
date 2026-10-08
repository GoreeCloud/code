# Planned features and remaining work

Every row is **not yet verified as completed**, unless separately reflected in IMPLEMENTED-FEATURES.md with evidence.

| Priority | Workstream | Required acceptance evidence |
| --- | --- | --- |
| P0 | Complete imported-source acceptance | Tracked source and SHA are present; Forgejo Go executable build has passed once; independent provenance, license audit, reproducibility and security acceptance still pending |
| P0 | Secure build, SBOM and dependency vulnerability scan | Reproducible build/test output, provenance attestations, scoped CI identity |
| P0 | Identity, authorization, Policy and Wardveil | Real scope-denial tests, actor attribution, pre-execution durable audit |
| P0 | Restore and operational recovery (Everkeep) | Repository and metadata restore to representative target |
| P1 | Glaze exact-Stable migration | Consumer acceptance, keyboard/screen-reader, contrast, reduced motion |
| P1 | Complete repository, issue, branch, PR, review and release parity | Read-only summary source exists; full workflow, live-provider and governed write acceptance remains pending |
| P1 | Governed mutations and idempotency | Replay and uncertain-outcome reconciliation tests |
| P1 | Containers-isolated pipelines and packages | Runner sandbox, supply-chain attestations and quota enforcement |
| P1 | GoreeCloud AI coding and review | Explicit approval and least-privilege capability broker |
| P1 | GitHub-to-Code migration | Reconciled issues, PR reviews, settings, permissions and metadata |
| P2 | Search, performance and usability | p95/p99 baselines, accessibility and load validation |
| P2 | Offline-first collaboration and first-party clients | Conflict-resolution, sync, permissions and device tests |

Do not promote a feature on the basis of README text, UI mockups or a planned workflow.

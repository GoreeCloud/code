# GoreeCloud Code — product specification

**Lifecycle:** Development. **Destination:** `GoreeCloud/code`. **Product identity:** GoreeCloud Code. **Path:** Fork-to-Native from Forgejo.

The platform will provide Git hosting, repository administration, branches, commits, issues, pull requests, reviews, packages, release management, portable CI/CD, developer collaboration, and optional AI-assisted engineering. It will own GoreeCloud-facing APIs and experiences rather than exposing Forgejo implementation details as permanent product contracts.

## Foundations and boundaries

- Standard Git is the repository format. Preserve import/export for issues, reviews, permissions, releases, workflows, packages and other non-Git metadata.
- Treat Forgejo as an upstream-derived, replaceable implementation. Preserve attribution and license compliance throughout source ingestion, modifications and distribution.
- Use the approved Glaze release from the `GoreeCloud/glaze` lifecycle registry. Product-specific accessible consumer acceptance remains required; a design preview is not conformance.
- Never expose Forgejo administrator/service tokens to browsers. Identity sessions, per-resource Code authorization, policy decisions, and auditable capabilities must precede mutation support.
- Evaluate all nine Integral Platform Systems: GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze, Mesh, Identity, Policy and Observability. Integration claims require running evidence.
- Source-import, build, tests, dependency inventory, security scanning, isolated runners, reversible deployment, backup/restore and migration acceptance precede a Stable claim.

## Initial implementation

An opt-in, loopback-bound read-only Code API and static design preview establish isolated testable foundations. A branch-scoped workflow will attempt a pinned upstream source import. Neither a prepared workflow nor a passing unit test establishes live Forgejo interoperability or complete implementation.

## Roadmap and decision

Prefer Forgejo **v15.0.9 LTS** for the initial reviewed source baseline due to its support window. Reassess compatibility/security against current upstream releases at deployment time. Transition toward native product-defined systems in controlled increments; retain stable, security-critical proven components when replacement is not justified.

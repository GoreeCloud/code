# Validation and evidence record

## Verified source acquisition

- `GoreeCloud/code` is the correct repository, default branch `main`, original base `f0f40ef5f467722f7a01059ca7b559484133c62f`.
- Bootstrap branch `feat/forgejo-bootstrap` contains the imported `vendor/forgejo` subtree.
- Forgejo importer completed with success: https://github.com/GoreeCloud/code/actions/runs/37716860233
- Source import commit: `77be991b24ec80a498c086b67a6975bb12fb38fd`.
- Manifest readback: `v15.0.9` / official upstream SHA `19b9b9d216bbfb501c18514bd1a8c980246ca3f7` from https://codeberg.org/forgejo/forgejo.git
- Upstream `LICENSE`, `go.mod` and source tree were visible during GitHub readback.
- GitHub Code CI on pre-manifest candidate `884e472f4f879723320fb65596df2c3b9eccac15` succeeded: https://github.com/GoreeCloud/code/actions/runs/37716860156

## Gates still required

1. Re-run exact-head CI after source import and latest validation updates, including `scripts/verify-forgejo-snapshot.mjs`.
2. Validate the imported tag's upstream provenance and license notices independently; run Forgejo source build, reproducible image, static analysis, dependency and security scans.
3. Execute real Forgejo API interoperability, repository permissions and migration tests.
4. Complete GoreeCloud Code Identity/Policy/Wardveil and all nine Integral Platform System consumer assessments.
5. Complete exact current Glaze Stable adoption and accessibility acceptance.
6. Verify Everkeep backup/restore, deployment protections, load tests and rollout/rollback drills.

Passing a Node test or a source tree check does not prove production runtime behavior, full license review, security acceptance, release readiness or Stable.

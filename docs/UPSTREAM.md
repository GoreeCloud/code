# Forgejo source provenance and license

Source authority: https://codeberg.org/forgejo/forgejo

**Initial target:** `v15.0.9` LTS, released 17 September 2026 and supported through 15 July 2027. A v16.0.5 stable candidate exists but its support window ends 29 October 2026. Re-evaluate the pin against current security notices before deploying.

Forgejo is licensed GPL-3.0-or-later. Upstream attribution and all embedded license/notice files must remain intact. If GoreeCloud distributes modified Forgejo binaries, it must comply with relevant copyleft and corresponding-source obligations. Do not remove origin credits or assert authorship of upstream code.

The importer produces a tracked **source snapshot** with a machine-readable commit manifest; it does not copy upstream Git commit history, build upstream code, certify dependencies, or establish Stable readiness. Security reviews, build acceptance, migration tests and rollback validation remain separate obligations.

Upstream changes must be reviewed and intentionally backported rather than silently auto-synchronized. Retain the upstream commit hash for reproducibility and migration traceability.

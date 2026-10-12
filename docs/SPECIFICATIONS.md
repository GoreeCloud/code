# GoreeCloud Code — specifications

**Development approach:** Fork-To-Native. **Lifecycle:** source baseline, not production ready.

**Purpose:** A GoreeCloud-controlled Git collaboration platform that incrementally preserves/improves Forgejo's hosted repositories, identity/permissions, organization administration, issues, pull requests, code review, CI workflows, artifacts, releases, packages, and service administration.

**Initial implementation:** Imported Forgejo v16.0.5 source at `8d24b7d52df327118cd776cf461d34e52e0ec1ce`. The codebase is the authoritative inherited functional baseline, not evidence of a completed GoreeCloud deployment.

**Required safeguards:** retain source/DB compatibility, APIs, authentication and data protections, backups/restore, accessibility and core workflow functionality. Maintain license notices and corresponding source obligations.

**Native transformation gates:** inventory features, establish functional/build/security baselines, isolate upstream implementation boundaries, adopt approved Glaze and product identity, integrate applicable nine GoreeCloud Integral Platform Systems, validate functionality/migrations, and provide reversible upgrades. Do not retire inherited components without verified parity.

**Evidence gaps:** no GoreeCloud-specific acceptance tests, production security audit, verified platform contracts, user-facing design review, restore/upgrade validation, or deployment confirmation at import time.

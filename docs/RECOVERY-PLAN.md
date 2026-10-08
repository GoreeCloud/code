# Recovery, portability and cutover plan

**Status:** Planned; no Everkeep restore acceptance yet.

Inventory both Git content and non-Git forge state: databases, users, organizations, repository permissions, issues, pull requests, reviews, comments, branch rules, releases, packages, CI secrets/references, artifacts, webhooks, audit evidence and configuration.

Before migration: capture encrypted access-controlled backups; validate checksums, retention and restore credentials; verify restore of a representative populated instance in an isolated test environment.

For cutover: establish a read-only source freeze or reconcile concurrent changes; map IDs and actors; import Git refs and non-Git metadata; compare counts, hashes and permission restrictions; test authenticated reads and writes through GoreeCloud Code policy; execute reversible traffic switch and rollback rehearsals.

Do not retire an old Forgejo or GitHub provider until Git and collaboration metadata are proven present, policy boundaries are sound, and at least one real recovery rehearsal succeeds. Backups existing without restore tests do not establish recovery readiness.

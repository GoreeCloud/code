## Change and risk scope

Explain what changes, why it is needed, and whether it affects upstream Forgejo, Code APIs, authentication, authorization, Glaze, migration, data, security, privacy, CI, or recovery.

## Verification

- [ ] Current intended base and exact head SHA checked
- [ ] Required source-level tests, lint and build checks passed
- [ ] Upstream provenance and license notices preserved, if relevant
- [ ] Security/privacy risks and credential scopes reviewed
- [ ] Required Glaze accessibility and responsive behavior evaluated
- [ ] Repository and non-Git metadata migration/rollback considered
- [ ] Documentation and planned/implemented feature lifecycle reconciled
- [ ] Runtime, deployment, backup/restore acceptance attached where claimed

**Status:** Development changes must never be described as Stable solely because CI passes. Protected-branch rules and appropriate reviewer decisions are mandatory before merge.

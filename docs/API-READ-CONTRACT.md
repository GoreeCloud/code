# Code API — development read contract

**Authority:** Source-level prototype only, not a production identity or platform-contract acceptance.

The service binds only to loopback and exposes `GET /healthz` plus four gated read routes:

| Route | Result |
| --- | --- |
| `/api/v1/repositories/{owner}/{repo}` | sanitized repository summary |
| `/api/v1/repositories/{owner}/{repo}/branches` | at most 20 branch summaries |
| `/api/v1/repositories/{owner}/{repo}/issues` | at most 20 open issue summaries |
| `/api/v1/repositories/{owner}/{repo}/pulls` | at most 20 open pull request summaries |

All repository routes require the interim application bearer credential, an owner-only provider token, an explicit HTTPS Forgejo origin and an owner-only `GOREECLOUD_CODE_ALLOWED_REPOSITORIES_FILE` JSON file with exact names, e.g. `{"repositories":["GoreeCloud/example"]}`. Unknown repositories are denied before provider requests.

Arbitrary query parameters, provider-chosen redirects, uncontrolled upstream response fields and writes are not available. The policy is refreshed per request. User identities, roles, Wardveil decisions, audit, migration and real deployment tests remain pending.

For implementation details, see `services/api/`; for operational setup, see `USER-MANUAL.md`.

# GoreeCloud Code validation

## Verified before this bootstrap

The repository `GoreeCloud/code` existed with the original README and `main` commit `f0f40ef5f467722f7a01059ca7b559484133c62f`. The feature branch was created from that commit. The canonical Glaze `VERSION` was `1.7.0` at inspection. All new Code feature claims require exact-head validation.

## Validation gates

- Run `node --check services/api/server.mjs` and `node --test services/api/server.test.mjs`.
- Run `bash -n scripts/import-forgejo.sh` without executing the remote clone during lint.
- Confirm the GitHub Actions checks on the exact final head.
- Confirm `vendor/forgejo/LICENSE`, upstream manifest, resolved SHA and full source before claiming successful import.
- Subsequently run representative Forgejo build, DB migration and live API tests.
- Complete nine Integral Platform System applicability review and exact Glaze Stable consumer acceptance.
- Do not confuse passing source tests with runtime deploy, migration, recovery or Stable.

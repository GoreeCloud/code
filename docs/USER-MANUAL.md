# GoreeCloud Code development manual

The platform is **not yet deployed**.

## Static preview

Open `apps/web/index.html` locally to inspect a non-functional design preview. No live repositories, authentication or modifications are available through it.

## Run the read-only development API

Requires Node.js 22+. Execute `node services/api/server.mjs`. It binds to `127.0.0.1:8731` only.

`GET /healthz` returns a development-state health indicator. For `GET /api/v1/repositories/{owner}/{name}`, a trusted operator must configure:

- `GOREECLOUD_CODE_FORGEJO_URL` — HTTPS root origin, no path, query or embedded credentials.
- `GOREECLOUD_CODE_ACCESS_TOKEN_FILE` — owner-only regular secret file with a development bearer secret.
- `GOREECLOUD_CODE_FORGEJO_TOKEN_FILE` — owner-only regular file with a narrowly scoped upstream read token.
- `GOREECLOUD_CODE_ALLOWED_REPOSITORIES_FILE` — owner-only JSON file with explicitly approved repository names (for example `{ "repositories": ["GoreeCloud/example"] }`).

Files must be 0600, owned by the process user and stored outside Git. Send `Authorization: Bearer <application-secret>` on the repository route. Do not publish this interim API as a production service. The endpoint intentionally omits most upstream fields and rejects writes.

## Source import

The proposed branch-bounded importer uses Forgejo v15.0.9 LTS. Import does not certify builds, security, licenses, migrations or deployment. Inspect its provenance manifest and CI evidence before using upstream code.

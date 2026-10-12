# GoreeCloud Code

**Status:** Fork-to-Native source baseline; active development, **not** a GoreeCloud-native or production-validated release.

GoreeCloud Code is GoreeCloud's self-hosted source-code collaboration and developer-platform project. Its initial source foundation is [Forgejo](https://codeberg.org/forgejo/forgejo), with a staged migration toward an independently maintained GoreeCloud product experience, architecture, security/privacy controls, and platform integrations.

## Source baseline

| Item | Value |
| --- | --- |
| Upstream | https://codeberg.org/forgejo/forgejo |
| Imported release | `v16.0.5` (17 September 2026) |
| Upstream commit | `8d24b7d52df327118cd776cf461d34e52e0ec1ce` |
| Import method | Original upstream Git history merged into the independent `GoreeCloud/code` repository |
| License | GPL-3.0-or-later for the imported Forgejo code, subject to component-level review |
| Development approach | Fork-To-Native, phase 1: source establishment |

The imported source includes existing Forgejo Git hosting, issues, pull requests, organizations, wiki, releases, packages, and CI/action capabilities. These are **inherited source capabilities**, not claims of a validated GoreeCloud deployment.

## Development and documentation

- [Specifications](docs/SPECIFICATIONS.md) and [features](docs/FEATURES.md)
- [Implemented/inherited capabilities](docs/IMPLEMENTED-FEATURES.md) and [planned migration](docs/PLANNED-FEATURES.md)
- [Branding](docs/BRANDING.md), [privacy status](docs/PRIVACY.md), and [user manual](docs/USER-MANUAL.md)
- [Provenance and upstream synchronization](docs/UPSTREAM.md) and [changelog](docs/CHANGELOGS.md)

Developers should follow the upstream [Forgejo source build instructions](https://forgejo.org/docs/latest/contributor/from-source/) for this pinned baseline and revalidate dependencies, tests, and configuration before deploying. Do not expose an unconfigured development instance on a public network.

## Licensing and acknowledgements

Forgejo and its contributors developed the imported application. Their copyrights, GPL license, notices, and applicable dependency licenses remain in force. See [LICENSE](LICENSE), [upstream provenance](docs/UPSTREAM.md), and [the preserved upstream README](docs/UPSTREAM-README.md). Rebranding does not change authorship or remove upstream legal obligations.

**Not yet verified:** GoreeCloud branding/Glaze runtime implementation; GoreeCloud Identity, Policy, Manager, Privacy Shield, Wardveil, Everkeep, Mesh, or Observability integration; migration, backup/restore, security acceptance, end-to-end operation, or production deployment.

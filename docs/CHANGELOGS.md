# GoreeCloud Code changelog

## Unreleased — Development

- Added Forgejo Fork-to-Native product architecture, initial Code API, tests, security controls, Glaze-oriented static preview and repository governance records.
- Imported official **Forgejo v15.0.9 LTS** tracked source into `vendor/forgejo`, with exact upstream SHA and license-preserving manifest; see [source provenance](UPSTREAM.md) and [validation](VALIDATION.md).
- Added source integrity checks and nine-system platform contract declaring current integration gaps.
- Extended the read-only Code API with owner-only repository allowlisting, branch/issue/PR summaries, identity-bound repository projections, and regression coverage for denied access and untrusted provider fields.

This is not an accepted binary build, a production deployment, certified Glaze application, or Stable release.

# Implemented at the repository-source level

- Development-only loopback Code HTTP server, including a no-secret health route.
- A single authenticated, read-only provider-neutral repository summary route.
- Controlled server-side Forgejo credentials loaded from owner-only files; fail-closed configuration and sanitized responses.
- Read-only request scope with explicit refusal of writes and a bounded provider response.
- Deterministic local test suite and CI validation **definition**; CI pass is separate evidence.
- Branch-bounded Forgejo source-import **definition**; import is not completed until `vendor/forgejo` is verified in GitHub after workflow execution.

Not yet implemented: full Forgejo source fork, publicly deployed forge, real live read acceptance, first-party Identity/Policy/Wardveil integration, complete Glaze adoption, CI runners, metadata migration and Stable qualification.

# Security — GoreeCloud Code

**Status: Development. Not security-accepted or production-ready.**

Report vulnerabilities privately through the repository's private vulnerability reporting where enabled, or an authorized private GoreeCloud security channel. Do not publish credentials or exploitable internal details in public issues.

The current Code API is loopback-only, read-only, bearer-gated and restricts provider credentials to the server. A token file is an interim development mechanism, **not** a replacement for GoreeCloud Identity, Code authorization, Wardveil, or Policy.

Before enabling writes: require exact actor and scope, default-deny per-resource policy, durable pre-execution audit evidence, idempotency and uncertainty reconciliation, credential minimization, abuse throttling, SSRF controls, recovery drills, vulnerability testing and authorization-negative tests. Do not expose the local API directly to the internet.

Vendored upstream source and CI runners are untrusted until provenance, license, dependency, execution and sandbox guarantees have been independently assessed.

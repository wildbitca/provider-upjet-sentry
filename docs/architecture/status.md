# Status and open items

All 24 resources of `jianyuan/sentry` v0.15.8 are implemented in seven API groups and two API trees (cluster and namespaced). Generate is idempotent; build, lint, tests, the guardrail scripts and the family package build pass. Evidence: [test-report.md](test-report.md).

## Not verified yet

Nothing has run against a live Sentry:

- token scopes needed per resource;
- composite versus bare id handling at `Read` (see [external-names.md](external-names.md));
- the id-less singletons `ProjectOwnership` and `AllProjectsSpikeProtection` (best effort until verified);
- `jsontypes` attribute diffs;
- how often the `production` environment protection fires per run.

Run one create/observe cycle per external-name class (bare, composite, parameter-driven, id-less) before tagging `v0.1.0`.

## Repository settings still to configure

See [release-and-security.md](release-and-security.md): branch and tag protection, `production` environment restricted to protected refs, code owner review, private vulnerability reporting, Dependabot and secret scanning with push protection.

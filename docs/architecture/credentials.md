# Credentials

## Secret format

A single Secret key (conventionally `credentials`) holds a JSON object:

```json
{"token": "sntrys_...", "base_url": "https://sentry.io/api/"}
```

| Key | Required | Meaning |
|---|---|---|
| `token` | yes | Organization Auth Token or Internal Integration token. Sensitive. |
| `base_url` | no | API root; must be an `https://` URL with a host, without user information, query or fragment or dot segments, ending with `/api/`. Defaults to `https://sentry.io/api/`. |

Implementation: `internal/clients/upjet-sentry.go`. The keys are whitelisted in `credentialKeys`; anything else in the JSON is ignored.

Rules enforced by `buildConfiguration` (unit tests in `upjet-sentry_test.go`):

- **Empty values are skipped**, so a blank `base_url` never overrides the provider default or the `SENTRY_BASE_URL` environment fallback with `""`.
- A missing or empty `token` is an error. Upstream's Configure health check fails every operation without a token, so failing early gives an actionable message instead of a stream of opaque tofu errors.
- A `base_url` that is not a plain `https://` URL ending with `/api/` (host required; no user information, query, fragment or dot segments) is an error with the expected form in the message. Plain `http://` is rejected on purpose because the token travels in every request; relaxing this later would be non-breaking.
- Error messages never contain credential values (tested with a canary token).

Region URLs: `https://us.sentry.io/api/`, `https://de.sentry.io/api/`. Self-hosted: `https://<your-host>/api/`.

## ProviderConfig kinds

| Tree | Kinds |
|---|---|
| cluster (`upjet-sentry.upbound.io`) | `ProviderConfig` (cluster scope), `ProviderConfigUsage` |
| namespaced (`upjet-sentry.m.upbound.io`) | `ProviderConfig` (namespaced; the Secret is read from the managed resource's namespace), `ClusterProviderConfig`, `ProviderConfigUsage` |

`ProviderConfig` types live in the root group and ship only in the family package `provider-family-sentry`. The credentials `source` enum is the Crossplane standard (`None`, `Secret`, `InjectedIdentity`, `Environment`, `Filesystem`); only `Secret` is documented and tested.

## Token scopes

**Unverified.** No live Sentry token was available. The Sentry API documents these scope families; the mapping below is expected, not tested:

| Group | Expected scopes |
|---|---|
| organization | `org:admin`, `member:admin` |
| team | `team:admin`, `member:admin` |
| project | `project:admin`, `team:read` |
| alert, monitor | `alerts:write`, `project:write` |
| integration | `org:integrations`, `project:write` |
| dashboard | `org:write` |

Recommendation: one `ProviderConfig` and one token per Sentry organization, with only the scopes of the groups you install.

## Operational notes

- The upstream provider runs a health check against Sentry each time it is configured, which costs one extra HTTP call per reconcile of each resource. There is no code workaround; lengthen `--poll` (see README).
- The credentials JSON is never logged by this provider.
- In CLI mode the OpenTofu state, including sensitive attributes such as a key's DSN secret, lives in the provider pod's working directory. Keep that filesystem non-persistent and do not mount a volume there.
- Connection details: `ProjectKey` publishes `secret` and the `dsn` map into its connection secret. `ProjectKey` deliberately omits the deprecated `dsn_secret`, `dsn_public` and `dsn_csp` attributes.
- Sensitive inputs are secret references: `ProjectSymbolSource` (`appConnectPrivateKey`, `password`, `privateKey`, `secretKey`), `IntegrationOpsgenie` and `IntegrationPagerduty` (`integrationKey`), `Project` (`clientSecurity.securityToken`).

The host of `base_url` is deliberately not allow-listed: whoever can write the credentials Secret already holds the token, so pointing it at another host (a regional or self-hosted Sentry) is the intended use.

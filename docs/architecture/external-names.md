# External names

Implemented in `config/external_name.go`; the map doubles as the include list, so all 24 resources must appear there (`TestIncludeListHasAllResources`).

Upjet writes a synthetic Terraform state before each refresh, with `id` set to the output of `GetIDFn` and all parameters copied in. The rule that follows from spike S1 (`spikes.md`): **`GetIDFn` returns exactly the `id` that the provider itself stores in state.** The provider's `Read` uses that `id`, plus the `organization` / `project` attributes taken from the parameters.

## Import ids (`GetImportIDFn`)

`terraform import` (the Observe-only and adoption path) is a different contract from the state id. Since upjet 2.4.0 the string passed to `terraform import` comes from `ExternalName.GetImportIDFn` and falls back to `GetIDFn` when it is nil. Several `jianyuan/sentry` importers require a composite string and then store a bare id, so importing with the bare id fails (the v0.1.0 bug). `withImportID(strategy, keys...)` in `config/external_name.go` sets `GetImportIDFn` to `<parameters in keys order>/<external name>` and leaves `GetIDFn` bare. While the external name is empty it returns an empty id, which upjet treats as "does not exist yet" and sends to Create.

Verified against the importers of `jianyuan/sentry` v0.15.8 (`internal/provider/resource_*.go`, `sentry/importer.go`):

| Terraform resource | Import string | State id after import |
|---|---|---|
| `sentry_team`, `sentry_project` | `org/slug` | `slug` |
| `sentry_organization_code_mapping`, `sentry_alert`, `sentry_cron_monitor`, `sentry_metric_monitor`, `sentry_uptime_monitor`, `sentry_notification_action` | `org/id` | `id` |
| `sentry_key`, `sentry_issue_alert`, `sentry_project_symbol_source` | `org/project/id` | `id` |
| `sentry_integration_opsgenie`, `sentry_integration_pagerduty` | `org/integration_id/id` | `id` |
| `sentry_organization_repository` | `org/integration_type/integration_id/id` | `id` |
| `sentry_plugin` | `org/project/plugin` | `plugin` |

`sentry_organization`, `sentry_organization_member`, `sentry_dashboard`, `sentry_metric_alert`, `sentry_team_member` and the `sentry_project_*` singletons already use the same string for import and state, so they keep `GetImportIDFn` nil.

## Strategies

| Strategy | Used by | `GetIDFn` | `GetExternalNameFn` |
|---|---|---|---|
| `slugAsIdentifier` | `sentry_organization`, `sentry_team`, `sentry_project` | external name | state `id`; `slug` is seeded from the external name and omitted from `forProvider` |
| `config.IdentifierFromProvider` | 12 resources whose state `id` is the bare provider id | external name | state `id` |
| `compositeFromProvider(tmpl)` | `sentry_organization_member` (`org/id`), `sentry_metric_alert` (`org/project/id`), `sentry_dashboard` (`org/id`) | template over `.parameters` and `.external_name` | last template part of state `id` |
| `paramDrivenID(idKeys, nameKeys)` | `sentry_plugin` (id = the `plugin` parameter), `sentry_team_member`, `sentry_project_inbound_data_filter`, `sentry_project_spike_protection`, and the id-less `sentry_project_ownership`, `sentry_all_projects_spike_protection` | parameters joined with `/` (error naming the missing parameter) | state attributes `nameKeys` joined with `/` (never from `id`) |

`DisableNameInitializer` is true for every strategy except `slugAsIdentifier`: provider-assigned ids must not default to `metadata.name`.

## Behavior to know

- **Slugs.** For organizations, teams and projects the external name is the slug and defaults to `metadata.name`. Set the `crossplane.io/external-name` annotation to adopt an existing object. Changing a project slug is destructive upstream (#725); treat it as immutable.
- **Param-driven resources** rebuild the Terraform id from the parameters each time, so the value of the external-name annotation does not change what is sent to the provider. They do not default the annotation to `metadata.name` (`DisableNameInitializer`), so adopting an existing object requires the annotation to be present (set it to the value the resource reports, for example the project slug for `ProjectSpikeProtection`); otherwise the first reconcile takes the create path.
- **Id-less resources** (`ProjectOwnership`, `AllProjectsSpikeProtection`) have no `id` in state; the import string is `org/project` and `org`. Upjet writes no `id` for them because their schema has none (`WithHasIDAttribute`).
- **Notification actions.** Failed creates can orphan actions upstream (#937, #960). Creates are not retried by this provider.

## Tests

`config/external_name_test.go` has table tests per strategy (`GetIDFn`, `GetExternalNameFn`, missing parameters). `config/external_name_test.go` also has `TestGetImportIDFn`. `config/contract_test.go` holds the stored-state-id contract for all 24 resources and checks the round trip `GetIDFn` -> state id -> `GetExternalNameFn`, plus that the pinned schema contains exactly the 24 resources, so a provider bump that adds a resource fails the build until it is configured. `TestImportIDRoundTrip` checks, for every composite importer, the exact import string, that its last segment is the stored state id, that `GetIDFn` stays bare, and that an empty external name or a missing organization is handled.

## Known limitation: empty id in the first synthetic state

For `IdentifierFromProvider` resources upjet 2.5.1 still writes `id = ""` into the first synthetic state (crossplane/upjet#789 is open). The provider then refreshes with an empty id before Create; see `spikes.md` section S6 for the measured effect.

Re-confirmed 2026-10-29 against the v0.15.8 sources: the importers of alert, cron/metric/uptime monitor use `ImportState2Part` (strict `org/id`); the registry docs show a URL form for them, which is stale.

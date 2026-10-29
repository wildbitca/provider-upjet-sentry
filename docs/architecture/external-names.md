# External names

Implemented in `config/external_name.go`; the map doubles as the include list, so all 24 resources must appear there (`TestIncludeListHasAllResources`).

Upjet writes a synthetic Terraform state before each refresh, with `id` set to the output of `GetIDFn` and all parameters copied in. The rule that follows from spike S1 (`spikes.md`): **`GetIDFn` returns exactly the `id` that the provider itself stores in state.** The provider's `Read` uses that `id`, plus the `organization` / `project` attributes taken from the parameters.

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
- **Id-less resources** (`ProjectOwnership`, `AllProjectsSpikeProtection`) have no `id` in state; the import string is `org/project` and `org`. Spike S3 explains why the empty-id synthetic state is safe on this fork.
- **Notification actions.** Failed creates can orphan actions upstream (#937, #960). Creates are not retried by this provider.

## Tests

`config/external_name_test.go` has table tests per strategy (`GetIDFn`, `GetExternalNameFn`, missing parameters). `config/contract_test.go` holds the stored-state-id contract for all 24 resources and checks the round trip `GetIDFn` -> state id -> `GetExternalNameFn`, plus that the pinned schema contains exactly the 24 resources, so a provider bump that adds a resource fails the build until it is configured.

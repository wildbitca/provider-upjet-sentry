# Resource matrix

24 managed resources of `jianyuan/sentry` v0.15.8, all generated, none deferred. Every resource exists in the cluster-scoped tree (`<group>.upjet-sentry.upbound.io`) and the namespaced tree (`<group>.upjet-sentry.m.upbound.io`), version `v1alpha1`.

Group and Kind are a **public contract**; `config/provider_test.go` fails when they change. Upjet derives the default Kind by dropping the first word after the `sentry_` prefix (`sentry_cron_monitor` would become `Monitor`), so every Kind is set explicitly.

Class: A bare provider-assigned id in state, B composite id in state, C slug is the id, D no `id` attribute (identity from parameters), P identity made of parameters only.

| Terraform resource | Group | Kind | Class | External name | Sensitive / omitted | Status |
|---|---|---|---|---|---|---|
| `sentry_organization` | organization | `Organization` | C | slug (seeded from external name) | | ok |
| `sentry_organization_member` | organization | `OrganizationMember` | B | member id; id `org/member` | | ok |
| `sentry_organization_code_mapping` | organization | `OrganizationCodeMapping` | A | provider id | | ok |
| `sentry_organization_repository` | organization | `OrganizationRepository` | A | provider id | | ok |
| `sentry_team` | team | `Team` | C | slug (seeded) | omitted `team_id` (deprecated) | ok |
| `sentry_team_member` | team | `TeamMember` | P | `member_id`; id `org/team/member` | | ok |
| `sentry_project` | project | `Project` | C | slug (seeded, immutable in practice) | `client_security.security_token` -> `securityTokenSecretRef` | ok |
| `sentry_key` | project | `ProjectKey` | A | key id | omitted `dsn_csp`, `dsn_public`, `dsn_secret`; `secret`, `dsn` -> connection details | ok |
| `sentry_plugin` | project | `ProjectPlugin` | P | `plugin` parameter (its id is the plugin name) | | ok |
| `sentry_project_ownership` | project | `ProjectOwnership` | D | `org/project` | | ok (round trip unit-tested only) |
| `sentry_project_spike_protection` | project | `ProjectSpikeProtection` | P | project slug; id `org/project` | | ok |
| `sentry_all_projects_spike_protection` | project | `AllProjectsSpikeProtection` | D | organization | | ok (round trip unit-tested only) |
| `sentry_project_inbound_data_filter` | project | `ProjectInboundDataFilter` | P | `filter_id`; id `org/project/filter` | | ok |
| `sentry_project_symbol_source` | project | `ProjectSymbolSource` | A | source id | `app_connect_private_key`, `password`, `private_key`, `secret_key` -> `*SecretRef` | ok |
| `sentry_issue_alert` | alert | `IssueAlert` | A | rule id | omitted `actions`, `conditions`, `filters` (deprecated JSON strings; use `*_v2`) | ok |
| `sentry_metric_alert` | alert | `MetricAlert` | B | rule id; id `org/project/rule` | | ok |
| `sentry_alert` | alert | `Alert` | A | alert id | omitted `action_filters.actions.plugin` (cannot be created) | ok |
| `sentry_notification_action` | alert | `NotificationAction` | A | action id | | ok (orphan risk, see README) |
| `sentry_cron_monitor` | monitor | `CronMonitor` | A | monitor id | | ok |
| `sentry_metric_monitor` | monitor | `MetricMonitor` | A | monitor id | | ok |
| `sentry_uptime_monitor` | monitor | `UptimeMonitor` | A | monitor id | | ok |
| `sentry_integration_opsgenie` | integration | `IntegrationOpsgenie` | A | service id (`team` is the Opsgenie team label, a plain string) | `integration_key` -> `integrationKeySecretRef` | ok |
| `sentry_integration_pagerduty` | integration | `IntegrationPagerduty` | A | service id | `integration_key` -> `integrationKeySecretRef` | ok |
| `sentry_dashboard` | dashboard | `Dashboard` | B | dashboard id; id `org/dashboard` | | ok |

## References

| Field | Target | Notes |
|---|---|---|
| `organization` (all resources except `Organization`) | `Organization` | Optional: a plain slug keeps working |
| `project` / `projects` | `Project` | Slug is the id |
| `teams` (`Project`), `team` (`TeamMember`) | `Team` | Slug is the id. `IntegrationOpsgenie.team` is an Opsgenie label, not a Sentry team: no reference |
| `member_id` (`TeamMember`) | `OrganizationMember` | External name is the member id |
| `owner.team_id` (monitors) | `Team` | Extracts the numeric `internal_id` from `status.atProvider` |

`OrganizationCodeMapping.project_id` takes the numeric project id. It is **not** a reference: `Project` already references `Organization`, and a reference from the `organization` group back to `project` is an import cycle in the generated Go packages.

`integration_id`, `repository_id`, `monitor_ids` and user ids have no managed target and stay plain strings.

`config/provider_test.go` (`TestReferences`, `TestReferenceGroupsAreAcyclic`) and `config/crd_smoke_test.go` (`TestReferenceFieldsExistInCRDs`) pin this table, including the fields that must stay plain.

## Verification status of the identity classes

Class A (bare provider id) and class B (composite) follow the stored state ids in the upstream research table, checked in `config/contract_test.go` as a round trip for all 24 resources. Whether the live provider accepts the synthetic state at `Read` is **unverified** for every class.

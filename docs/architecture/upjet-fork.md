# Upjet fork pin

This provider builds against upstream Upjet v2.5.1 plus a small set of patches carried in the fork `github.com/wildbitca/upjet`, pinned in `go.mod`:

```
replace github.com/crossplane/upjet/v2 => github.com/wildbitca/upjet/v2 v2.5.1-wb.1
```

Tag `v2.5.1-wb.1` is upstream v2.5.1 with:

- numeric state ids (Terraform ids that are numbers stay numbers in the synthetic state);
- no CEL "required parameter" rule for required dynamic parameters;
- empty external-name guard: no synthetic state is written when the external id is empty (`EnsureTFState` returns early when `tfID == ""`).

## Why the pin is required here

The empty-id guard matters for the 12 `IdentifierFromProvider` resources. Without it, the first reconcile writes `id = ""` into the state, `tofu apply` refreshes first and calls e.g. `GET /api/0/projects/<org>/<project>/keys//`, and fails before Create is reached (see `spikes.md` S6, upstream issue crossplane/upjet#789).

## Direction

Upstream-pure is the long-term goal: when the guard lands upstream, bump Upjet and drop the `replace`. No upstream PR is open for now. The composite import-id fix (`ExternalName.GetImportIDFn`, see `external-names.md`) uses upstream API only and does not depend on the fork.

# Gumroad release flow

This document describes the proposed Gumroad publishing integration for Fe. It is a design contract, not a description of functionality that exists today.

Fe currently builds and packages release artifacts, writes an artifact manifest, and can publish to itch.io through Butler. Gumroad should be an optional destination in that pipeline; it should not introduce a second build or packaging system.

## Goals

- Publish stable, versioned software releases through Gumroad.
- Maintain one persistent Gumroad product for rolling builds.
- Reuse the artifacts and manifest produced by Fe's existing release pipeline.
- Keep Gumroad optional: projects that do not sell through Gumroad should not need Gumroad configuration or credentials.
- Make publishing explicit, previewable, and safe to retry.
- Use documented Gumroad interfaces only. Do not scrape the dashboard or call undocumented upload endpoints.

## Product model

Fe should support two product roles:

| Role | Gumroad product lifecycle | Release behavior |
|---|---|---|
| Stable | A separate product for each stable version | Publish that version's release artifacts to its matching product |
| Rolling | One persistent product | Publish the newest rolling build to the same product |

The product IDs are owned by the Gumroad account and should be configured explicitly. Fe must not guess which product to update based on its display name.

A project may configure only one role, both roles, or neither. Stable releases should never overwrite the rolling product, and rolling releases should never mutate a stable product.

## Proposed command surface

The names below are proposals, not implemented commands.

- `fe gumroad products` — list products available to the authenticated account and show their IDs.
- `fe gumroad release` — publish a stable release to the configured product for that version.
- `fe gumroad rolling` — publish a rolling build to the configured rolling product.
- `fe release --gumroad` — include Gumroad as an explicitly selected destination in the normal release pipeline.

The Gumroad commands should also support `--dry-run`. A dry run must show the selected product, version/channel, artifact files, and any unsupported operation without modifying Gumroad.

Publishing to Gumroad should remain independently selectable from itch.io. Do not make `fe release` require Gumroad credentials when Gumroad is not selected.

## Configuration direction

Illustrative only; this schema is not implemented or finalized:

```yaml
gumroad:
  enabled: false
  # Product IDs are created in Gumroad and copied from the dashboard/API.
  stable:
    # "1.2.0": "PRODUCT_ID_FOR_1_2_0"
  rolling:
    product_id: "ROLLING_PRODUCT_ID"
```

Credentials must come from the environment or a secure credential store, never from committed `.fe.yaml`:

```sh
GUMROAD_ACCESS_TOKEN=...
```

Fe must never print the token, include it in error output, or write it to the artifact manifest. The exact authentication mechanism should follow Gumroad's current API documentation.

## Release pipeline integration

The intended order is:

1. Run the existing Fe preflight and build/export steps.
2. Run optional asset packing.
3. Produce the release artifacts and `artifact-manifest.json`.
4. Resolve the selected Gumroad product from explicit configuration.
5. Validate that every artifact can be delivered through a supported Gumroad operation.
6. Publish only when the user selected Gumroad and the delivery operation is supported.
7. Report the product URL/ID, version, uploaded files, and any skipped destinations.

The existing build output and manifest are the source of truth. Gumroad integration must not rebuild, rename, or silently omit platform artifacts.

## Gumroad API capabilities and limitations

Research checked against Gumroad's own help documentation on 2026-10-31:

### REST API v2

Gumroad's help center currently says the REST API cannot create products or upload product content. Product setup and downloadable content therefore cannot be treated as fully automatable through the REST API.

- Product creation must be done outside the REST API.
- Product content files cannot be uploaded through the REST API.
- Do not invent endpoints such as `POST /v2/products/:id/files`.
- Do not report a release as successfully delivered merely because product metadata was updated.

Official reference: [Gumroad license keys and API FAQ](https://gumroad.com/help/article/76-license-keys).

### Gumroad MCP

Gumroad now documents an MCP connector for supported assistants. Its help article says the connector can create unpublished product drafts, edit product details, attach files from public URLs, list attached files, and publish/unpublish products. Files added this way are copied into Gumroad and are limited to 250 MB per file.

However, the documented connector is assistant-specific (currently Claude; Gumroad says ChatGPT is not available yet). This is not the same thing as a generally documented, stable CLI API endpoint for Fe. The public documentation also does not establish a safe file-replacement/removal operation suitable for updating a single rolling product in place.

Fe must not assume that it can call the assistant connector as a normal REST API. Revisit this option only if Gumroad publishes a supported machine-to-machine interface and documents the required product-content lifecycle operations.

Official references:
- [Connecting AI assistants to your Gumroad store](https://gumroad.com/help/article/355-connect-ai-assistants-to-gumroad)
- [Gumroad MCP](https://gumroad.com/mcp)

## Implementation stages

### Stage 1 — Document and inspect

- Keep this document as the source of truth for the intended product model and current API limits.
- Add a read-only product-listing command only after its authentication and response contract are verified against the official REST API docs.
- Add validation for configured product IDs without attempting to upload files.

### Stage 2 — Supported delivery

Do not start this stage until Gumroad provides a documented way for Fe to attach release files and replace/remove old rolling files, or an officially supported MCP interface that Fe can use non-interactively.

When that contract exists:

- Implement a small Gumroad client behind a focused package.
- Add stable and rolling commands.
- Add `--dry-run` and clear partial-failure reporting.
- Keep the client independent from the core build pipeline.
- Add tests for request construction, authentication redaction, product mapping, and retry behavior.

### Stage 3 — Release integration

- Add Gumroad as an optional destination to `fe release`.
- Keep itch.io and Gumroad failures independently visible.
- Do not claim overall release success if a selected destination failed.
- Allow publishing existing artifacts without rebuilding, consistent with Fe's release design.

## Safety and failure behavior

- Never create or modify a product unless its ID is explicitly configured or selected by the user.
- Never delete a product as part of a release.
- Never overwrite a stable product with rolling artifacts, or vice versa.
- Validate artifact paths and sizes before attempting delivery.
- Treat a timeout as an unknown result, not proof that nothing was published. Reconcile product state before retrying a potentially non-idempotent operation.
- Show which files were delivered and which were not.
- A missing Gumroad capability must produce a clear unsupported-operation error, not a fallback to browser automation or an undocumented endpoint.

## Open questions

These need answers before the upload implementation is designed:

1. Does Gumroad provide a documented non-interactive API/MCP operation for replacing or removing product content files?
2. Should each stable product contain one combined archive or separate per-platform downloads?
3. Should the rolling product expose only the latest build, or retain a visible history of prior rolling builds?
4. What artifact hosting source will provide public URLs if the supported Gumroad workflow requires URL-based file attachment?

Until those are answered by documented capabilities and an explicit product policy, Fe should not claim to support automated Gumroad uploads.

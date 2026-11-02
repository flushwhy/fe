# Gumroad release flow

This document describes the proposed Gumroad publishing integration for Fe. It is a design contract, not a description of functionality that exists today.

Fe should reuse its existing release artifacts and manifest. Gumroad is an optional destination, not another build or packaging system.

## Product model

| Role | Gumroad product lifecycle | Release behavior |
|---|---|---|
| Stable | One product per stable version | Upload that version's artifacts to its matching product |
| Rolling | One persistent product | Replace the product's current files with the newest rolling artifacts |

Product IDs are configured explicitly. Fe must never guess a target product from its display name.

### Rolling update policy

Use Gumroad's existing product content as the rolling download location:

1. Keep one Gumroad product for the rolling channel.
2. For each rolling release, prepare the same named platform artifacts as the previous release.
3. Replace the current files in that product's Content editor; do not append every build as another download and do not create a new Gumroad product version for each build.
4. Verify the new files in the buyer-facing content page before announcing the release.
5. Keep historical builds in Fe's normal release artifact store, not in the rolling product.

Gumroad says buyers can access the product's up-to-date files from their Library. This gives the rolling product a single current download set while keeping release history outside Gumroad. See [I'm not receiving updates](https://gumroad.com/help/article/211-im-not-receiving-updates) and [The Gumroad Library](https://gumroad.com/help/article/247-what-your-customers-see).

This is a **manual Gumroad step** until Gumroad documents a non-interactive file replacement API that Fe can safely call. Do not substitute public GitHub download links for paid Gumroad files: that would bypass Gumroad's content access.

Stable products follow the same manual upload process, but their content is immutable after publication; a new stable version gets a new product.

## Proposed command surface

These names are proposals, not implemented commands:

- `fe gumroad products` — list products available to the authenticated account and show IDs.
- `fe gumroad release` — prepare a stable release for its configured product.
- `fe gumroad rolling` — prepare a rolling release for the persistent product.
- `fe release --gumroad` — include the Gumroad handoff in the normal release pipeline.

The preparation commands should support `--dry-run`. Until upload is supported, they must clearly say that the files still need to be replaced/uploaded in Gumroad; they must never report a successful Gumroad publication.

Publishing to Gumroad remains independently selectable from itch.io. Projects that do not use Gumroad need no Gumroad configuration or credentials.

## Configuration direction

Illustrative only; not implemented or finalized:

```yaml
gumroad:
  enabled: false
  stable:
    # "1.2.0": "PRODUCT_ID_FOR_1_2_0"
  rolling:
    product_id: "ROLLING_PRODUCT_ID"
```

Credentials must come from the environment or a secure credential store, never committed `.fe.yaml`. The eventual client must redact credentials from logs and errors.

## Release pipeline integration

1. Run the existing Fe preflight and build/export steps.
2. Run optional asset packing.
3. Produce the release artifacts and `artifact-manifest.json`.
4. Resolve the selected Gumroad product from explicit configuration.
5. Validate the artifact set and show the exact files intended for Gumroad.
6. Until automated file delivery is officially supported, stop at a clear manual handoff.
7. Report Gumroad as published only after the actual product content has been verified.

The existing build output and manifest remain the source of truth. Gumroad must not rebuild, rename, or silently omit platform artifacts.

## Gumroad API capabilities and limitations

Research checked against Gumroad's own help documentation on 2026-11-01.

### REST API v2

Gumroad's help center says the REST API cannot create products or upload product content. Product setup and downloadable content therefore cannot be fully automated through REST.

- Product creation must be done outside the REST API.
- Product content files cannot be uploaded through the REST API.
- Do not invent endpoints such as `POST /v2/products/:id/files`.
- Do not report a release as delivered merely because product metadata was updated.

Official reference: [Gumroad license keys and API FAQ](https://gumroad.com/help/article/76-license-keys).

### Gumroad MCP

Gumroad documents an MCP connector for supported assistants. It can create unpublished drafts, edit product details, attach files from public URLs, list attached files, and publish/unpublish products. Files are copied into Gumroad and are limited to 250 MB each.

The connector is currently available for Claude; Gumroad says ChatGPT is not available yet. It is not a general-purpose REST endpoint for Fe, and its public documentation does not establish a safe non-interactive rolling file replacement operation.

Fe must not pretend to be an MCP client by calling undocumented endpoints or scraping the dashboard.

Official references:
- [Connecting AI assistants to your Gumroad store](https://gumroad.com/help/article/355-connect-ai-assistants-to-gumroad)
- [Gumroad MCP](https://gumroad.com/mcp)

## Implementation stages

### Stage 1 — Read-only account support

- Add product listing only after verifying the official REST authentication and response contract.
- Validate configured product IDs without changing products or uploading files.
- Keep the manual rolling/stable handoff explicit.

### Stage 2 — File delivery

Do not implement an uploader until Gumroad documents a supported non-interactive way to add and replace product content files.

When that exists:
- Use a small client and the existing release artifacts.
- Replace the rolling product's current file set rather than accumulating obsolete downloads.
- Keep stable product content version-specific.
- Add dry-run, credential redaction, and partial-failure reporting.

### Stage 3 — Release integration

- Add Gumroad as an optional destination to `fe release`.
- Keep itch.io and Gumroad failures independently visible.
- Allow publishing existing artifacts without rebuilding.
- Never claim overall release success if a selected destination failed.

## Safety

- Never modify a product unless its ID is explicitly configured or selected.
- Never delete a product as part of a release.
- Never overwrite a stable product with rolling artifacts, or vice versa.
- Never publish public download URLs as a substitute for paid Gumroad content.
- Treat a timeout as an unknown result; reconcile product state before retrying.
- Show which files were delivered and which still need manual action.
- Unsupported operations must fail clearly, not fall back to browser automation or undocumented endpoints.

## Remaining decision

The rolling policy is settled: **one persistent product, current files replaced each release, historical builds retained outside Gumroad**.

The remaining implementation blocker is Gumroad's lack of a documented non-interactive file upload/replacement API. Stable/rolling artifact layout can be finalized against Fe's existing release asset naming when the preparation command is implemented.

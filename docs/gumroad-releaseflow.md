# Gumroad release flow

Fe uses the installed Gumroad CLI for stable release file uploads. Gumroad is an optional destination; it does not replace Fe's build, packaging, or artifact manifest.

## Product model

| Role | Product lifecycle | Release behavior |
|---|---|---|
| Stable | One pre-created product per stable version | Upload that version's artifacts to its matching product |
| Rolling | One persistent product | Not automated yet; keep the current manual workflow |

Product IDs are configured explicitly. Fe never guesses a product from its display name.

## Stable releases

Stable upload is supported by:

- `fe gumroad release --version 1.2.3` — upload an existing artifact manifest.
- `fe gumroad release --version 1.2.3 --dry-run` — inspect the local release plan without contacting Gumroad.
- `fe release --gumroad --userversion 1.2.3` — include the stable upload after Fe builds and creates the artifact manifest.
- `fe release --gumroad --skip-bmp --userversion 1.2.3` — publish to Gumroad without selecting the itch.io destination.

Each version must map to a product that already exists in Gumroad:

```yaml
gumroad:
  stable:
    "1.2.3": "PRODUCT_ID_FOR_1_2_3"
    "1.2.4": "PRODUCT_ID_FOR_1_2_4"
```

Create the product in Gumroad first and copy its product ID into the matching version entry. Use a separate product for every stable version. Fe does not create products or change product pricing, descriptions, visibility, or publication state.

Before upload, Fe reads `<butler.directory>/artifact-manifest.json`, requires its version to exactly match the requested version, and validates every artifact path. It then asks the Gumroad CLI for a non-mutating dry-run plan. If the product already contains files, Fe refuses to append another release. After upload, Fe runs the same plan again and verifies the expected file count.

This makes stable uploads intentionally one-shot per product. If an upload command returns an error, the result may be partial: inspect the Gumroad product before retrying. Do not clear or reuse a stable product to force a retry.

## Gumroad CLI setup

Install the Gumroad CLI separately from [antiwork/gumroad-cli](https://github.com/antiwork/gumroad-cli) and authenticate it using its own login flow:

```sh
gumroad auth login
gumroad products list
```

Fe invokes the executable as a subprocess and does not store or manage Gumroad credentials. It does not install the CLI automatically. The CLI's own authentication and product commands remain available for account management.

## Rolling releases

Rolling uploads are not automated yet. The intended policy remains:

1. Keep one persistent Gumroad product for the rolling channel.
2. Replace its current files for each rolling release; do not append another set of downloads.
3. Keep historical artifacts in Fe's normal release artifact store.
4. Verify the buyer-facing content before announcing the update.

The current CLI upload operation appends files while preserving existing files. Fe will not use that behavior for rolling releases. Rolling automation stays blocked until a safe replacement workflow is available.

## Safety and boundaries

- Gumroad is optional; projects that do not use it need no Gumroad configuration.
- Product IDs must be explicitly configured for each stable version.
- Stable uploads require a product with no existing files.
- Fe uploads only files listed in the existing artifact manifest.
- Fe never deletes products or silently overwrites a stable product.
- A failed or ambiguous upload must be inspected before retrying.
- Fe reports an upload as successful only after the CLI preflight confirms the expected file count.
- itch.io and Gumroad remain separately selectable destinations.

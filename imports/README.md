# Optional data imports

This directory documents the manifest format for new golingua data. Importing
is optional; the application does not fetch remote sources or execute imported
files. JSON manifests describe dataset identity, source attribution, and
relative data file paths.

`internal/importer.DecodeManifest` performs bounded, strict JSON decoding and
validates that every file names a declared source. Paths are lexical relative
paths and reject absolute paths and traversal. A future file reader must still
resolve the path under its configured data root before opening it. An optional
`checksum_sha256` detects byte changes when supplied; no historical JavaScript
digest is required.

`license` records the license label supplied for a source. The field is
metadata, not legal advice, license verification, or permission to redistribute
the source. Review the source terms and permitted use before adding external
records to the repository. Keep imported content data-only; never execute code
from an import.

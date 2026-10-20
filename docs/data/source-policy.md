# Data source and license policy

Every manually authored or imported record should retain a `provenance.SourceRef`
with a stable source ID, exactly one relative `path` or absolute `uri`, a source
version, and a license label. `author` and `checksum_sha256` are optional bounded
metadata. When a checksum is present it is lowercase hexadecimal SHA-256 of
the referenced bytes. A missing checksum is allowed and does not imply that a
historic source revision is required.

The reference validator checks shape only. It does not resolve a URI, read a
file, verify copyright ownership, interpret license terms, or authorize reuse.
Before including external data, review the source and license scope for the
intended use and record that decision separately. Retain attribution and
version information even when the source can be redistributed.

Importer manifests are bounded JSON data. They reject unknown or duplicate
fields, duplicate IDs and paths, unknown source references, malformed optional
checksums, and absolute or traversal paths. The manifest parser does not read
files. Any later reader must resolve a relative path beneath an explicitly
configured data root, defend against symlink escape, apply its own byte limit,
and never execute content from the data file.

The golingua data schema is independent of the retired JavaScript repository.
Do not require a historical SHA or copy expected output from it. Provenance
records where a datum came from; it does not establish linguistic correctness,
semantic approval, or translation quality.

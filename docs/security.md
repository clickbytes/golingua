# Input and dataset boundaries

The Go v0 CLI treats user text and optional language datasets as data. JSON
does not execute Go code. The importer does not fetch network sources or run
scripts from source packages. A provenance record identifies a source but
does not approve its contents or license terms.

`internal/text` rejects malformed UTF-8 and bounds request text before NFC
normalization and tokenization. Normalized spans retain byte offsets into
the original input. The CLI also bounds argument counts and output records;
typed `invalid`, `resource_limit`, `unknown`, and `unsupported` statuses must
remain distinct. A green build does not validate linguistic correctness.

Import manifests are strict versioned JSON with bounded fields, no duplicate
keys, and only relative declared data paths. The importer opens those paths
through Go's `os.Root`, confined to the selected data directory even when a
path contains symlinks. It validates the opened regular-file descriptor and
reads at most 32 MiB per file and 64 MiB per dataset. It checks declared size
and SHA-256 when supplied. Staging retains the exact validated bytes; pack
compilation rechecks the manifest, digest, CORE references and lexicon before
exposing immutable indexes. The active store pointer is not an approval of
semantic or linguistic mappings.

This is a bounded local-file workflow. It does not claim that arbitrary
untrusted directories being concurrently modified by another process are
safe to use as a dataset root. Use a controlled staging directory for imports.

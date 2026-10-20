# Reviewing and activating local datasets

The importer separates a candidate dataset from the active local snapshot.
`snapshot preview` validates the manifest and its files beneath the selected
data root, then reports the changes that would be applied. It does not modify
the active snapshot. `snapshot apply` performs the same validation and writes
the candidate as a new active snapshot only when explicitly requested.

```sh
go run ./cmd/golingua-import snapshot preview \
  --manifest ./data/manifest.json \
  --data-dir ./data \
  --store ./.golingua-store

go run ./cmd/golingua-import snapshot apply \
  --manifest ./data/manifest.json \
  --data-dir ./data \
  --store ./.golingua-store
```

The store is local to the chosen directory and is separate from source files.
Applying a candidate creates a complete immutable snapshot and changes the
active pointer only after the snapshot has been written and validated. If
validation or snapshot writing fails before the atomic pointer replacement,
the previous active snapshot remains selected. The report includes
`durability_confirmed`: if the pointer was replaced but syncing its directory
failed, the new snapshot is active for the running system, while a crash could
lose that pointer update. Older snapshots are retained for local audit and
review. The store is not a network service, shared database, or automatic
learning mechanism; applying a dataset never promotes content from user text.

Preview counts and file changes describe structural data only. The importer
does not judge whether a lexical mapping or language rule is correct. The owner
can inspect Go output manually after a separate runtime build is available.

Each imported record retains its source ID and the manifest's source metadata,
including version and license label. These fields preserve attribution context;
they do not establish copyright ownership or grant redistribution rights.
Review the source and applicable license before including external data in a
distributed dataset. An `unknown` license remains unknown and is not permission
to copy or publish the data.

Paths in the manifest are relative to `--data-dir`. The loader applies bounded
reads and rejects paths that escape that root, including symlink escapes. The
store path is an explicit local filesystem destination; it should be kept
outside the source dataset directory to avoid importing prior snapshots as
input data.

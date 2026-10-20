# Provenance and generation traces

`provenance.SourceRef` records a bounded source ID, one local path or URI,
version, license statement, and optional author/checksum. It is attribution
metadata; validation does not grant redistribution rights or activate data.

`provenance.Index` maps a source ID to deterministic record references. Each
reference carries a typed record kind and its stable record ID plus a defensive
copy of the full `SourceRef`. Multiple versions or sources for one record are
kept as separate entries. Exact duplicates, malformed IDs, invalid source
metadata, and more than 100,000 references are rejected. Lookups return copies.

Morphology cell rules may carry an explicit `procedure_id` and `provenance`.
Bundled English and Polish seeds declare both. Generation traces retain the
selected paradigm and cell, the applied procedure ID, and its source ID; an
exception step retains the exception rule and source ID. The shared
`status.Trace` has a separate 32-step bound and marks truncation. Detailed
morphology traces remain bounded by the finite-table contract.

The current seed APIs expose source indexes for declared procedure rules and
Polish exception rules. Lexeme, paradigm, and construction record kinds are
available to the index API when those records have explicit `SourceRef`
metadata; the current seed lexeme data does not invent such references.

Missing paradigm assignments and missing cell rules return an unavailable
generation state with a diagnostic code. They do not claim that a form is
ungrammatical or that a translation is unsupported. The current bootstrap has
no generation CLI or sentence engine, so traces are available through package
APIs only. CI and an API smoke do not evaluate translation quality.

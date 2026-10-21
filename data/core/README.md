# Neutral CORE data

The active editorial source for new CORE declarations is
`internal/catalog/concepts.go`, written as typed Go data. This directory does
not contain a second JSON copy or an activated runtime dataset. Existing
import and snapshot code can still read older JSON documents, but new catalog
entries should have one source of truth.

Add a `ConceptID` only for a distinct, language-neutral meaning. Supply a
definition, kind, optional parent and role signatures, and source metadata.
Do not store PL/EN words, flexion, case, word order or direct translation pairs
here. `LexicalSense → ConceptID` is reviewed in language-local data and is not
approved merely because its label resembles the concept ID.

The current count and coverage rules are described in
`docs/data/1000-goal.md`. The runtime still requires an activated snapshot
and language adapters before CLI translation can use these definitions.

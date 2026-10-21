# CORE catalog authored in Go

`internal/catalog` holds language-neutral concept definitions. The initial
batch contained 49 definitions. The first growth batch adds 187 distinct
definitions from the working concept areas in G07 issue #59. `catalog.Count()`
compiles the definitions and reports the actual count beside the editorial
target of approximately 1000. The target is not a build gate.

Each entry has a stable `ConceptID`, kind, short distinguishing definition,
optional parent and semantic role signatures. The source metadata identifies
the project-authored Go file; its license is recorded as `unknown` until the
project records an explicit license. No external dataset has been imported.
Do not infer a language-specific sense or approved binding from an ID or
English description. `seeing` and `looking`, for example, are separate events
because the latter specifies directed attention without requiring successful
perception. Grammatical negation, tense and word order belong to utterance
operators or language-local grammar.

The catalog compiles through `core.CompileConceptGraph` and returns an immutable
snapshot. It is not yet an active translation dataset: the current CLI runtime
loads a separately activated store snapshot, and PL/EN sentence adapters need
approved language data. Subsequent issues #60 and #61 add their own lexemes,
senses, paradigms and constructions. Their number of lexical records follows
each language's actual distinctions and need not equal the CORE count.

To add a concept, first write a distinct language-neutral definition in
`internal/catalog/concepts.go`, choose a canonical ID, kind and parent, and add
roles only when their meaning is clear. Keep provenance, check the compiled
graph for duplicate IDs, cycles and dangling parents, and report the new
actual count. Do not create synonyms as separate concepts just to approach the
target. Only explicit, approved `SenseID → ConceptID` mappings may authorize
translation; no PL↔EN lexical pair is stored in CORE.

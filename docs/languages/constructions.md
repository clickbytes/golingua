# Construction templates

The files under `languages/{pl,en}/constructions/` are language-local,
structural templates. They describe NP, VP, and PP slots and permitted slot
orders. They do not contain lexemes, sentence examples, semantic roles, or
sense-specific valency frames. A family is shared independently of any one
lexeme or sense.

`internal/construction.DecodeConstructionPackage` reads schema version 1 JSON,
rejects unknown and duplicate object fields, checks its declared language
against the family, and compiles the variants into a deterministic
`VariantIndex`. Compilation proves only that the structure is well formed.
These files are not loaded by the application runtime today.

No package currently has an approved `LexicalSense` to CORE mapping, a
sense-specific valency frame binding, or an analysis/realization provider.
Therefore none of these templates licenses translation or produces sentence
output. In particular, the Polish OVS order is a marked structural option; it
does not assert case marking, topic/focus, or interchangeable grammatical
roles. The English auxiliary-question and Polish `czy` templates represent
limited orders, not complete question grammars. A polarity operator and
meaning evidence must be supplied explicitly by a future planner.

When approved bindings and providers exist, they can refer to the same family
and variant data without copying a construction per lexeme. No external
lexical or linguistic source is claimed by the current structural templates;
their provenance records identify them as internal project data.

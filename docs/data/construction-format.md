# Construction and valency dataset format

`golingua-import constructions` loads a strict, bounded JSON dataset alongside
the exact language feature profile named by its `profile_version`. It validates
language local construction families and variants, sense specific valency
frames, and declarative operators. It compiles deterministic construction plan
signatures; it does not execute metadata or activate translation.

```sh
go run ./cmd/golingua-import constructions \
  --input ./data/pl-constructions.json \
  --features ./languages/pl/features.json
```

## Package shape

The dataset has `schema_version: 1`, one language/profile version, profile
provenance, `construction_packages`, `valency_frames`, and an optional
`operators` array. Each construction package contains one `ConstructionFamily`
and its `Variant` records, using the strict format decoded by
`construction.DecodeConstructionPackage`. Each frame names a typed `SenseID`
and explicitly lists the family IDs it permits. Frame IDs, sense IDs, family
IDs, variant IDs and selectional CORE references are kept as separate types.

The JSON below is a schematic PL transitive family. `sense:example.pl.transitive`
is a placeholder ID and must be replaced by an existing, reviewed lexical sense
before this data can participate in a runtime binding.

```json
{
  "schema_version": 1,
  "language": "pl",
  "profile_version": "v2026-10-20.1",
  "profile_provenance": {"id":"pl-profile","path":"languages/pl/features.json","version":"v2026-10-20.1","license":"internal"},
  "construction_packages": [{
    "schema_version": 1,
    "language": "pl",
    "family": {
      "id": "construction:pl.transitive",
      "language": "pl",
      "slot_specs": [
        {"id":"subject","position":0,"category":"NP","cardinality":{"minimum":1,"maximum":1},"optionality":"required","argument_binding":{"semantic_role":"agent"},"case":"nominative"},
        {"id":"verb","position":1,"category":"VP","cardinality":{"minimum":1,"maximum":1},"optionality":"required"},
        {"id":"object","position":2,"category":"NP","cardinality":{"minimum":1,"maximum":1},"optionality":"required","argument_binding":{"semantic_role":"patient"},"case":"accusative"}
      ],
      "provenance": {"id":"pl-construction","path":"data/pl-constructions.json","version":"1","license":"internal"}
    },
    "variants": [{
      "id":"construction:pl.transitive.svo",
      "family_id":"construction:pl.transitive",
      "sequence":[{"kind":"slot","slot_id":"subject"},{"kind":"slot","slot_id":"verb"},{"kind":"slot","slot_id":"object"}],
      "markedness":"neutral","priority":0,"direction_policy":"both",
      "provenance":{"id":"pl-transitive-variant","path":"data/pl-constructions.json","version":"1","license":"internal"}
    }]
  }],
  "valency_frames": [{
    "id":"pl_transitive_example",
    "sense_id":"sense:example.pl.transitive",
    "semantic_roles":["agent","patient"],
    "arguments":[
      {"id":"agent","role":"agent","required":true,"grammatical_function":"subject","case":"nominative"},
      {"id":"patient","role":"patient","required":true,"grammatical_function":"object","case":"accusative"}
    ],
    "allowed_family_ids":["construction:pl.transitive"],
    "provenance":{"id":"pl-valency-example","path":"data/pl-constructions.json","version":"1","license":"internal"}
  }],
  "operators": []
}
```

For an intransitive Polish example, declare an `agent` role and a required
subject slot, then omit an object slot and patient role. For an English
transitive example, declare its own family and frame, for example a subject
slot with `case: "subject"` and an object slot with `case: "object"`; these
are language local labels. The English profile and `profile_version` must match
`languages/en/features.json`. The semantic roles must be authored explicitly;
the importer does not derive them from word order or POS.

## Validation and activation boundary

Family and variant packages are decoded by the existing strict construction
loader. Valency frames keep grammatical function, case, and adposition policy
separate from semantic roles. Every allowed family must resolve in the same
dataset and language. Local feature names and values in family constraints and
variant conditions must exist in the supplied feature profile. Operators are
declarative closed types with typed family references and bounded provenance;
they contain no executable code.

`CompilePlanIndex` deduplicates identical structural plans and retains each
separate SenseID/frame permission. It does not make a bilingual lexical mapping
or a runtime `ConstructionBinding`. The CLI reports `compiled_not_activated`.
CORE and lexicon existence of the typed concept/sense references is unchecked
unless those indexes are provided to a higher level; no missing reference is
guessed to be valid. Unknown fields, duplicate IDs, dangling family/slot/role
references, mismatched language profiles, and resource limits fail closed.

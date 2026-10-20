# Phrase plans

`internal/construction` exposes separate `CompileNP`, `CompileVP`, and
`CompilePP` entry points and `CompilePhraseGraph` for bounded nested plans.
Phrase IDs identify reusable templates; child plans form a DAG and cycles,
missing references, cross-language nesting, and unsupported phrase roles are
rejected. NP inputs distinguish head, determiner, modifiers, and quantity.
VP requires an explicitly supplied sense-specific `ValencyFrame`; each
required frame role needs a typed NP/PP input. PP requires an NP complement,
and any governed case must be represented as a feature value licensed by the
language profile.

`languages/{pl,en}/constructions/phrases.json` declares only reusable
structural template shapes and the feature dimensions present in each language
profile. `DecodePhraseProfile` checks the schema, supported roles, IDs, and
feature references. Profiles contain no lexical entries, valency frames,
auxiliary choices, generated forms, or sentences. They are not loaded by the
application runtime.

Agreement constraints in an NP plan name explicit input IDs and a profile
feature. They are structural constraints for a later planner; this compiler
does not infer feature values or inflect a word. VP feature bundles can carry
explicit tense/aspect/mood/polarity/voice values when those dimensions exist
in a language profile. Omission stays omission: no default tense, polarity,
voice, argument, or auxiliary is inserted. The current repository still has
no approved sense-frame bindings or phrase realization provider. An auxiliary
placeholder is rejected until an approved provider or evidence licenses the
specific use. These
plans cannot produce surface text.

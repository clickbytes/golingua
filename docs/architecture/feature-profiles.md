# Language-local feature profiles and unification

Feature dimensions and their values belong to a language profile, not CORE. `internal/features` loads each versioned profile into an immutable lookup. The committed `languages/pl/features.json` and `languages/en/features.json` are separate data files with language-specific inventories; the Go package does not merge them into one global category union. The presence of profiles does not mean a parser, generator or language plugin is available in CLI.

## Values and requests

A `FeatureBundle` identifies its language and profile version. Each explicit dimension has one of three states:

- `known`: carries exactly one value declared by that profile;
- `unknown`: the dimension applies but its value is unresolved, optionally tied to a local scoped variable;
- `absent`: the input explicitly says the feature is not present or does not apply.

An omitted dimension means no information was supplied. It is distinct from explicit `absent` and from `unknown`. A profile can reject either explicit state for a dimension. `ValidateFeatureRequest` checks supported values and dimensions and enforces dimensions marked required by a profile. An unknown required value returns `incomplete`; an omitted or explicitly absent required value returns `invalid`. Unsupported feature IDs/values return stable `unsupported` diagnostics.

Profiles are immutable after `BuildProfile`/`DecodeProfile`; their accessors return copies. Bundles are validated against one exact language and profile version. They carry no CORE IDs, surface forms or rule code.

## Unification

`UnifyFeatureBundles` returns a fresh feature bundle and never mutates either input. Equal known values unify; differing known values return `unsupported` with `FEATURE_CONFLICT`. Unknown remains unknown when combined with a known or unknown value; it is never guessed to be a concrete match. Explicit absent conflicts with known/unknown, while a dimension omitted from one input can be filled from the other unless a required `FeatureConstraint` demands the feature in both inputs. Optional constraints permit omission; they do not turn an explicit absent or unknown state into a value.

Unknown variables have bounded IDs and an explicit scope. The unifier preserves the same variable reference but refuses to merge distinct scoped variables. This avoids accidental variable capture across local scopes. Failed unification returns no partial bundle.

## Profile coverage

The Polish profile includes its case inventory, grammatical gender distinctions, aspect and inflectional dimensions. The English profile keeps pronoun case/person/number and verb form/agreement dimensions of its own; it has no general noun case dimension. The English tense inventory is present/past; future realization belongs to later constructions and is not asserted as inflection here. These files define bounded labels, not complete grammars or approved word forms.

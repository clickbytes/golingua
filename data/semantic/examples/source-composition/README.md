# Source composition pilot

This is a small PL/EN composition example for the existing source-analysis path. It contains four perception sentences per language, one role swap per language, and one two-dog example per language. These are explicit inputs and observed outputs, not a test runner or gold corpus.

The selected language packages are [`pl/source.json`](../../languages/pl/source.json) and [`en/source.json`](../../languages/en/source.json); their exact source plans are beside them. [`metadata.json`](metadata.json) supplies the six neutral, agent-reviewed primitives used by this pilot. `pilot.vision_event` and `pilot.auditory_event` are distinct symbols with an explicit event port and experiencer/stimulus roles. `probe.teacher`, `probe.dog`, and `probe.definite` are unary Entity predicates. These are bounded formal proposals; owner language acceptance is **NOT EVALUATED**.

The package keeps the original lexical and frame records separate from the added source mappings in [`source-records.json`](source-records.json). In particular, the English and Polish teacher senses retain their original pending-audit provenance labels even though the current compiled lexicons authorize their use. The new `reference_predicate` mappings have separate #395 provenance; they do not alter or promote the compiled CORE lexicon.

## Explicit grounding

[`contexts/main.json`](contexts/main.json) is one shared input for the eight main cases and two role swaps. It declares E1 as a present Event, A and B as distinct singular Entity references, and leaves `role_references` empty. The mentions use the actual parser phrase-candidate IDs recorded in the source traces. Separate `resolution_evidence` rows manually link teacher mentions to A and dog mentions to B. This is request-local grounding evidence, not a class-to-identity rule or a world fact.

[`contexts/two-dogs.json`](contexts/two-dogs.json) is separate: it links two dog mentions to distinct Entity references B and C and explicitly supplies definite-reference conditions for both. The resulting PL and EN expressions preserve both IDs, and their final request-context payloads match. The conditions are caller-supplied context, not an automatic Polish definiteness rule.

The English plan maps the known `determiner=definite` feature to a `probe.definite` presupposition. Other determiner values are not guessed. The Polish plan has no determiner feature mapping; its references and conditions come from the supplied context. The plans allow only the observed present declarative and present negative variants. The PL negative frames are the existing `pl.widziec.visual.transitive_negative` and `pl.slyszec.auditory.transitive_negative` frames, with their source-authored genitive object arguments.

## WH questions

A separate four-input subset under [`wh-questions/`](wh-questions/README.md) exercises typed subject and object gaps for `see/widzieć`. It yields standalone `Ask` expressions with an answer variable at the exact selected frame role port. The caller supplies the other role reference and event; the answer is not filled from facts. Each input/package carries explicit #395 neutral provenance. This small subset does not claim general WH coverage or owner acceptance.

## Cases and observations

The 12 inputs are stored under [`cases/`](cases/), with literal CLI JSON observations under [`observations/`](observations/).

| Case | Input |
| --- | --- |
| PL visual positive | `Nauczyciel widzi psa.` |
| PL visual negative | `Nauczyciel nie widzi psa.` |
| PL auditory positive | `Nauczyciel słyszy psa.` |
| PL auditory negative | `Nauczyciel nie słyszy psa.` |
| EN visual positive | `The teacher sees the dog.` |
| EN visual negative | `The teacher does not see the dog.` |
| EN auditory positive | `The teacher hears the dog.` |
| EN auditory negative | `The teacher does not hear the dog.` |
| PL role swap | `Pies widzi nauczyciela.` |
| EN role swap | `The dog sees the teacher.` |
| PL two dogs | `Pies widzi psa.` |
| EN two dogs | `The dog sees the dog.` |

Ten stored observations (the eight main cases and the two role swaps) came from frozen binary SHA `31104c4637342b2ea519fbaf6f23dc426cdc8915cc4b764fb7bb408444371aa3`. After adding the explicit B/C context conditions, the two two-dog observations were re-run with binary SHA `3e6e8ca9a601749dce867640e856f862ba13db33fd03f614b6c2ce576ab080e4`. All 12 returned exit 0, `status=ok`, `composition.status=ok`, and one candidate. Each positive body is one `Atom`; each negative body is exactly one `Not(Atom)`. See/hear use different symbols. Main cases and role swaps resolve to E1/A/B, with role swaps E1/B/A. The two-dog cases resolve to E1/B/C. The selected source outputs retain `present` and `one` attributes and the supplied conditions. Independent `algebra check` on each extracted canonical document returned `ok`.

After the comparison CLI fix and the explicit two-dog context update, all six full-context PL/EN pairs (visual positive/negative, auditory positive/negative, role swap, and two dogs) compare as `proved_equivalent` under an empty rewrite profile and the same request context. The corrected comparison binary SHA is `3e6e8ca9a601749dce867640e856f862ba13db33fd03f614b6c2ce576ab080e4`. These comparisons check the selected expressions and context payload; they do not prove translation quality or world truth.

## Reproduce

Run from the repository root. The runtime config below uses a temporary scratch store; it does not read or modify a user store.

```sh
tmpdir=$(mktemp -d)
repo=$(pwd)
cat > "$tmpdir/runtime.json" <<EOF
{"schema_version":1,"dataset_store":"$tmpdir/store","dataset_id":"golingua-authored","language_package_dirs":[]}
EOF
for case_file in data/semantic/examples/source-composition/cases/*.json; do
  lang=$(jq -r '.language' "$case_file")
  text=$(jq -r '.text' "$case_file")
  context=$(jq -r '.context_file' "$case_file")
  ./bin/golingua algebra analyze \
    --language "$lang" --text "$text" \
    --runtime-config "$tmpdir/runtime.json" \
    --language-package "$repo/data/semantic/languages/$lang/source.json" \
    --source-plan "$repo/data/semantic/languages/$lang/source-plan.json" \
    --context "$repo/data/semantic/examples/source-composition/cases/$context" \
    --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json
done
```

For one independent check, save the `canonical_document` from an observation and check it without re-supplying metadata because that document is self-contained:

```sh
jq '.composition.candidates[0].canonical_document' \
  data/semantic/examples/source-composition/observations/en-see-positive.json > "$tmpdir/en.json"
./bin/golingua algebra check --input "$tmpdir/en.json" --json
```

To compare two complete canonical documents, extract them without changing their references or request contexts, then use an empty finite profile:

```sh
printf '%s\n' '{"id":"source-395-empty-profile","rules":[]}' > "$tmpdir/profile.json"
jq '.composition.candidates[0].canonical_document' \
  data/semantic/examples/source-composition/observations/pl-see-positive.json > "$tmpdir/pl.json"
jq '.composition.candidates[0].canonical_document' \
  data/semantic/examples/source-composition/observations/en-see-positive.json > "$tmpdir/en.json"
./bin/golingua algebra compare --left "$tmpdir/pl.json" --right "$tmpdir/en.json" \
  --profile "$tmpdir/profile.json" --context-id source-395-main \
  --metadata data/semantic/examples/source-composition/metadata.json --json
```

The analyses and comparison outputs are bounded observations of this authored profile. They do not establish general PL/EN coverage, an asserted fact, or owner acceptance.

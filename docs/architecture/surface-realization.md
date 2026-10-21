# Surface realization and source segments

This document records the bounded surface-output contract. The current Go
runtime is still an incremental pipeline: a licensed realization plan does not
by itself establish that every requested linguistic feature has a target
surface form. Execution currently supports only plans whose
`ConstructionPlan.Steps` is empty and whose construction remains in its input
family. Plans with composition steps or an output-family-changing operator are
explicitly refused: composed variant selection and surface realization of
operators are not carried through the execution pipeline yet. The segment API
only attaches exact byte spans and explicit source identities to strings
already emitted by a linearizer.

## Segment construction

`realizer.BuildSegments` accepts the final UTF-8 text and the ordered
`LinearizedPart` values emitted while producing it. Each part carries its exact
surface string, a coarse kind (`lexical`, `terminal`, `separator`, or
`morphology`), and at least one explicit trace identity such as source,
Meaning node, lexeme, sense, construction, or procedure ID. Spaces and other
bytes must be included in parts as emitted; the builder does not insert or
discard them.

The builder walks parts in order and assigns half-open `[start, end)` spans in
UTF-8 **bytes**. A span directly slices the final Go string. It verifies that
each part exactly matches the text at the current offset and that the last part
ends at `len(text)`. Therefore spans are ordered, non-overlapping, and cover
every output byte. Repeated spellings remain attributable to their explicit
part identities; the implementation never searches by spelling to guess a
source. It rejects missing trace IDs, empty parts, mismatched output, invalid
UTF-8, and unknown segment kinds.

The default bounds are 64 KiB of output and 4,096 parts. Package ceilings are
1 MiB and 16,384 parts. Exceeding a valid configured limit returns
`resource_limit`; malformed text, parts, identities, or budget configuration
returns `invalid`, with a typed diagnostic through `SegmentError.Result()`.
No partial segment list is returned on either outcome. A caller may omit the
budget to use the finite defaults.

## Current word and terminal spacing

The linearizer currently inserts one ASCII space between adjacent realized
slots and before a word-like terminal. Opening punctuation is separated from
preceding text and stays adjacent to the following part; closing punctuation is
attached to the preceding part with no inserted space. This is a small generic
spacing rule, not a language-specific punctuation or typography policy. A
word-like terminal may contain single internal whitespace between words;
repeated or boundary whitespace is rejected. The linearizer's own inter-part
spacing is still represented by explicit separator parts.

## Current execution safeguards

`Execute` does not infer morphological features from semantic roles. A slot
request must supply its feature bundle explicitly. When the selected family or
valency frame declares a case, the caller must also provide `CaseFeature`, the
language-local feature ID used to represent that case. A missing mapping or an
absent/unknown case value returns `incomplete`; a known value that differs
from the licensed case returns `unsupported`.

Every semantic role present in the selected Meaning must have a bound family
slot and a supplied lexicalization. A role without a family binding, or an
active role slot omitted from the selected variant, returns `unsupported`. A
missing request or generated form for a present role returns `incomplete`.
Optional role slots may be omitted only when their role is absent from Meaning;
an optional slot does not make a present role optional.

Hard family agreement constraints are checked only against the explicit
feature bundles supplied for both slots. Every constrained feature must be
known on both slots and have the same value; missing/unknown values return
`incomplete`, while unequal values return `unsupported`. The executor does
not infer a controller's values from Meaning or from a role label.

When a slot request targets a PP slot or a slot declaring an adposition,
execution refuses it with `unsupported`. The current surface API has no
separately licensed adposition lexeme generator or corresponding source-traced
adposition segment, so it does not emit or silently omit the adposition.

The selected variant's conditions are checked in both `Execute` and
`Linearize`; calling the linearizer directly does not bypass them. A
`feature_equals` condition requires the named feature to be known in the
explicit `SlotForm` feature bundle. An `information_structure` condition
requires explicit evidence for its semantic role. Missing slots or absent or
unknown evidence return `incomplete`; a known value that fails the condition
returns `unsupported`.

## Provenance, assumptions, and losses

Segment identities are copied from the linearizer's explicit source trace.
They do not prove semantic correctness and do not reconstruct Meaning. The
caller retains the Meaning graph and realization plan. Any plan assumptions
and losses remain attached to that plan and its outcome; generating valid spans
does not turn an `incomplete` plan into `ok`. An `unsupported`, `unknown`,
`ambiguous`, or `incomplete` result must remain visible to the caller with its
diagnostics. The surface layer must not silently switch to an unlicensed
construction, omit a feature, or fabricate a trace to return text.

For a lexical segment, `ProcedureID` identifies the final morphology trace step
when one exists. The complete ordered morphology trace remains on its
`SlotForm` in the `ExecuteResult`; the segment is a compact source reference,
not a copy of every trace step.

This segment API does not reverse-parse text, infer word boundaries, assign
lexical meaning, or recover a missing source trace. No reverse parse is claimed
unless a later implementation demonstrates and documents one separately.
Language-specific morphology, word order, agreement, and punctuation must come
from explicit target-language plans and procedures, not from this byte-span
builder. Translation quality still requires manual review of owner-provided
examples; valid coverage is only a structural guarantee. Reverse-parse
roundtrip validation and owner-supplied manual language review are NOT RUN for
this task.

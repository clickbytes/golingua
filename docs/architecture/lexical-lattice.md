# Lexical lattice

G05-02 connects source-spanned tokens to every licensed morphology reading,
lexical sense and MWE match. It does not parse constructions or select a
translation.

## Alternatives and evidence

Each token keeps its exact original UTF-8 byte span. Morphology readings remain
separate, including homographs and syncretic forms. A reading whose lexeme is
absent from the lexical snapshot is retained as `unknown_lexeme`; a present
lexeme without attached senses is `unknown_sense`. Every attached sense and
all of its semantic bindings remain metadata. An unmapped or pending sense is
still a lexical candidate; bindings do not authorize filtering at this stage.

No-reading morphology lookup means only `unknown_surface`. The analyzer does
not guess whether the token is an unknown lemma or an unsupported inflection.
Unknown tokens stay in the lattice alongside candidates for neighboring
tokens.

Flat `MWEFixed` and `MWEInflected` plans match consecutive word-token spans.
Every matching plan and overlap is retained with its source range and component
reading IDs. Component analyses remain in the lattice. `MWEInput` requires
phrase parsing and therefore marks the result `incomplete` with the plan ID.
`MWEFixed` matches the referenced lexeme without a feature constraint;
`MWEInflected` additionally requires every declared feature value to match.
Candidate identity comes from stable IDs and source ranges; any score is only
advisory and cannot merge, drop or choose alternatives.

An explicitly empty MWE-plan slice means the supplied inventory contains no
plans. A nil slice means no MWE inventory was supplied and therefore marks the
lattice `incomplete` with `MWE_INDEX_UNAVAILABLE`.

## Budgets and outcomes

Defaults are 16 KiB input, 256 tokens, 256 total retained reading, sense,
binding, MWE-edge and component-reading-reference units, and 256 compiled MWE
plans. Exceeding a limit returns the bounded partial
lattice with outcome `incomplete`, an affected token span or MWE plan path, and
a diagnostic. It never reports a truncated lattice as complete.

A scan without truncation can still have global `unknown` because one or more
tokens have no licensed morphology reading; neighboring candidates are kept.
Global `incomplete` means at least one requested analysis was left unresolved
because of an unsupported MWE component or an exhausted budget. It is distinct
from an infrastructure error. Invalid setup and incompatible observed
reading/lexeme language data are reported as typed errors. The current
morphology API exposes language identity through returned readings, so an empty
index cannot independently establish a language mismatch.

When the MWE plan count exceeds its budget, only the bounded prefix in the
supplied inventory order is examined; the result is explicitly `incomplete`.

## Current boundary

The lattice preserves lexical alternatives; it does not infer semantics from a
word alone, resolve ambiguity by insertion order, or authorize transfer. The
compiled CLI can translate one exact `seeing` route after loading an active
snapshot; other routes remain explicitly unsupported or unknown. Language
output has not received the owner's manual evaluation.

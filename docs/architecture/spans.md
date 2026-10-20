# Source spans for text analysis

This document defines how future language analysis preserves evidence from the
user's input. It describes the tokenizer work in G05-01; it does not claim that
the bootstrap CLI already analyzes or translates text.

## Offset unit and ownership

Every source span is a half-open interval `[start, end)` measured in UTF-8
**bytes** from the beginning of the original input. It can therefore be used
directly to slice the original Go string:

```go
surface := input[span.Start:span.End]
```

The interval must satisfy `0 <= start <= end <= len(input)`, and both edges
must be UTF-8 rune boundaries. A UI that counts Unicode code points, grapheme
clusters, or UTF-16 code units must explicitly convert from byte offsets; those
units are not interchangeable (especially for Polish combining marks and
emoji).

The original input remains authoritative evidence. Normalization and case
handling create matching views, never replacements for the input. A token's
surface is always recoverable from its source span, including original case,
punctuation, and decomposed Unicode sequences.

## NFC mapping

Matching may use Unicode NFC. Normalization can change the byte length: for
example, a letter followed by a combining mark may become one precomposed
character. Keep an explicit mapping from each normalized rune or normalization
cluster to its source byte interval. When a match covers normalized content,
derive its source evidence from the covering original intervals and validate
the resulting boundaries against the original input.

Do not calculate a source offset by adding an index in normalized text to an
index in the input. They belong to different strings and may use different byte
lengths. The shared text API does not impose a language-specific case-folding
policy. A tokenizer profile can opt into Unicode folding for matching while
the original surface and NFC form remain available for evidence.

## Token and sentence evidence

Tokens and sentence segments carry spans into the original input. Punctuation
is retained as token evidence and whitespace remains recoverable in the gaps.
They must not be silently absorbed into or discarded from the evidence needed
to reconstruct the original surface. `TokenizerProfile` supplies trusted
language-specific word-rune, punctuation, apostrophe, and case-folding choices;
it does not put language rules in the shared semantic core. NFC remains
available as `Normalized`; when a profile opts into Unicode case folding,
`Folded` is an additional matching form using locale-neutral Unicode folding;
profiles should opt in only when that behavior matches the language policy.
Adjacent token spans allow a later lexicon lookup to consider multiword entries
without changing token boundaries.
Tokenization does not assign lexical meaning or invent words.

`SegmentSentencesWithRules` accepts bounded additional terminal marks,
recognized punctuation and exact abbreviation spellings from a language
profile. The default inventory covers common Latin and CJK sentence terminators;
other punctuation stays in the returned sentence and is marked recoverable.
The shared layer has no abbreviation dictionary or regex rules by default.

Input length and token/rule expansion are bounded. Invalid UTF-8 is invalid
input; exceeding a configured byte or rule budget is `resource_limit`. Neither
condition is an infrastructure failure. Unrecognized punctuation must remain
visible and must not be turned into fabricated lexical content. The current
tokenizer returns a single deterministic token sequence; lexical ambiguity and
multiword matching are left to later analysis stages rather than guessed here.

## Current implementation boundary

G05-01 provides the source-preserving text primitives. The CLI still reports
translation as unsupported until later G05 tasks connect tokenization to
morphology, constructions, meaning, and realization. No language-quality claim
follows from having valid spans.

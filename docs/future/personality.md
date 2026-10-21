# FUTURE: presentation style and personality

Go v0 has no personality engine or per-user style memory. A future style
profile could control presentation choices such as formality, brevity,
addressing, and optional wording preferences. Those choices belong to the
language-specific realization layer. They must not change a MeaningGraph's
asserted facts, roles, polarity, question scope, or source evidence.

Style preferences should have an explicit scope and provenance, and a user
should be able to inspect, change, or remove them. A request without a
selected profile would use a documented default rather than infer a user's
personality from unrelated facts. Missing linguistic constructions remain
`unsupported` even if a preferred style asks for them.

This is a future design boundary, not a v0 capability or a promise of one
global voice across PL and EN.

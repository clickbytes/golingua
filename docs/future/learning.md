# FUTURE: proposals for new language knowledge

Go v0 does not learn from dialogue, change its CORE catalog at runtime, or
promote a user's correction into a language rule. This note sketches a
possible later workflow if the owner chooses to add one.

An observation could propose a new `ConceptDefinition`, lexical sense,
semantic binding, valency frame, construction, paradigm, or explicit exception.
The proposal must name its target layer and include source, license, language,
exact input context, and the claimed rule. A surface pair alone cannot justify
a cross-language mapping: each language's sense must be reviewed against the
neutral meaning independently. One observed form also cannot justify a whole
paradigm without a stated scope.

Proposals would be stored separately from the active immutable catalog. Only
an explicit review decision could produce a new versioned snapshot. Runtime
requests would pin one snapshot and never see a half-applied rule. Unmapped or
unreviewed observations remain `pending` or `unsupported`.

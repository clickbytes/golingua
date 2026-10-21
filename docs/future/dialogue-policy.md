# FUTURE: dialogue policy

A later dialogue policy could choose whether to answer, ask for
clarification, present alternatives, or decline an unsupported request. The
choice would consume typed outcomes from reasoning and language planning,
including `unknown`, `ambiguous`, `incomplete`, and `unsupported`. It should
never turn an unknown fact into a confident assertion for conversational
smoothness.

The policy would be separate from CORE, request-scoped facts, and PL/EN
grammar. A selected act would still need a licensed construction and surface
realization in the target language. If no realization is available, the
system should expose that gap rather than substitute a canned claim.

Go v0 has typed question execution and guarded translation but no dialogue
policy runtime. Whether personalization or multi-turn state is warranted
remains a later product decision.

# FUTURE: natural language generation goals

Naturalness is a later goal after the system can safely realize the intended
meaning. Future work might choose among licensed variants, handle discourse
references, information structure, morphology, and language-local style.
The system should retain the underlying MeaningGraph and source evidence so
that a fluent surface cannot hide a changed fact or swapped role.

PL and EN may need different realizations for the same neutral meaning.
Polysemy and grammatical alternatives should remain explicit until the
available context supports a choice. Unsupported tense, negation, case,
operator scope, or construction must be reported rather than guessed.

No mass set of invented human responses, automatic naturalness percentage,
or NLG module is included in Go v0. Future quality review should use actual
owner-provided inputs and clearly separate technical status from linguistic
judgment.

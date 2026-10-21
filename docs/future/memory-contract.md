# FUTURE: session and fact memory

This is a design note for a possible later product decision. Go v0 has no
persistent memory, no learning, and no memory-backed answer provider.

`ConceptDefinition` belongs to the versioned, language-neutral CORE catalog.
It defines a type of meaning and is not a remembered claim about an actual
person or event. A request-scoped `FactSet` contains asserted observations
for one question execution; it is not automatically saved. A future memory
would hold user-authorized records with an explicit scope:

- **Session state:** temporary references, preferences, and dialogue context
  for a named session. It expires or is discarded when the session ends.
- **Fact memory:** attributed claims about a world or user, with source,
  timestamp, polarity, confidence/status, and a target scope. Contradictions
  remain visible instead of silently overwriting older claims.
- **Episodic memory:** bounded records of an interaction or event with a
  pointer to the originating conversation and a retention policy. It does
  not become a CORE concept merely because it was mentioned repeatedly.

A future answer may consult a selected memory snapshot, but the resulting
MeaningGraph must still carry evidence and context. A missing fact yields
`unknown`; memory should not fill semantic gaps by inventing facts. Reading
memory and writing memory are separate operations with separate authority.
No current API, schema, database, or CLI flag is promised by this note.

# FUTURE: possible memory API boundary

A later design could define three separate operations:

1. `ReadSnapshot(scope, at)` returns a bounded, immutable view with record IDs
   and provenance. Reading does not authorize saving new claims.
2. `ProposeRecord(source, claim, scope, retention)` returns a reviewable
   proposal. It does not mutate active memory or CORE.
3. `ApplyDecision(proposalID, decision)` stores, corrects, or rejects a record
   only after the required approval and conflict checks.

All records would use stable IDs, versioned schemas, typed outcomes and
bounded queries. A request would pin one memory snapshot, just as the current
runtime pins its CORE dataset. Whether such an API is useful, where records
live, and who may approve writes are undecided. No code or dependency is
added to Go v0 by this note.

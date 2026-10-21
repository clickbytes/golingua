# FUTURE: memory data boundaries

Go v0 stores no conversation memory. If a later release adds it, users need a
clear way to see which records are saved, why they are saved, and which answer
used them. A memory record should retain its source and creation time and
distinguish user statements from system inferences or imported claims.

Proposed controls are explicit consent before persistence, per-session and
longer-lived scopes, bounded retention, correction, deletion, and export.
Deletion should invalidate indexes and derived summaries built from the
removed record; a corrected claim should keep an audit reference only where
the user has allowed that retention. Imported facts require separate source
and license review. Access to one user's memory must not arise from another
request merely because both mention the same CORE concept.

These are future design constraints, not a statement that v0 implements
consent flows, encryption, retention jobs, or a privacy database.

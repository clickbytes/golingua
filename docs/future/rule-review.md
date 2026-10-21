# FUTURE: rule review, conflict and rollback

Review should identify the exact IDs and version being changed, the source
and license, and which existing senses or forms a proposal might affect.
Conflicting claims should coexist as review items; the system must not choose
one just because it appeared more recently or frequently. A reviewer may
approve, narrow, reject, or request more evidence for each proposal.

Applying an approved change would compile a fresh CORE/language snapshot and
publish it atomically. The previous snapshot should remain identifiable for
diagnosis and rollback. A rollback selects the prior validated snapshot; it
does not rewrite old provenance or silently erase the review record. If a
rule was derived from private data later deleted, any retained derivative
needs a separate policy decision.

No review queue, learner, automatic promotion, or rollback command exists in
Go v0. These are constraints for a future design discussion.

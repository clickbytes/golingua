# FUTURE: evaluation questions for learning

Before building a learner, the owner would need to decide:

- Which kinds of observations may become proposals, and which data may be
  retained or shared for review?
- Who can approve a concept definition or a language-specific rule, and what
  evidence is needed for each scope?
- How should conflicts, polysemy and dialect variation be represented without
  forcing one global answer?
- How can a reviewer inspect the effect of one proposed rule on selected
  owner-supplied examples before activating a new snapshot?
- What is the rollback and deletion policy for bad or private source data?

Go v0 has no automatic linguistic score, held-out corpus gate, JS parity
requirement, or training dataset. Future evaluation must be designed after
the owner defines data rights and review authority; this note does not add a
runner or acceptance threshold.

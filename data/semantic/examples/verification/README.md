# Neutral goal-verification example

This small example uses uninterpreted relation symbols `r` and `p` with a
two-edge path and a unary property. It is a mathematical structure, not a
language example, stored fact, or claim about a world.

Files `metadata.json` and `goal.json` are the exact input records used by the
frozen CLI observation. `candidate.json` is the canonical algebra document
produced for that goal; `certificate.json` is the producer's untrusted
`constructed_unverified` claim. Its `goal_verified` field is false. The
separate `verification.json` records the independent checks. The
`candidate-wrong-orientation.json` mutation preserves types but swaps the
directed path roles; `verification-wrong-orientation.json` shows it being
refused.

Run the check from the repository root:

```sh
go run ./cmd/golingua algebra verify-goal \
  --goal data/semantic/examples/verification/goal.json \
  --candidate data/semantic/examples/verification/candidate.json \
  --certificate data/semantic/examples/verification/certificate.json \
  --metadata data/semantic/examples/verification/metadata.json --json
```

The successful result checks only the selected finite structural profile. It
does not evaluate facts, establish world truth, prove global consistency,
establish search completeness or uniqueness, or record owner language
acceptance. That status remains **NOT_EVALUATED**.

# Bounded proof checking

Issue [#390](https://github.com/clickbytes/golingua/issues/390) adds two
positive, bounded proof routes: comparison under an explicitly selected rewrite
profile, and independent checking of supplied positive evaluation witnesses.
The routes share a CLI entry point, but their claim types, inputs, and trust
checks remain distinct from goal and rewrite certificate checking.

## Profile-relative equivalence

`algebra compare` normalizes each supplied expression using only the rules in
the selected profile, then independently checks both rewrite certificates and
compares the resulting terms. A positive result is relative to the selected
metadata snapshot, `context_id`, and exact profile. Alpha-renaming, beta
reduction, or conjunction ordering count only when explicitly licensed. The
checker preserves reference and parameter IDs, argument roles, event IDs,
types, and scope. It does not infer that terms with different normal forms are
inequivalent: role, reference, event, and unlicensed-order differences remain
`undetermined`.

The certificate binds the original left and right document fingerprints,
snapshot, context, profile, and the two rewrite certificates. The verifier
replays each side against the explicitly selected inputs; a submitted
certificate is not trusted as its own proof. Certificate validity establishes
only the licensed profile-relative claim. It is not a model-theoretic
equivalence or a world-truth claim.

## Supplied evaluation witnesses

`algebra verify-proof --claim evaluation` checks positive answer witnesses
against the actual query, selected metadata, facts, selection, context, and
profile. The checker independently validates the cited fact IDs, named roles,
reference bindings, source evidence, time/scope, proof premises, existential
witnesses, and projected answer values. Positive and conflicting source
records remain separately visible.

This route does not run evaluation or search for answers. A valid proof
supports only its supplied answer tuple. `completeness_verified` is always
false; an empty witness set is `unknown`, not a negative answer. The
open-world profile has no countermodel or entailment checker. Those claim types
are refused as unsupported, and resource exhaustion never becomes a negative
proof. The pre-existing `SemanticEquivalence` and binary question APIs are
unchanged.

## Commands

Comparison and independent equivalence-certificate checking:

```sh
golingua algebra compare \
  --left data/semantic/examples/proofs/equivalence/alpha-left.json \
  --right data/semantic/examples/proofs/equivalence/alpha-right.json \
  --profile data/semantic/examples/proofs/equivalence/profile-alpha.json \
  --context-id proof-neutral \
  --metadata data/semantic/examples/proofs/equivalence/metadata.json --json

golingua algebra verify-proof --claim equivalence \
  --left data/semantic/examples/proofs/equivalence/alpha-left.json \
  --right data/semantic/examples/proofs/equivalence/alpha-right.json \
  --profile data/semantic/examples/proofs/equivalence/profile-alpha.json \
  --context-id proof-neutral \
  --metadata data/semantic/examples/proofs/equivalence/metadata.json \
  --certificate data/semantic/examples/proofs/equivalence/certificate-alpha.json --json
```

Independent checking of a supplied evaluation proof:

```sh
golingua algebra verify-proof --claim evaluation \
  --expression data/semantic/examples/proofs/evaluation/evaluation-expression.json \
  --facts data/semantic/examples/proofs/evaluation/facts-branch-conflict.json \
  --select data/semantic/examples/proofs/evaluation/select-evaluation-original.json \
  --context-id eval-neutral --profile positive-facts-v1 \
  --metadata data/semantic/examples/proofs/evaluation/metadata.json \
  --proof data/semantic/examples/proofs/evaluation/evaluation-proof.json --json
```

`compare` and all proof routes require `--json`. Metadata is selected explicitly
with one or more `--metadata` flags. Comparison takes two expression documents,
a profile JSON file, and a context ID. Equivalence verification also takes the
certificate emitted by comparison. Evaluation verification takes the original
expression, facts, selection, context ID, positive-facts profile, and a proof
document. Rewrite and goal claims keep their existing, separate routes. The
CLI refuses entailment and countermodel claims rather than interpreting them as
equivalence.

The [neutral examples](../../../data/semantic/examples/proofs/) and the
[#390 evidence receipt](../../../docs/work/dse/evidence/05-03/README.md) preserve
concrete inputs and observed results. Schema definitions for equivalence
certificates and evaluation results are additive `$defs` in the
[proof schema](../../../schemas/semantic/proof.schema.json); the existing
schema root union remains unchanged.

## Assessment boundary

These checks validate positive claims within finite, explicit profiles and
inputs. They do not supply a complete model checker, infer countermodels,
prove entailment, establish answer-set completeness, or authorize language
meanings. Technical observations, independent agent assessment, and owner
language acceptance are separate; owner acceptance is **NOT_EVALUATED**.

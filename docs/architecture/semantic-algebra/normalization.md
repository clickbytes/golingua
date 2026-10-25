# Bounded algebra normalization

`algebra normalize` applies only the rewrite rules explicitly licensed by a finite profile. It returns a checked expression, its canonical document, a step certificate, and a status. `algebra verify-rewrite` checks a submitted certificate independently against the original request and the explicitly selected metadata snapshots.

This is a bounded expression normalizer. It does not evaluate facts, establish truth, synthesize a proof from a goal, or prove unrestricted logical equivalence. `normalized: true` means that this implementation reached a normal form under the supplied profile and its fixed strategy. With an empty `rules` list, unchanged input is a normal form under that empty profile; no simplification is licensed.

## Command flow

Both commands require `--json` and one or more explicit `--metadata <file>` arguments. Repeat the flag for each selected package and its declared dependencies. Input comes from stdin by default or from exactly one `--input <file>`; `--input -` explicitly selects stdin. Paths are CLI arguments and are never read from request JSON.

```sh
go run ./cmd/golingua algebra normalize \
  --metadata data/semantic/examples/normalization/metadata.json \
  --input data/semantic/examples/normalization/ab.json --json \
  > /tmp/golingua-normalized.json

python3 -c 'import json; request=json.load(open("data/semantic/examples/normalization/ab.json")); result=json.load(open("/tmp/golingua-normalized.json")); json.dump({"schema_version":1,"request":request,"certificate":result["result"]["certificate"]},open("/tmp/golingua-verification.json","w"))'

go run ./cmd/golingua algebra verify-rewrite \
  --metadata data/semantic/examples/normalization/metadata.json \
  --input /tmp/golingua-verification.json --json
```

The verification envelope contains the original normalization request and the claimed certificate. The verifier rebuilds the environment from the selected metadata; it does not trust a snapshot ID or alias definition supplied by the certificate. Both commands report the selected package IDs, versions, and resource fingerprints separately from the certificate's formal snapshot ID.

The [proof schema](../../../schemas/semantic/proof.schema.json) defines the request and certificate wire shapes. The normalizer uses the same strict duplicate-key, UTF-8, surrogate, and JSON-depth checks as the typed expression decoder. Input requests, canonical documents, and individual terms are each limited to 1 MiB; certificates to 8 MiB; verification envelopes to 16 MiB. Runtime byte limits are not fully expressible in JSON Schema.

## Licensed rules and strategy

The closed rule vocabulary is:

- `alpha_v1`: rename bound variables using deterministic preorder names. Free parameter IDs and reference IDs remain unchanged.
- `beta_v1`: contract a known `Apply(Lambda, argument)` with capture-avoiding substitution. If contraction would erase a free parameter or reference dependency, the step is refused as unsupported.
- `alias_v1`: expand an approved, closed alias explicitly supplied by validated selected metadata. Arguments follow declared port order; alias binders are freshened against the complete root and checked after substitution. An expansion that erases a free argument dependency is refused.
- `and_flatten_v1`: flatten directly nested conjunctions.
- `and_order_v1`: sort conjunction operands by a deterministic structural key. This ordering is a canonicalization convention, not a claim about discourse order or scope.
- `and_deduplicate_v1`: remove only exact, identity-sensitive, binder-free duplicate operands after the licensed preceding transformations. It does not merge operands containing binders, alpha-equivalent `Exists` terms, distinct event instances, or distinct reference IDs.

Normalization visits beta and alias candidates bottom-up, restarting after each successful rewrite. It then applies licensed conjunction transformations bottom-up, again restarting after a successful change, and performs alpha naming as a final whole-root phase. Conjunction transformations inspect conjunction nodes only; they do not reorder across quantifier, discourse, or modal boundaries. The selected profile is the complete license: a rule not listed there is not applied.

The independent verifier checks the request's original input/profile, selected formal snapshot, each named rule and path, before/after roots, binder correspondence, side conditions, final output, and completion status. It does not call the normalizer to accept a producer's claim. A valid prefix certificate with `normalization_complete: false` is useful evidence of checked steps, but does not establish that the requested profile reached normal form. Neither a complete certificate nor an `ok` result establishes truth in a world or entails a fact.

## Bounds and statuses

Default budgets and maximums are 128 successful steps, 64 alias expansions, 32,768 retained trace nodes, and 262,144 deterministic logical work units. Work units count environment declarations and ports plus alias AST/interface nodes, initial root admission, whole-tree scans, and each applied step's full before/after roots. Candidate attempts also charge the environment inventory; verifier completion inspections charge their inspected roots. These are not CPU-time or energy measurements. The shared expression checker bounds a term to 4,096 nodes and depth 64, a type to depth 32, declarations to 512, and relation ports to 16. The JSON nesting limit is 256. Requests can lower budgets; zero selects the default, and negative or over-cap values are invalid.

Every candidate is checked against the full typed root before it is retained. If a step, expansion, trace, work, term-byte, certificate-byte, or canonical-document-byte limit is reached after valid steps, the result keeps the last valid prefix and certificate, reports `resource_limit`, and sets `normalized: false`. The current document and each serialized term/certificate retain their separate byte limits. An incomplete typed hole remains `incomplete`; the normalizer does not treat it as an existential witness or declare completion. Unsupported dependency-erasing reductions remain refusals rather than guessed simplifications.

Exit codes are `ok` 0, `invalid` 2, `unsupported` 3, `resource_limit` 4, `unknown` 5, `ambiguous` 6, `incomplete` 7; infrastructure errors use 1. An `ok` normalization and an independently valid certificate describe only this bounded procedure. They are not GoalSpec satisfaction, fact evaluation, or a theorem-prover result.

## Neutral examples and assessment

The files in [`data/semantic/examples/normalization/`](../../../data/semantic/examples/normalization/) are bounded mathematical controls with issue #382 provenance. They contain neutral relation symbols and no claims of linguistic meaning or owner approval. In particular, `ab.json` and `ba.json` exercise conjunction ordering; `no-license.json` leaves its expression unchanged; alias, beta, binder-collision, distinct-event/reference, and budget inputs exercise declared controls and refusals. The examples are inputs, not gold outputs or an automatic acceptance gate.

The typed IR and route boundaries are specified in the [formal contract](formal-contract.md) and [Meaning v1 compatibility contract](compatibility.md). The coordinator owns the scoped mathematical assessment, supported by independent Luna review. Owner language acceptance remains **NOT_EVALUATED**. Runtime observations belong in the DSE evidence record and must be reported separately from this design and from owner acceptance.

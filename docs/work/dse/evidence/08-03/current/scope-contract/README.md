# Scope contract boundary — #398, 2026-10-28

Status: **SPEC blocker for full #398 acceptance**, separate from agent language review and owner acceptance. No scope implementation, replacement IR, migration, new issue or test framework was introduced.

## Reproduced observations on code 56ddff9

- Both independent `Not(All(...))` and `All(Not(...))` documents still pass `algebra check` as typed propositions. Their existing expected binders/restrictors/operators were not edited.
- Production `algebra realize` refuses both independent documents in both languages with `UNSUPPORTED_EXPRESSION_PART`, retaining the entire scoped expression and no text.
- Four explicit scope paraphrase proposals are refused by production `algebra analyze` with `UNSUPPORTED_SOURCE_FOREST`. The existing source contexts are supplied only to make the loader request well formed; parsing stops before reference composition. These agent-proposed paraphrases are not owner-approved linguistic data.
- The separate invalid diagnostic `bound-event-time-field-probe.json` adds only `temporal_frame=present` to the existing bound Event `Exists e` term. `algebra check` rejects it with `UNKNOWN_FIELD` at `expression.body.body.body.operands[1].temporal_frame`. This diagnostic is not a new accepted expected document or a request to weaken validation.

Commands, raw outputs and exits are recorded here; they use the immutable binary identified by `../coordination-translation/environment.json`.

## Exact missing contract

The authoritative Algebra AST already distinguishes `Not`, `Quantifier`, `Exists`, `ID`, `Type`, `Restrictor` and `Body`: [terms.go](../../../../../../../internal/algebra/terms.go). The supplied scope documents bind Entity x/y and Event e. Bound terms have no temporal attribute slot; free [declarations](../../../../../../../internal/algebra/types.go) do have temporal/multiplicity metadata. The fixture's e therefore supplies no present-time value. Source present-tense evidence cannot be moved to a free reference or discarded while claiming preservation.

Meaning v1's [quantifier](../../../../../../../internal/meaning/operator.go) has kind/count over one target, and its closed [node schema](../../../../../../../internal/meaning/types.go) has neither binder identity nor restrictor. Converting Algebra variables to ordinary request-local referents would invent identity and lose the scope contract. Existing [execution](../../../../../../../internal/realizer/execute.go) and [linearization](../../../../../../../internal/realizer/linearize.go) APIs require a validated Meaning/construction plan; they do not accept the scoped Algebra binders directly.

A bounded contract must establish (1) how the original Algebra tree stays authoritative through target slots, preserving binders/restrictors and Not order; (2) how source time is represented for bound Event e, or how an explicitly untimed target assumption is compared without claiming identity with timed source meaning; and (3) how checked construction-role evidence reaches morphology without manufacturing Ref identity. This is a foundational representation/bridge decision, not an omitted dictionary entry or an owner-language-approval delay.

The active [#398 contract](https://github.com/clickbytes/golingua/issues/398) says: “Nie zmieniać algebry pod z góry wybrany tekst; błąd fundamentalny wraca do właściwego kontraktu/zadania.” Consequently the current adapter jobs did not silently extend the closed IR, change the independent scope expectations or flatten the operators. The owner plan's scope step remains unfulfilled until this contract is settled. No claim that scope realization is impossible in principle is made; it needs an explicitly defined contract absent from the current accepted route.

## Serial gate

#398 remains open and PARTIAL/BLOCKED; #422 stays draft and is not merged. #399, #400 and learning #329 are not started. Required scope source/target/round-trip flows remain unsupported/NOT RUN, while the independent typed ASTs remain valid. Owner linguistic acceptance: **NOT EVALUATED**.

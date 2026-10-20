# golingua agent execution contract

## Project boundary

Target repository `clickbytes/golingua` only. Read-only migration source: [`clickbytes/flexLingua@79c4a9586f7f94b6065da12dc25c4cb272ebdc92`](https://github.com/clickbytes/flexLingua/tree/79c4a9586f7f94b6065da12dc25c4cb272ebdc92). Go-first independent successor. The old JavaScript runtime is **not** a code-generation template and may have bugs; approved, independently reviewed golds are the semantic oracle.

## Exactly one delivery lane

- ONE implementing agent total; do not ask a second agent to work in parallel.
- At most ONE ACTIVE epic, ONE ACTIVE task, ONE working branch and ONE implementation PR at a time **across this repository**. All future epics/tasks remain QUEUED.
- Finish and review the current task and merge its PR before starting the next. Re-read `main`, active GitHub issues and open PRs each time. Do not auto-close an epic while any acceptance criterion is NOT RUN.
- Before editing: post the active task's owner, base main SHA, file-by-file plan, expected positive/negative behavior and test commands to its issue. Scope the PR to that issue. No parallel PRs or agent worktrees.
- A GitHub connector is not a local filesystem or terminal. Never claim tests, vet or manual CLI were executed by a connector call. Note PASS/FAIL/INFRA/NOT RUN precisely, with real CI run links, SHA, review and source provenance.

## Semantic integrity

- Concept != Lexeme != Sense != MorphLexeme != Valency != Construction != Procedure. Language-specific grammar never enters the language-neutral ConceptGraph.
- Source incomplete != unique; unknown != false; absent != unknown; ambiguous != unsupported; resource_limit != controlled refusal; invalid != infrastructure failure. Never guess a missing interpretation or fact to make CI pass.
- Morphology may operate on unmapped lexical items. Bilingual semantic transfer requires approved LexicalSense→CORE mapping and a licensed ConstructionBinding.
- Every interpreted claim, meaning, transfer, utterance and rule needs bounded structure, validity checks and provenance. Hard/typical relation distinctions matter; no fact may be asserted from a negation, question or hypothetical.
- Never change an approved oracle to fit observed output. All new sentence/fact oracles require independent review with named case IDs. Keep refusal and ambiguity separate from translation PASS.

## Build and release rules

- Standard Go tools (`go test ./...`, `go vet ./...`, `go build ./cmd/golingua`), fixed module path, Go 1.27, gofmt. CI `go-ci` is mandatory before merge. Prefer stdlib initially; new dependencies require an issue with justification/license.
- Target first: bounded PL↔EN with ~1,000 approved bilingual concept bindings; no claim of unrestricted translation, CEFR certification, learned intelligence or general-purpose LLM.
- No memory, auto-learning, UI, external network, FST, SQL or native `.so` language plugins in bootstrap. Future acceptance may revisit those explicitly and with evidence.
- Handoff lives in GitHub issue/PR with result classes and proofs, not `HANDOFF.MD`.

## Canonical design and implementation readiness

Before changing a Go package, read [`docs/architecture/README.md`](docs/architecture/README.md), the relevant layer contract(s), the current EPIC and the active task. **Every module has one owner**: ``internal/core`` concepts, ``internal/lexicon`` lexemes/senses, ``internal/features`` neutral feature mechanism, ``internal/morph`` forms, ``internal/valency`` sense frames, ``internal/construction`` language construction and phrase plans, ``internal/procedure`` typed operations, ``internal/meaning`` sentence semantics, ``internal/facts`` request facts, ``internal/reason`` evidence-based inference. No cyclic Go imports, and language packages never mutate CORE.

**No required `ROCK.VERB.MORPH.PAST` navigation API**: it was a discarded sketch, not a product requirement. Use stable typed SymbolIDs and snapshot-local dense runtime IDs. A generated form is not a new ConceptGraph node.

The design docs are **PROPOSED**, not evidence of shipped translation, memory, learning, reasoning or 1,000 bilingual concepts. If an issue, source contract or implementation contradicts the docs, reconcile in the active GitHub task and update the canonical design before proceeding. Never silently reinterpret a Go interface, let a typical relation prove truth, or promote a staged import as approved.

Each PR requires the **relevant cross-layer invariants and negative witnesses** (scope, referent identity, incomplete search, exceptional forms, provenance, capabilities). G00-03 and G00-04 will implement the authoritative status-result schema and parity-source manifest; current examples are conceptual. See [`docs/architecture/13-issue-traceability.md`](docs/architecture/13-issue-traceability.md) for per-EPIC contract links.

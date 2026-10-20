# golingua

**Go-first, bounded, concept-driven language engine** — independent successor prototype to [clickbytes/flexLingua](https://github.com/clickbytes/flexLingua).

> Status: BOOTSTRAP. No PL/EN translation, dialogue, morphology or graph-learning capability is claimed until backed by code, accepted examples and CI. This repository does **not** contain a port of the JavaScript runtime.

## Project goal

Build a small, auditable, procedural Graph Language Engine with:
- a language-neutral `ConceptGraph` and per-request typed `MeaningGraph`;
- language-local `LexicalGraph`, sense-specific `ValencyGraph`, and first-class `ConstructionGraph`;
- a bounded `ProcedureGraph` of typed generators, constraints and explicit exceptions;
- `EvidenceGraph` recording independent witnesses, provenance and approved quality checks;
- **PL ↔ EN** as initial languages; target **1,000 reviewed bilingual concept mappings**, not unrestricted CEFR or LLM-level language coverage;
- controlled reasoning and response planning from explicitly provided facts, *not* default memory, learning, or external world facts.

A concept is **not a word**. Lexical senses map to concepts; construction families belong to language plugins. Paths such as `EN.ROCK.VERB.MORPH.PAST` are navigation/API syntax, **not** one stored semantic graph node per inflectional form.

## Migration anchor

Source reference: [flexLingua main@79c4a9586f7f94b6065da12dc25c4cb272ebdc92](https://github.com/clickbytes/flexLingua/tree/79c4a9586f7f94b6065da12dc25c4cb272ebdc92), reviewed 2026-10-20. The JS source remains read-only. Source JSON/schema behavior is a migration input, not an implementation to copy line-by-line. Existing approved sentence golds are **independently reviewed oracles**, not outputs to retrofit.

## Delivery model

**ONE agent, ONE ACTIVE epic, ONE task/branch/implementation PR at a time.** All other epics are `QUEUED`. Finish each PR, review, merge, refresh `main`, then take the next task. Epics close only after their full Definition of Done (DoD); issues must carry explicit file scopes, invariants and evidence.

Planning and governance are tracked in GitHub Issues. First implementation PR supplies `go.mod`, a minimal CLI, typed identifiers and CI; no functionality is advertised before it is implemented and verified.

## Boundaries

Never silently invent a source reading, an approved bilingual sense, a sentence interpretation, a referent, a fact, or an exception. Absent, unknown, ambiguous, unsupported, invalid, incomplete and resource-limit outcomes must remain distinct. A reliable refusal is preferable to a false translation.

No server, UI, network dependency, SQL database, FST, native Go plugins, persistent memory or learning in the bootstrap. Those are optional future proposals gated by functional evidence and profiling.

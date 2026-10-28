# EN→PL→IR reverse round-trip observation — #398

This is a coordinator-observed production CLI flow on source commit `e4f93549ac03ba5c1bdf043ceb28381c95d39e4c`. The binary and runtime pins are recorded in [environment.json](environment.json); exact commands, exits and output hashes are in [commands.jsonl](commands.jsonl).

The source input `The teacher sees the dog, and the dog sees the teacher.` produces the Polish text `Nauczyciel widzi psa i pies widzi nauczyciela.` as one text-bearing `incomplete` candidate. CLI exit is 7. The target still has two `definite_reference` residuals, so this is not a complete translation.

The Polish text reanalyzes with CLI exit 0 as one `ok`, `complete` checked `And`: E1 has experiencer A/stimulus B, and E2 has experiencer B/stimulus A. The source context was independently supplied from the existing PL source coordination context. The resulting expression, with request context projected out, compares to the independent declared PL expression as `proved_equivalent` under the selected rewrite profile. The projected expression is preserved in [reanalyze-pl-expression.json](reanalyze-pl-expression.json); the independent comparator input is the existing `../source-coordination/pl-independent-expression-declared.json`.

The recorded comparison argv uses its original `/tmp/golingua398-reverse-evidence/reanalyze-pl-expression.json` path. For replay from this checkout, replace that `--left` value with the linked `reanalyze-pl-expression.json`; the copied bytes are identical.

This proves expression-level reanalysis for this selected preview. It does not prove full request-context equivalence, does not remove either definiteness residual, and does not establish owner linguistic acceptance (**NOT EVALUATED**). The complete EN→PL translation and full-context round trip remain unproved. Raw stdout/stderr are preserved losslessly in this directory.

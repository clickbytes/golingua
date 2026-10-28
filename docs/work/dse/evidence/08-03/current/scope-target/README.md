# Scope target evidence

Source commit: `465473b6937f3a40eecc847625daa48863bcd934`. The recorded binary SHA-256 is `112e374fd6d328a0346b2674cd1fb332fbb7228e78c0352001fc40d1e9537338` (see `environment.json`). CI run: [40655717647](https://github.com/clickbytes/golingua/actions/runs/40655717647).

## Observations

- Four timed independent IR → target realization runs returned `ok`: `not-all-en`, `not-all-pl`, `all-not-en`, and `all-not-pl`. Their raw stdout/stderr and invocation records are preserved here.
- `past-en`, `past-pl`, `swap-en`, and `swap-pl` returned `unsupported` (exit 3) with no text. The untimed EN and PL inputs returned `incomplete` (exit 7).
- The plan verification used compiled-plan comparison and fresh morphology regeneration by the same adapter. This is not independent linguistic proof.
- The present Event frame check returned `ok`. The invalid Entity temporal frame and invalid temporal value returned `invalid`.
- Present/past comparison returned `undetermined` with different normal forms. This does not prove inequality.
- Artifact replay returned `ok`.

The timed inputs were derived from existing independently authored IR by adding the indicated `present` or `past` temporal frame to its Event Exists binder; the existing expected results were left unchanged. The replay request contains an artifact exported earlier from code `c33403b`; replay was run with the binary and source commit recorded here. This was a bounded target probe, not a source-language evaluation.

Source analysis, translation, and round-trips: **NOT RUN**. Owner language acceptance: **NOT EVALUATED**.

## Reproduction

`commands.jsonl` records the original argv, which references `/tmp/golingua398-scope-commit` and `/tmp/golingua398-scope-receipt/`. The corresponding JSON inputs and raw output files are saved alongside this README. To reproduce, substitute the saved input paths in this directory and the recorded binary path if that binary is available. The command records also include the exact language package, runtime config, metadata, exit code, source SHA, binary hash, timestamps, and stdout/stderr hashes.

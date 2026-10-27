# DSE-08-01 realization evidence

## Provenance and scope

- Base/main SHA: `3bddd60266795ab1ea77cc772054b61e6bba2247`.
- CLI build: Go `go1.27.0 linux/amd64`; binary SHA-256 `d7cc6ed2b2953801c4e1d6ec2b54dc34525ee6ef2c947351403dba09e0415f6d`.
- Source SHA-256: `internal/engine/algebra_realization.go` `09285fd894487bbf196130977c67d7bd191fd4fd326c3d18d015a8961631b2f4`; `internal/cli/algebra.go` `92cb31b069e7babd3f4ec908f6cfb9e606d8222a01f9ba3857e34406c0e73e20`; `internal/cli/cli.go` `eb5136df696eb99f83163f9ef7ecbe47d9801bfc5a03ea65cf2638868fc514ba`.
- Selected metadata: `dse395.independent.probe` version `2026-10-26.1`, fingerprint `sha256:ad5be13b6bda00502a4c259e18b44e6b353c134362d86b760c7f971c89d8b67e`.
- Selected metadata file SHA-256: `data/semantic/examples/source-composition/metadata.json` `808576b35c4da6a798355239cbd8134fe3315b3635f634defce433c85c02beea`.
- Language package SHA-256: EN `data/semantic/languages/en/pilot.json` `3eabf59392c0235f22d21f4a9a856b9da52e8e6348a981477dba1f3d5571da0f`; PL `data/semantic/languages/pl/pilot.json` `84e4885f76eed641007e4d15106f36f0af6d9c220235d4cd544b2f584a4d08b0`.
- Runtime config: `../07-02/inputs/runtime.json`; SHA-256 `a94873c38de5cd4f04cdf2888513c163623397c54d7bcfbef0e88a2b2a2543b3`.
- Independent input SHA-256: `inputs/anonymous-vision.json` `8416937b3628e1cdd1d7a44d1394f44df5cf05e5e4c5ccf1f202c369c937cc4d`.
- Renamed input/package/metadata SHA-256: `anonymous-vision-renamed.json` `918f40d6b77e18561e2bd6139f79ee1daebe792df5d40731f4835208d4838386`; `en-renamed-language-package.json` `288b01a30f248976c1e35ddf5a29517ef72744953553927ce3355cdfec6ac41c`; `renamed-vision-metadata.json` `692a1c18648c5d76fb43770a3b69cb33412635bdb59099470455c8903a4e2301`.
- The positive input is typed algebra only. It has no source sentence or source spans. The selected language package and metadata are explicit CLI inputs.
- This delivery follows the owner START scope: one approved direct symbol target (`pilot.vision_event`), with an event identity and typed participant roles. General GoalSpec/compound-pattern matching is not claimed by this route.

## Build and smoke

| Command | Result |
| --- | --- |
| `gofmt -l internal/engine/algebra_realization.go internal/cli/algebra.go internal/cli/cli.go` | PASS, no output |
| `go vet ./...` | PASS, exit 0 |
| `go build -o /tmp/golingua-dse/bin/golingua-396-final ./cmd/golingua` | PASS, exit 0 |
| `git diff --check` | PASS, exit 0 |
| `/tmp/golingua-dse/bin/golingua-396-final --help` | PASS, exit 0; lists `algebra realize` |
| `/tmp/golingua-dse/bin/golingua-396-final --version` | PASS, exit 0; `0.1.0-alpha.1` |
| `go test` | NOT RUN; not required for this issue |

## Observed CLI results

Each `outputs/*.json` is the literal CLI stdout; the adjacent `.exit` file stores its exit code. Nonzero codes below are the CLI's status mapping for `ambiguous`, `incomplete`, or `unsupported`, with the full structured result retained.

| Input | Result | Observed output |
| --- | --- | --- |
| `anonymous-vision.json`, PL | `ok`, exit 0 | `Nauczyciel widzi psa.`; exact generated binding evidence includes Polish sense, frame, construction family and variant. |
| Same input, EN | `ambiguous`, exit 6 | Four explicit forms: `The teacher sees the dog.`, `The teacher sees a dog.`, `A teacher sees the dog.`, `A teacher sees a dog.`; each carries explicit determiner assumptions. No single reading was selected. |
| `anonymous-vision-plus-not.json`, PL | `incomplete`, exit 7 | The positive Atom has a surface; the `Not` expression remains `UNSUPPORTED_EXPRESSION_PART` in residuals. |
| `anonymous-vision-swapped.json`, PL | `ok`, exit 0 | `Pies widzi nauczyciela.`; coverage maps `scene` to experiencer and `viewer` to stimulus. |
| Renamed input + renamed metadata + renamed EN package | `ambiguous`, exit 6 | Same four EN alternatives; changing the symbol ID in all selected data did not require a Go selector change. |
| Near-miss symbol with a declaration but no matching language target | `unsupported`, exit 3 | No alternative; `NO_DIRECT_SYMBOL_TARGET` is retained as a residual. No nearest-symbol fallback occurred. |
| `anonymous-vision-extra-context.json`, PL | `incomplete`, exit 7 | The sentence is retained, while unused referent `extra` is explicitly reported as `UNCONSUMED_CONTEXT_REFERENT`. |
| `anonymous-vision-missing-concept.json`, PL | `incomplete`, exit 7 | No fabricated participant noun; `MISSING_REFERENT_CONCEPT` is reported once. |

The generated EN forms are intentionally `ambiguous`: the input Meaning contains no definiteness choice, and the target adapter exposes its alternatives and assumptions. This is a successful refusal to guess, not a claim that the source encoded one of those forms.

## Reproduction

Positive PL and EN commands (run from the repository root):

```sh
/tmp/golingua-dse/bin/golingua-396-final algebra realize --language pl --language-package data/semantic/languages/pl/pilot.json --runtime-config docs/work/dse/evidence/07-02/inputs/runtime.json --input docs/work/dse/evidence/08-01/inputs/anonymous-vision.json --metadata data/semantic/examples/source-composition/metadata.json --json
/tmp/golingua-dse/bin/golingua-396-final algebra realize --language en --language-package data/semantic/languages/en/pilot.json --runtime-config docs/work/dse/evidence/07-02/inputs/runtime.json --input docs/work/dse/evidence/08-01/inputs/anonymous-vision.json --metadata data/semantic/examples/source-composition/metadata.json --json
```

The corresponding stdout and exit codes are in `outputs/pl.json`, `outputs/pl.exit`, `outputs/en.json`, and `outputs/en.exit`. Other variants use the same command with the explicitly named input and package listed in their filenames.

## Review and acceptance status

- Coordinator observed both target pipelines; the PL plugin generated inflected accusative `psa`, while the EN plugin retained its article alternatives. The shared algebra bridge does not contain a language-specific word switch.
- Luna Low performed the bounded implementation and a read-only review. OpenCode v2.0.15 had no usable model/auth configured in this environment, so the required fallback was used. The coordinator separately reviewed and integrated several narrow CLI-discovered corrections.
- Agent assessment and coordinator verification are recorded separately from owner language acceptance. Owner language acceptance: `NOT EVALUATED`.
- `GoalSpec`, nested composition, multiple realized atoms, and richer context selection remain outside this direct-symbol route and are not presented as implemented.

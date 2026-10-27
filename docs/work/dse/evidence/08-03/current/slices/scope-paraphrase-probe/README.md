# Scope paraphrase syntax probe

Diagnostic only. These four EN/PL paraphrases remain proposals; this is not a language review, acceptance case, gold, or owner approval.

At source HEAD `165045d66b4b2d364aece956fac93309e895f20c`, the production source pipeline parsed each simple transitive control to one clause root. The four more explicit scope paraphrases all produced `forest_outcome=unknown` and zero roots. The syntax-only output identifies the first boundary as missing licensed word readings and a start construction. Public `algebra analyze` consequently returned `UNSUPPORTED_SOURCE_FOREST` for all four.

| Scope proposal | Surface | Unknown tokens | Roots |
| --- | --- | --- | ---: |
| EN `Not(All(...))` | `There is at least one teacher who sees no dogs.` | `There`, `least`, `one`, `no` | 0 |
| EN `All(Not(...))` | `For every teacher, there is no dog that teacher sees.` | `For`, `every`, `there`, `no` | 0 |
| PL `Not(All(...))` | `Istnieje co najmniej jeden nauczyciel, który nie widzi żadnego psa.` | `Istnieje`, `co`, `najmniej`, `jeden`, `który`, `żadnego` | 0 |
| PL `All(Not(...))` | `Dla każdego nauczyciela nie istnieje pies, którego ten nauczyciel widzi.` | `Dla`, `każdego`, `istnieje`, `którego` | 0 |

These observations localize refusal before semantic composition. They do not establish that the proposals have the intended natural-language readings. The independent algebra trees distinguish the requested scopes, but production source analysis and target realization do not yet support them. Adding isolated word readings would not provide the missing quantifier, relative-clause, and scope-composition contracts.

The syntax helper was copied to a detached worktree at the source HEAD, with only its input sentence list changed. It ran with Go `go1.27.0 linux/amd64` and `-mod=readonly`. Exact commands, statuses, raw outputs, and hashes are in `commands.jsonl` and `manifest.json`. The older public CLI binary has SHA-256 `46758c5a47a3c031697b0ebdc0ccc5be6cded475b8eb2d99b1a7a43a141f545e`; the Go source tree did not change between its build commit `943a8acd9cdcdc3b1cc6c77a09a6bb9dc8186140` and this probe HEAD.

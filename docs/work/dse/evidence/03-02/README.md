# DSE-03-02: explicit generic composition observations

Agent mathematical assessment: **ACCEPT within the finite profile below**. Owner linguistic acceptance: **NOT_EVALUATED**. These are literal CLI observations, not a corpus gate or an assertion of world facts.

`observations.jsonl` records exact argv, stdin where used, stdout, stderr and exit code. `environment.json` pins the base SHA, actual binary hash and source-file hashes of the uncommitted implementation used for the run. Inputs in `inputs/` are neutral structural examples authored for #381; they contain no asserted facts. The original absolute temporary paths are preserved in the log; `input-path-map.json` maps them to this directory for reproduction. Build the recorded source with `go build -o /tmp/golingua-dse/bin/golingua ./cmd/golingua`, then substitute the mapped input paths in the recorded commands.

## Observed behavior

- `two-hop`: interface `[a,b]`, one Entity witness, two r atoms in the path `(a,z),(z,b)`.
- `three-hop`: interface `[a,b]`, two distinct witnesses, three r atoms in path order. Nested And/Exists is retained; normalization is not claimed.
- `other-relation`: same construction over independent symbol s. `mapping-reversed`: explicit left-a/right-b sharing yields interface `[b,a]` and different wiring.
- `nary`: two ternary atoms preserve `[a,c1,b,c2]` with one explicitly shared witness; no unconnected argument is lost.
- `apply-free` and `bind-nonfirst`: the declared Param remains free; only the selected leading/mapped lambda is removed. Same textual spelling does not merge namespaces.
- `restrict`: r(a,b) and c(b), interface retained. `project`: one omitted port becomes Exists. `project-all`: both ports become Exists. `permute`: leading order `[b,a]`, original body argument bindings retained.
- `unify-success`: P maps to Ref(P), with separate ParameterID/ReferenceID namespaces. `unify-empty`: explicit `result_kind:substitution` and `{}`. `earlier-result-kind`: reports the requested earlier expression rather than the last trace step.
- Missing mapping returns incomplete/7; incompatible paired types and unknown ports return invalid/2; occurs-check cycle returns invalid/2; higher-order Lambda unification returns unsupported/3. No orientation is guessed.
- Forward references, duplicate steps, closed free parameters, null limits and irrelevant fields are refused. Unknown operation is unsupported. A per-term node bound returns resource_limit/4. A typed hole remains incomplete/7.
- Default stdin, the existing metadata-backed check, and checking generated two-hop/three-hop/nary canonical documents all execute successfully.

An auxiliary metadata manifest (`metadata-cli-complete`) was encountered by temporary directory enumeration and refused as a construction request with UNKNOWN_FIELD. It is retained with an explicit note; it is not one of the required recipe cases and is not evidence of a failed recipe requirement. The neutral manifest was completed with its required empty arrays before the coordinator run; no semantic requirement or expected result was changed.

## Verification and limits

Actual coordinator checks: gofmt list empty, `go vet ./...`, build, help and version succeeded. No new test file, test runner, fixture gate, automatic acceptance, language corpus or CI change was introduced. A separate GPT Luna Low reviewer inspected the output structures against expectations written before execution and accepted the scoped invariants. The coordinator owns this assessment.

Earlier isolated worker probes also checked aggregate substitution expansion, independent colliding inner binders under Restrict, cyclic/over-depth programmatic declarations and null-versus-omitted limits. Those were reported by the worker and reviewed in source; this directory's literal CLI log is the coordinator's independently observed evidence. Do not treat source review as a separate recorded CLI execution of those direct API probes.

This builder executes at most 64 explicitly ordered operations, checks actual intermediate/final expressions and preserves a bounded trace. It does not search for a GoalSpec solution, normalize equivalence classes, evaluate facts, prove transitivity or register a named derived relation. Unification supports only the documented typed first-order structural fragment; unsupported higher-order input is refused. Technical PR/main CI and the post-merge receipt are pending (NOT RUN at this pre-PR snapshot). The coordinator will record actual results on #381 after execution and merge; neither grants owner linguistic approval.

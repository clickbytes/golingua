# DSE-08-03 baseline observations

These are diagnostic probes against the unchanged Go source at the base SHA in environment.json. They identify first currently observed boundaries; they do not count as any of the six-route slice acceptances. Full #398 runs remain pending.

## Observed boundaries

- The four existing source WH inputs passed algebra analyze and produced one complete Ask each. Subject gaps map to experiencer; object gaps map to stimulus. These are existing #395 agent-authored proposals, and are not owner language approval.
- The configured production DSE route on the same WH inputs is incomplete with SOURCE_PERMISSION_MISMATCH; the active #397 translation plans permit only their direct transitive SVO routes.
- Independent algebra realize of canonical WH Ask documents is unsupported with UNSUPPORTED_EXPRESSION_PART: the target route realizes positive Atoms only.
- Passing those source-derived Ask documents to the generic evaluator with an explicit positive fact is unsupported at references.E1.temporal_frame, because the current positive-facts evaluator has no relative-time mapping. This correctly leaves the answer unresolved; the WH question is not converted into a fact.
- Separate generic evaluator controls work for the neutral positive Ask: facts-simple.json returns supported answers with proofs; facts-empty.json returns knowledge: unknown and no answers. A role-swapped neutral Ask returns different projected values. This confirms the evaluator control but does not substitute for evaluating the source-derived WH question.
- The proposed scope-language sentence returned UNSUPPORTED_SOURCE_FOREST; this is before DSE quantifier composition. The current result does not establish a linguistic interpretation.
- The ordinary definite EN to PL DSE control remains unsupported because definite-reference context conditions are not consumed by the target route.

Exact argv, exits, and raw stdout/stderr are in commands.jsonl and outputs/. The four projected source Ask documents are in inputs/; only inline sorts/signatures were removed because the separately selected metadata registry supplies those declarations. wh-evaluator-facts.probe.json is a synthetic explicit technical probe with agent-authored provenance. It is not a claim about real-world facts or owner-supplied data.

go build was run for the base binary; go test was not run. These observations are diagnostic only. Owner language acceptance: NOT EVALUATED.

# Bounded request reasoning

`internal/reason.Evaluate` operates only on one validated `facts.Set` and its pinned CORE snapshot. Rules are explicit Go data selected by the caller. The v0 schemas are positive symmetry and positive transitivity, and each can be activated only when its CORE `RelationType` declares that property. No rule text is executed. Typical ontology edges, concept definitions, lexical senses, questions, hypotheses and negation scopes are not premises.

The evaluator filters facts by time before deriving claims. A query without `at` sees only facts with unbounded validity; a UTC `at` includes facts whose closed validity interval contains that instant. Negative facts are direct evidence only. The evaluator never derives false from the absence of a positive fact. Its result is `true`, `false`, `conflict` or `unknown`; conflict retains both a positive proof and direct negative evidence.

Proof nodes name a direct fact or a declared rule with immediate premise proof IDs, all originating fact IDs, and their source evidence. Facts and rules are sorted by stable ID, so equivalent input orders select the same proof. Claim indexes by predicate and argument support bounded transitive joins. The evaluator stops with typed `resource_limit` when its work, derivation, proof-depth or proof-origin ceiling is reached; it never reports an incomplete closure as `unknown` or `false`.

This is a library stage. Query execution, content planning and response realization are later G06 issues. There is no persistent memory, general Datalog interpreter, automatic question corpus or language-quality metric. Manual owner question/answer review remains `NOT RUN` until there is a usable CLI path and owner inputs.

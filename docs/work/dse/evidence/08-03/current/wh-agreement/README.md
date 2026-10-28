# English WH agreement

The observations in this directory use source commit `b6742f7c8878549a6c6d744ea06aa1532751fd84`. Exact commands, binary and output hashes, exit codes and timestamps are in `commands.jsonl`. The runtime configuration uses the external snapshot store named in `../runtime-config-positive.json` and `../runtime-config-wh-object.json`.

The object-question input and source context are copies of the existing independent WH example with the subject's `multiplicity` explicitly changed from `one` to `many`. Realization returns `Whom do the teachers see?`; the existing singular example returns `Whom does the teacher see?`. Both remain `incomplete` because their retained context and target assumptions are not fully realized. The subject question returns `Who sees the dog?` as `incomplete`, explicitly recording the selected singular number and third person as assumptions.

The plural and singular questions both analyze as `ok` with the `stimulus` answer role. The mismatched `Whom does the teachers see?` is rejected as `unsupported`. A comparison of the analyzed plural expression with the independent plural expression returns `proved_equivalent` under the recorded profile after removing `request_context` from both documents. This checks the expression, not full context equivalence or a fact about the world.

These are agent observations, not owner approval. Owner linguistic acceptance: **NOT EVALUATED**. Remaining WH context residuals and target assumptions are not consumed by this evidence.

golingua algebra binding inspect --language-package data/semantic/languages/pl/pilot.json --json
golingua algebra binding validate --language-package data/semantic/languages/pl/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --json
golingua algebra binding instantiate --language-package data/semantic/languages/pl/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --input data/semantic/examples/lexical-bindings/pl-see-request.json --json

TASK_TMP=$(mktemp -d)
golingua algebra binding instantiate --language-package data/semantic/languages/pl/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --input data/semantic/examples/lexical-bindings/pl-see-request.json --json > "$TASK_TMP/instantiation.json"
jq '.result.symbol.document' "$TASK_TMP/instantiation.json" > "$TASK_TMP/document.json"
golingua algebra check --input "$TASK_TMP/document.json" --json

TASK_TMP=$(mktemp -d)
golingua algebra binding instantiate --language-package data/semantic/languages/en/pilot.json --metadata data/semantic/examples/lexical-bindings/metadata.json --input data/semantic/examples/lexical-bindings/en-link-indirect-request.json --json > "$TASK_TMP/instantiation.json"
jq '.result.goal' "$TASK_TMP/instantiation.json" > "$TASK_TMP/goal.json"
jq '.result.construction.canonical_document' "$TASK_TMP/instantiation.json" > "$TASK_TMP/candidate.json"
jq '.result.certificate' "$TASK_TMP/instantiation.json" > "$TASK_TMP/certificate.json"
golingua algebra check --input "$TASK_TMP/candidate.json" --json
golingua algebra verify-goal --metadata data/semantic/examples/lexical-bindings/metadata.json --goal "$TASK_TMP/goal.json" --candidate "$TASK_TMP/candidate.json" --certificate "$TASK_TMP/certificate.json" --json

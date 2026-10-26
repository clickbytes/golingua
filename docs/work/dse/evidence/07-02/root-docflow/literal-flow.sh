TASK_TMP=$(mktemp -d)
repo=$(pwd)
printf '{"schema_version":1,"dataset_store":"%s/store","dataset_id":"golingua-authored","language_package_dirs":[]}\n' "$TASK_TMP" > "$TASK_TMP/runtime.json"
golingua snapshot --apply --store "$TASK_TMP/store" --json

golingua algebra analyze --language en --text 'The teacher sees the dog.' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package "$repo/data/semantic/languages/en/source.json" \
  --source-plan "$repo/data/semantic/languages/en/source-plan.json" \
  --context "$repo/data/semantic/examples/source-composition/contexts/main.json" \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json \
  > "$TASK_TMP/en.json"
jq '.composition.candidates[0].canonical_document' "$TASK_TMP/en.json" > "$TASK_TMP/en-document.json"
golingua algebra check --input "$TASK_TMP/en-document.json" --json

golingua algebra analyze --language pl --text 'Nauczyciel widzi psa.' \
  --runtime-config "$TASK_TMP/runtime.json" \
  --language-package "$repo/data/semantic/languages/pl/source.json" \
  --source-plan "$repo/data/semantic/languages/pl/source-plan.json" \
  --context "$repo/data/semantic/examples/source-composition/contexts/main.json" \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json \
  > "$TASK_TMP/pl.json"
jq '.composition.candidates[0].canonical_document' "$TASK_TMP/pl.json" > "$TASK_TMP/pl-document.json"
printf '%s\n' '{"id":"source-profile-empty","rules":[]}' > "$TASK_TMP/profile.json"
golingua algebra compare --left "$TASK_TMP/en-document.json" --right "$TASK_TMP/pl-document.json" \
  --profile "$TASK_TMP/profile.json" --context-id source-395-main \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json \
  > "$TASK_TMP/comparison.json"
jq '.result.certificate' "$TASK_TMP/comparison.json" > "$TASK_TMP/equivalence-certificate.json"
golingua algebra verify-proof --claim equivalence \
  --left "$TASK_TMP/en-document.json" --right "$TASK_TMP/pl-document.json" \
  --profile "$TASK_TMP/profile.json" --context-id source-395-main \
  --certificate "$TASK_TMP/equivalence-certificate.json" \
  --metadata "$repo/data/semantic/examples/source-composition/metadata.json" --json

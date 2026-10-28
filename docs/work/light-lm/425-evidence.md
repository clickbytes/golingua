# #425 — obserwacje CLI

Baza `ddbfd1bbe1a8059ab8e0212e49297a9a8286375b`; lokalny diff #425; binarka `go build -o /tmp/golingua-learning-425-final ./cmd/golingua`, SHA-256 `18bb16858b2eb3cee14ceb6b3c786a2935343a72e0264cacc48596f196a69f29`. Każde wywołanie jest osobnym procesem. Konfiguracja `/tmp/golingua-learning-runtime.json` wskazuje snapshot `golingua-authored` w `/tmp/golingua-dse-400-store`, utworzony przez `/tmp/golingua-learning-425-final snapshot --apply --store /tmp/golingua-dse-400-store --json`. Plik konfiguracji zawiera `{ "schema_version": 1, "dataset_store": "/tmp/golingua-dse-400-store", "dataset_id": "golingua-authored" }`.

`TEACHERS` to wyłącznie techniczny znacznik wprowadzony przez agenta w celu obserwacji propagacji, nie ocena poprawności językowej ani akceptacja właściciela. Tłumaczenia pozostają `ambiguous`, a kandydatury `incomplete`. Ocena właściciela: **NOT EVALUATED**.

Żądanie korekty (`/tmp/golingua-425-verified-request.json`):

```json
{
  "id": "agent-teacher-marker",
  "language": "en",
  "lexeme_id": "lexeme:en.teacher.n",
  "features": [
    {
      "feature": "noun_number",
      "state": "known",
      "value": "plural"
    }
  ],
  "original_form": "teachers",
  "corrected_form": "TEACHERS",
  "initiated_by": "Codex by owner delegation",
  "recorded_at": "2026-10-28T11:40:59Z",
  "reviewed_by": "codex-agent",
  "reviewed_at": "2026-10-28T11:40:59Z",
  "provenance": {
    "id": "agent-technical-marker",
    "uri": "https://github.com/clickbytes/golingua/issues/425",
    "version": "2026-10-28.1",
    "license": "internal",
    "author": "Codex by owner delegation"
  }
}
```

Żądanie konfliktowe różni się ID `agent-teacher-conflict` i proponowaną formą `TEACHERSS`. Wyniki podają status, exit i hash pełnego stdout; surowe lokalne stdout są w `/tmp/golingua-425-verified-<name>.json`.

### base-form

```json
{
  "name": "base-form",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=plural",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [
    "teachers"
  ],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "3aa3788352924c241d1653dd2edcc92ccdbd11e2a0d2db54f17f177093360f6f",
  "message": null
}
```

### add

```json
{
  "name": "add",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "add",
    "--store",
    "/tmp/golingua-425-verified-store.json",
    "--input",
    "/tmp/golingua-425-verified-request.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "3f05bdbfff596dcf3e5eba23fb2a1792f42742a68b3641ae067fbe26ee237948",
  "message": null
}
```

### form-restart

```json
{
  "name": "form-restart",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=plural",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [
    "TEACHERS"
  ],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "aceba166b564180d937201c63c59c13fd4eae0ac32fdbe4b0c4ac74df4bf73c6",
  "message": null
}
```

### analysis-restart

```json
{
  "name": "analysis-restart",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--surface",
    "TEACHERS",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "056355c8bf9167cea7f960114470f90891361b2df7df3d7f0d29409e49574e55",
  "message": null
}
```

### other-cell

```json
{
  "name": "other-cell",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=singular",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [
    "teacher"
  ],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "2f22d110219721ec3c82d02d9f5f131de5c60f60447f67aad96c3b3c6c57d145",
  "message": null
}
```

### other-lexeme

```json
{
  "name": "other-lexeme",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.dog.n",
    "--feature",
    "noun_number=plural",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [
    "dogs"
  ],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "7b4cd246defcca1c26e811b138f273fa126e394f2a689d7cfb2d713d04ac7f75",
  "message": null
}
```

### new-sentence

```json
{
  "name": "new-sentence",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "translate",
    "--from",
    "pl",
    "--to",
    "en",
    "--text",
    "Nauczyciele widzą psa.",
    "--runtime-config",
    "/tmp/golingua-learning-runtime.json",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 6,
  "status": "ambiguous",
  "forms": [],
  "texts": [
    "The TEACHERS see the dog.",
    "TEACHERS see a dog.",
    "TEACHERS see the dog.",
    "The TEACHERS see a dog."
  ],
  "diagnostics": [
    "TARGET_FRAME_VARIANT_MISMATCH",
    "TARGET_ALTERNATIVES_AMBIGUOUS"
  ],
  "output_sha256": "b2d37e4f98037e30b3bbab43cbde4a05a985068d7706c98d7597f19ce864b49e",
  "message": null
}
```

### conflict

```json
{
  "name": "conflict",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "add",
    "--store",
    "/tmp/golingua-425-verified-store.json",
    "--input",
    "/tmp/golingua-425-verified-conflict.json",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "forms": [],
  "texts": [],
  "diagnostics": [
    "LEARNING_RULES_CONFLICT"
  ],
  "output_sha256": "b46731781a547b62873383e0d2facf3360523397634cf2a4616c03fd9a0d2b8f",
  "message": "compile English exceptions: overlapping exception rules require an explicit override from the more-specific rule"
}
```

### revoke

```json
{
  "name": "revoke",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "revoke",
    "--store",
    "/tmp/golingua-425-verified-store.json",
    "--id",
    "agent-teacher-marker",
    "--actor",
    "codex-agent",
    "--at",
    "2026-10-28T11:40:59Z",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "fadb08a223bd49b5d9a012bc8ca44601e64ea21c80ec3698505a1017718fc937",
  "message": "correction revoked"
}
```

### form-after-restart

```json
{
  "name": "form-after-restart",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=plural",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [
    "teachers"
  ],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "3aa3788352924c241d1653dd2edcc92ccdbd11e2a0d2db54f17f177093360f6f",
  "message": null
}
```

### sentence-after-restart

```json
{
  "name": "sentence-after-restart",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "translate",
    "--from",
    "pl",
    "--to",
    "en",
    "--text",
    "Nauczyciele widzą psa.",
    "--runtime-config",
    "/tmp/golingua-learning-runtime.json",
    "--learning-store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 6,
  "status": "ambiguous",
  "forms": [],
  "texts": [
    "The teachers see the dog.",
    "Teachers see a dog.",
    "Teachers see the dog.",
    "The teachers see a dog."
  ],
  "diagnostics": [
    "TARGET_FRAME_VARIANT_MISMATCH",
    "TARGET_ALTERNATIVES_AMBIGUOUS"
  ],
  "output_sha256": "ac4201d36da2cf491350f17dd46ea6f5fc4f96a812309f1274317e3085e0d117",
  "message": null
}
```

### revoke-again

```json
{
  "name": "revoke-again",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "revoke",
    "--store",
    "/tmp/golingua-425-verified-store.json",
    "--id",
    "agent-teacher-marker",
    "--actor",
    "codex-agent",
    "--at",
    "2026-10-28T11:40:59Z",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "forms": [],
  "texts": [],
  "diagnostics": [
    "LEARNING_REVOKE_FAILED"
  ],
  "output_sha256": "f2a10488c7896a0cf11f16643a7e59410a72b8401c57be91f8935142cefd7eaa",
  "message": "learning record \"agent-teacher-marker\" is already revoked"
}
```

### inspect

```json
{
  "name": "inspect",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "inspect",
    "--store",
    "/tmp/golingua-425-verified-store.json",
    "--json"
  ],
  "exit": 0,
  "status": "ok",
  "forms": [],
  "texts": [],
  "diagnostics": [],
  "output_sha256": "ace10ba5e94bc8b83450e07c50ab16f2c5dbf1271a7f00309857a8a1f4fca307",
  "message": null
}
```

### missing-store

```json
{
  "name": "missing-store",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=plural",
    "--learning-store",
    "/tmp/golingua-425-verified-missing.json",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "forms": [],
  "texts": [],
  "diagnostics": [
    "LEARNING_RUNTIME_SETUP_FAILED"
  ],
  "output_sha256": "51b129bb15891b1d33aa21e85fe7e900378925f92953e42a43fb6361bd85757b",
  "message": "validate active morphology corrections: open learning store: open /tmp/golingua-425-verified-missing.json: no such file or directory"
}
```

### wrong-baseline

```json
{
  "name": "wrong-baseline",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "add",
    "--store",
    "/tmp/golingua-425-verified-negative-store.json",
    "--input",
    "/tmp/golingua-425-verified-wrong-baseline-input.json",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "diagnostics": [
    "BASELINE_SURFACE_MISMATCH"
  ],
  "output_sha256": "4583b04c2087a008ce04ab295f5cb8b4eff302d2bed3252bd5d473534b360220"
}
```

### unknown-cell

```json
{
  "name": "unknown-cell",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "add",
    "--store",
    "/tmp/golingua-425-verified-negative-store.json",
    "--input",
    "/tmp/golingua-425-verified-unknown-cell-input.json",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "diagnostics": [
    "BASELINE_GENERATION_FAILED"
  ],
  "output_sha256": "4044c5245e3f315ca8ce127cfab4f814361ba8a983b4506432145b4907ee8930"
}
```

### corrupt-store

```json
{
  "name": "corrupt-store",
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=plural",
    "--learning-store",
    "/tmp/golingua-425-verified-corrupt-store.json",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "diagnostics": [
    "LEARNING_RUNTIME_SETUP_FAILED"
  ],
  "output_sha256": "0c00faec74d87b2bbdb6219787eba4d4e9e545df00a8086945bf5a1f319368e0"
}
```

### unknown-lexeme

```json
{
  "argv": [
    "/tmp/golingua-learning-425-final",
    "learning",
    "add",
    "--store",
    "/tmp/golingua-425-verified-negative-store.json",
    "--input",
    "/tmp/golingua-425-verified-unknown-lexeme-input.json",
    "--json"
  ],
  "exit": 7,
  "status": "incomplete",
  "diagnostics": [
    "BASELINE_NOT_UNIQUE"
  ],
  "output_sha256": "e9da633429c2011ba45e704473bc87bf18ea9e8af4b0a179bc7998a2d790e08b"
}
```

### stale-snapshot

Kopia aktywnego wpisu ma digest zastąpiony 64 zerami; wynik to odmowa całego overlayu.

```json
{
  "argv": [
    "/tmp/golingua-learning-425-final",
    "forms",
    "--language",
    "en",
    "--lexeme",
    "lexeme:en.teacher.n",
    "--feature",
    "noun_number=plural",
    "--learning-store",
    "/tmp/golingua-425-verified-stale.json",
    "--json"
  ],
  "exit": 2,
  "status": "invalid",
  "diagnostics": [
    {
      "code": "LEARNING_RUNTIME_SETUP_FAILED",
      "message": "validate active morphology corrections: records[0] (agent-teacher-marker): baseline digest does not match requested snapshot",
      "path": "learning.forms",
      "severity": "error"
    }
  ],
  "output_sha256": "d54e44c5778371b345a15fdbbeb911356baf4216e449c3ff2d2609e0683e2190"
}
```

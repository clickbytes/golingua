# Target coordination diagnostic probe

This source-independent input contains exactly two approved `pilot.vision_event` atoms under one root `And`, with explicit event identities and shared entity reference B. It deliberately contains no `pilot.r` alias, lambda, quantifier, source text, or source evidence. It isolates target composition from the earlier source-parser and relation-alias boundaries; it is diagnostic evidence, not an owner-approved language example or acceptance gold.

At source SHA `691a8fa1b6790c1c45aa89c197bb9f601d5ba8b2`, runtime capabilities load for both EN and PL. `algebra realize` returns `ambiguous` / `MULTIPLE_REALIZABLE_ATOMS` for both languages (exit 6), retaining both atoms as residuals and producing no partial surface alternative. The observation confirms that the current target adapter refuses multiple exact atoms even when each atom uses an existing exact target binding. It does not show that source parsing, clause coordination, or the `pilot.r` path works.

`environment.json` records source and binary identity, config and package fingerprints. `commands.jsonl` records exact arguments, exits, output paths, hashes, and observed status.

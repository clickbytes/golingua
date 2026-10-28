# Anonymous relation: explicit alias and incomplete previews

Observed on source commit `e427779a607153eaa7302d1b3382bcb5ad96ff9a` using the binary and exact command lines recorded in `commands.jsonl`. Each command's exit code and stdout/stderr hashes are recorded there. The configured runtime uses the local snapshot store named in `../runtime-config-positive.json`; that store must contain an active `golingua-authored` snapshot when reproducing the realization commands.

The independent `two-hop.json` expression contains two applications of `pilot.r`. `request.json` explicitly selects the agent-reviewed alias proposal and the `alias_v1` rewrite rule. Normalization returned `ok`, and independent certificate verification returned `certificate_valid: true` and `normalization_complete: true`. This establishes equivalence only under that selected formal definition and rewrite profile; it does not establish a general natural-language meaning or any fact.

The normalized expression retains two independently bound vision events. The English and Polish target packages are explicitly selected; neither treats the alias as an operational primitive. Both CLI realizations return exit 7 / `incomplete` and one descriptive surface. Each retains two `EVENT_TEMPORAL_FRAME_UNSPECIFIED` residuals: the original function does not specify event time, and present-tense morphology does not supply that missing meaning. The output is a preview, not a complete realization or a claim about the world.

Linguistic assessment: agent proposal, limited to these forms. Owner linguistic acceptance: **NOT EVALUATED**. This evidence alone does not complete #398: context-preserving round trips and remaining relation and WH acceptance criteria still require review.

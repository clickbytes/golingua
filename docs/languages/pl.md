# Polish language data — Go lexical graph, valency and morphology

## Current scope

The executable `languages/pl` package builds **982 typed Polish lexemes, 1007 distinct lexical senses, 1006 binding records covering all 1000 CORE concepts**: 876 approved equivalent bindings and 130 pending records across 124 pending concepts. No CORE concept is absent. `pl.LexicalDocument()` assembles `lexicon_*.go` domains in deterministic order and `pl.CompileLexicon(coreGraph)` validates the graph. The editorial worklist `data/pl/1000-core-review.tsv` records 876 approved rows and 124 pending rows. `entity` is represented by the register-limited broad candidate `byt`; its binding remains pending. The 876 lexical decisions are recorded in delegated review batches 1–20 for issues #152, #154, #156, #158, #160, #162, #164, #166, #168, #170, #172, #174, #176, #178, #180, #182, #184, #186, #188, and #190; they do not approve morphology, valency, constructions or sentence translation.

Polish lemmas are distinct from CORE concepts. Examples of polysemy include `szkoła` facility/institution, `bank` institution/office, `ryba` animal/food, `skóra` organ/material. Many verbs have sense-specific cases and prepositions in `ValencyFrame`. They do **not** have automatic morphology or authorized construction paths.

## Valency — 273 frames

`pl.CompileValency()` builds 273 frames with roles copied from actual CORE definitions and separately authored Polish case/adposition constraints. The active source files include `valency.go`, `valency_motion.go`, `valency_actions_physical.go`, `valency_actions_communication.go`, `valency_actions_communication_more.go`, `valency_actions_communication_last.go`, `valency_actions_outcomes.go`, `valency_actions_health_routine.go`, `valency_actions_phrase_level.go`, `valency_actions_transfer_cognition.go`, `valency_actions_remaining_2.go`, `valency_actions_exchange.go`, `valency_actions_health_remaining.go`, `valency_actions_abstract.go`, `valency_actions_remaining.go`, `valency_actions_activity.go`, `valency_actions_production.go`, `valency_actions_life.go`, `valency_actions_social.go`, `valency_actions_cognition.go` and `valency_actions_multirole.go`. Frame families include `pl.intransitive`, `pl.transitive`, `pl.oblique`, `pl.ditransitive`, `pl.transitive_oblique` and `pl.double_np`. These structural family references are **not approved `ConstructionBinding` records** and do not activate translation.

Batch 8 adds pending senses and transitive frames for opening/closing objects, breaking, heating/cooling, ignition/extinguishing and battery charging. The charging sense uses the distinct lexeme ID `pl.ladowac_baterie.v`; `pl.ladowac.v` remains `lądować`. The current batch corrects several questionable editorial choices and retains each decision in the TSV. Entries with phrasal or context-sensitive realizations may require different roles or optionality after owner review. This is an authoring prototype, not exhaustive Polish valency.

Batch 9 adds eight pending communication senses and their selected frames. `prosić` and `informować` receive explicit `o` complements, `zapraszać` uses `na` + accusative, and recipient roles are dative where selected. These authored frames remain structural candidates, not approved constructions.

Batch 10 adds eight pending activity/event candidates: `porównywać`, `ranić`, `sadzić`, `ostrzegać`, `polecać`, `zaczynać`, `kończyć` and `kontynuować`. Frames record selected accusative, dative and `przed` + instrumental arguments. These authored frames remain structural candidates; infinitival complements are not encoded.

Batch 11 adds six pending candidates: `kraść`, `kategoryzować`, `rysować`, `nagrywać`, `współpracować` and `ubierać się`. Selected frames describe their accusative, dative and prepositional complements; these do not approve constructions or supply verb morphology.

Batch 12 adds seven pending senses for `zgadzać się`, `odmawiać`, `dyskutować`, `uznawać`, `dziedziczyć`, `wzrastać` and the quantity-change sense of `spadać`. The existing downward-motion sense of `spadać` remains distinct; no verb morphology is added.

Batch 13 adds four pending senses: `skarżyć się`, `żartować`, `wolać` and `próbować`. Frames record the selected dative, accusative, genitive and `z` patterns; infinitival complements remain unencoded.

Batch 14 adds four pending communication senses: `stwierdzać`, `nie zgadzać się`, `przyjmować` an offer, and `wysyłać` a message. The message sense shares a lexeme with goods shipping while keeping a separate meaning and frame.

Batch 15 adds five pending senses: receiving a message, taking medicine, writing a document, boiling food and harvesting a crop. The message, writing and boiling senses reuse existing lexical lemmas where appropriate.

Batch 16 adds four pending multiword senses: `nie lubić`, `czerpać przyjemność`, `dochodzić do siebie` and `robić zakupy`. Frames record their selected case patterns; multiword lexical entries have no inferred morphology.

Batch 17 adds six pending event senses for watching a screen, imagining, hoping and three money-transfer events. Their selected case and adposition patterns remain candidates, without morphology or construction bindings.

Batch 18 adds verbal frames for lending, ordering goods, typing and desiring, alongside event nouns for locomotion and navigation. The nouns do not receive verb valency frames; all new bindings remain pending.

Batch 19 adds pending frames for locking, assembling, bargaining, gifting and sharing. It adds a distinct knowledge-acquisition sense to existing `uczyć się`; none of these candidates adds morphology or construction approval.

Batch 20 adds pending senses for `zachorować`, brushing teeth, exchanging goods and tissue healing, with selected argument frames. No morphology or construction binding is added.

Batch 21 adds pending senses for hiring a service, misunderstanding and spoken interpreting, with selected argument frames. No morphology or construction binding is added.

Batch 22 adds pending senses for commanding, succeeding and failing, with selected argument frames. No morphology or construction binding is added.

Batch 23 adds five pending multiword candidates: `osoba starsza`, `rodzic adopcyjny`, `członek zespołu`, `roślina kwitnąca` and `grunt rolny`. No morphology, valency frames or construction bindings are added.

Batch 24 adds twelve gender-specific lexical senses for six kinship concepts. Existing `brat` and `siostra` lexemes receive additional sibling senses; the other ten forms are separate lexemes. The six concepts each have two pending candidate bindings; no morphology, valency frames or construction bindings are added.

Batch 25 adds `małe dziecko` as a broad candidate for toddler and the event noun `przekazanie` as a narrow candidate for transfer. No morphology, valency frames or construction bindings are added.

Batch 26 adds pending lexical category senses for `część ciała` and `zdarzenie`. No morphology, valency frames or construction bindings are added.

Batch 27 adds `celowe działanie` as a pending event noun for intentional action. Batch 28 adds `komunikowanie` as a pending event noun for communication. Batch 29 adds `byt` as a pending, broad, register-limited candidate for the entity category. None adds morphology, noun valency or construction bindings.

## Morphological subset — shared class templates, explicit forms

Existing `pl.Load()` uses a small morphology seed. `pl.Generate()`/`Seed.Analyze()` work only for lexemes that have explicitly assigned compatible paradigms and principal parts.

The new `morphology_attested_nouns.go` provides **15 additional noun lexemes × 14 explicit form cells = 210 declared forms**. For every noun, the 7 cases in singular and plural are stored as individual principal parts. Five shared, gender-specific `morph.Paradigm` selectors simply retrieve the correct declared form. These are **not productive rules deriving endings for arbitrary nouns**:

- Masculine animal: `pies`.
- Masculine inanimate: `dom`, `stół`, `samochód`, `szpital`.
- Masculine personal: `nauczyciel`, `lekarz`.
- Feminine: `woda`, `szkoła`, `książka`, `ręka`, `ryba`, `pizza`.
- Neuter: `oko`, `dziecko`.

The older `kot` noun pattern remains restricted to 6 cells; the separate `lexeme:student-n` and `lexeme:czytac-v` seeds retain their existing IDs. For lexeme `pl.ryba.n`, for example, both its animal and food senses share one morphology entry; morphology alone does not choose a semantic sense. Some alternate, dialectal, archaic or context-sensitive forms have deliberately not been declared. No new verb conjugation, aspect, reflexive handling or adjective agreement is implied.

### Manual CLI checks

```sh
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.pies.n --feature number=plural --feature case=genitive --feature gender=masculine_animal --json
go run ./cmd/golingua forms --language pl --surface psów --json
go run ./cmd/golingua forms --language pl --lexeme lexeme:pl.dziecko.n --feature number=plural --feature case=instrumental --feature gender=neuter --json
```

These calls must use the actual bundled seed, not the editorial TSV. A technical smoke verifies that a form can be produced/read, **not that it is linguistically approved**.

## Hard boundaries

`languages/pl/manifest.json` still declares `sentence_analysis` and `meaning_realization` as `unsupported`. The `catalog` command separately reports the CORE, lexical and valency indexes; it is not a translation result. No language provider is installed in the CLI. The owner-delegated lexical assessments are recorded in issues #152, #154, #156, #158, #160, #162, #164, #166, #168, #170, #172, #174, #176, #178 and the ledger; CLI sentence translation and owner inflection review remain **NOT RUN**.

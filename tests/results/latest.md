# Latest language matrix result

> Observational report generated on push to `main`. Matrix outcomes do not gate the technical `go-ci` workflow.

- Source SHA: `39e482d9a59845044916bab0d060aa7936f29d9a`
- Generated (UTC): `2026-10-25T07:52:46Z`
- Command: `go test ./tests/... -count=1 -v`
- Command exit code: `1` (NONZERO)
- Registered records: 24; record checks: 72
- Additional observed checks: 147

| Case ID | Direction | Analysis | Realization | Integration | Declared coverage tags |
|---|---|---|---|---|---|
| `A1-AC-001-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, semantic_predicate, valency |
| `A1-AC-002-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, semantic_predicate, valency |
| `A1-AC-003-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, semantic_predicate, valency |
| `A1-AC-004-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, semantic_predicate, valency |
| `A1-ID-001-EN-PL` | EN → PL | PASS | PASS | PASS | determiner_ambiguity, reference, semantic_classification |
| `A1-ID-002-PL-EN` | PL → EN | PASS | PASS | PASS | determiner_ambiguity, reference, semantic_classification |
| `A1-ID-003-EN-PL` | EN → PL | PASS | PASS | PASS | reference, semantic_property_black |
| `A1-ID-004-PL-EN` | PL → EN | PASS | PASS | PASS | reference, semantic_property_black |
| `A1-LQ-001-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | locative, semantic_location |
| `A1-LQ-002-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | case, locative, semantic_location |
| `A1-LQ-003-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | exactly_two, number_morphology, semantic_quantity |
| `A1-LQ-004-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | exactly_two, number_morphology, semantic_quantity |
| `A1-NEG-001-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-NEG-002-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-NEG-003-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-NEG-004-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-PO-001-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | case, participant_roles, semantic_possession, valency |
| `A1-PO-002-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | case, participant_roles, semantic_possession, valency |
| `A1-PO-003-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | case, participant_roles, semantic_possession, valency |
| `A1-PO-004-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | case, participant_roles, semantic_possession, valency |
| `A1-Q-001-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | reference, semantic_speech_act, yes_no_question |
| `A1-Q-002-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | reference, semantic_speech_act, yes_no_question |
| `A1-Q-003-EN-PL` | EN → PL | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | reference, semantic_speech_act, wh_location |
| `A1-Q-004-PL-EN` | PL → EN | SPEC_PENDING | SPEC_PENDING | SPEC_PENDING | reference, semantic_speech_act, wh_location |

## Additional observed checks

| Package | Check | Status |
|---|---|---|
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestIdentificationMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributesReason/unresolved_time_and_multiplicity=_remain_plannable` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributesReason/unresolved_time_and_multiplicity=one_remain_plannable` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributesReason/known_past_frame_is_unsupported_even_with_At` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributesReason/known_present_frame_is_unsupported_even_with_At` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributesReason/many_experiencer_referent_is_unsupported` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributesReason/many_stimulus_referent_is_unsupported` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/present_versus_past` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/present_versus_absent_frame` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/one_versus_many` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/one_versus_absent_multiplicity` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_temporal_null` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_temporal_empty` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_temporal_unknown` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_multiplicity_null` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_multiplicity_empty` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_multiplicity_unknown` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_temporal_frame_on_referent` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/reject_multiplicity_on_predicate` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/GraphMatches_rejects_temporal_frame_changed` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/GraphMatches_rejects_temporal_frame_omitted` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/GraphMatches_rejects_multiplicity_changed` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/GraphMatches_rejects_multiplicity_omitted` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/GraphMatches_rejects_participant_roles_swapped` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/temporal_frames` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/referent_multiplicity` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestNegation` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestSeeingMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/bearer_is_object` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/bearer_is_second_cat_despite_matching_concept` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/event_time_absent` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/event_past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/subject_many` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/object_multiplicity_absent` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/property_is_timed` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/additional_property_root` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/negation_scopes_event_alongside_property` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/anchored_subject_context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/context_graph_has_a_different_bearer` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/offspring_modifiers_and_article_licensing` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/pending_modifier_mappings_do_not_silently_compose/black_equivalent_is_pending` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/pending_modifier_mappings_do_not_silently_compose/competing_pending_equivalent_exists` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-cat-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-dog-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-bird-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-cat-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/indefinite_black_subject_and_object` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/analyzed_source_candidate_realizes_without_ambiguity` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/refuse_This_is_a_water` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/refuse_This_is_a_money` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionRealization/composition.black-cat-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionRealization/composition.black-dog-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionRealization/composition.black-bird-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionRealization/composition.black-cat-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationDogAnalysis/identification.classification.this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationDogAnalysis/identification.black.this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationDogAnalysis/negation.classification.this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationDogRealization/identification.classification.this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationDogRealization/identification.black.this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationDogRealization/negation.classification.this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationCountAndArticleRefusal/mass_water_with_unknown_multiplicity` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationCountAndArticleRefusal/mass_money_with_unknown_multiplicity` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestGeneralizationCountAndArticleRefusal/count_eye_without_licensed_a_article` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestIdentificationNativeRealization/I-EN-classification` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestIdentificationNativeRealization/I-EN-black` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestIdentificationUnresolvedBlack` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestNegationRealization/I-EN-negative-classification-this` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesAnalysis/The_teacher_reads_the_message.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesAnalysis/The_teacher_read_the_message.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesAnalysis/The_teachers_read_the_message.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesRealization/H-I-EN-reading-teacher-present` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesRealization/H-I-EN-reading-teacher-past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesRealization/H-I-EN-reading-teachers-present` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesRealization/H-I-EN-reading-teachers-past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesSeeingMultiplicity/plural_teacher` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesSeeingMultiplicity/plural_dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesMassAnalysis` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesRefuseConflicts/classification_past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesRefuseConflicts/classification_many` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesEvidenceConflictAndUnknownFallback/time_conflict` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesEvidenceConflictAndUnknownFallback/number_conflict` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestMeaningAttributesMassKnownMultiplicityRefusal` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestNegationContextPolarity/negative-meaning-positive-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestNegationContextPolarity/positive-meaning-negative-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/bearer_is_object` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/bearer_is_second_cat_despite_matching_concept` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/event_time_absent` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/event_past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/subject_many` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/object_multiplicity_absent` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/property_is_timed` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/additional_property_root` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/negation_scopes_event_alongside_property` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/anchored_subject_context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionBoundaries/context_graph_has_a_different_bearer` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionAnalysis/composition.black-cat-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionAnalysis/composition.black-dog-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionAnalysis/composition.black-bird-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionAnalysis/composition.black-cat-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-cat-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-dog-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-bird-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-cat-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestGeneralizationDogAnalysis/identification.classification.to` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestGeneralizationDogAnalysis/identification.black.ten` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestGeneralizationDogAnalysis/negation.classification.to` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestGeneralizationDogRealization/identification.classification.to` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestGeneralizationDogRealization/identification.black.ten` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestGeneralizationDogRealization/negation.classification.to` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestIdentificationNativeRealization/I-PL-classification` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestIdentificationNativeRealization/I-PL-black` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestIdentificationUnresolvedBlack` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegationRealization/I-PL-negative-classification-to` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesAnalysis/H-PL-reading-teacher-present` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesAnalysis/H-PL-reading-teacher-past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesAnalysis/H-PL-reading-teachers-present` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesAnalysis/H-PL-reading-teachers-past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesRealization/H-I-PL-reading-teacher-present` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesRealization/H-I-PL-reading-teacher-past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesRealization/H-I-PL-reading-teachers-present` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesRealization/H-I-PL-reading-teachers-past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesSeeingMultiplicity/plural_teacher` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesSeeingMultiplicity/plural_dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesMassAnalysis` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesRefuseUnsupportedAndConflicting/classification_past` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesRefuseUnsupportedAndConflicting/classification_many` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesEvidenceConflictAndUnknownFallback/time_conflict` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesEvidenceConflictAndUnknownFallback/number_conflict` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMeaningAttributesMassKnownMultiplicityRefusal` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegationContextPolarity/negative-meaning-positive-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegationContextPolarity/positive-meaning-negative-context` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestRequiredSurfaceContracts` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestInventory` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestInventoryIdentifierSupportsBatches` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-dog_EN-to-PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-dog_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-dog-sees-cat_EN-to-PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-dog-sees-cat_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-bird-sees-dog_EN-to-PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-bird-sees-dog_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-cat_EN-to-PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-cat_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/English_present_to_Polish` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/English_past_to_Polish` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/Polish_plural_present_to_English` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/Polish_plural_past_to_English` | PASS |

## Package summaries

Packages with no observed test run are reported as infrastructure/build failures when Go returned `FAIL`; summaries do not replace per-check outcomes.

| Package | Go summary | Observed test names |
|---|---|---|
| `github.com/clickbytes/golingua/tests/interlingua` | PASS | 31 |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | FAIL | 102 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | FAIL | 91 |
| `github.com/clickbytes/golingua/tests/matrix` | PASS | 1 |
| `github.com/clickbytes/golingua/tests/translation` | PASS | 16 |
| `github.com/clickbytes/golingua/tests/translation/en-pl` | FAIL | 16 |
| `github.com/clickbytes/golingua/tests/translation/pl-en` | FAIL | 16 |
| `(package summary not observed)` | NOT RUN | 0 |

## Declared coverage

Declared tags are metadata attached to records; they do not prove that a construction executed successfully.

case: 5 records | clause_negation: 4 records | determiner_ambiguity: 2 records | exactly_two: 2 records | locative: 2 records | number_morphology: 2 records | participant_roles: 8 records | reference: 8 records | semantic_classification: 2 records | semantic_location: 2 records | semantic_polarity_scope: 4 records | semantic_possession: 4 records | semantic_predicate: 4 records | semantic_property_black: 2 records | semantic_quantity: 2 records | semantic_speech_act: 4 records | valency: 8 records | wh_location: 2 records | yes_no_question: 2 records

## Counts

PASS: 36 | FAIL: 0 | GAP_ANALYZE: 0 | GAP_REALIZE: 0 | GAP_I: 0 | SPEC_PENDING: 36 | INFRA: 0 | NOT RUN: 0

`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching record check result was observed. Additional checks retain their own observed statuses above. Consult the full raw log artifact for every run detail.

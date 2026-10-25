# Latest language matrix result

> Observational report generated on push to `main`. Matrix outcomes do not gate the technical `go-ci` workflow.

- Source SHA: `20cc6b8bacbe0e116cae592d849c8fb18a487b37`
- Generated (UTC): `2026-10-25T08:30:47Z`
- Command: `go test ./tests/... -count=1 -v`
- Command exit code: `1` (NONZERO)
- Registered records: 88; record checks: 264
- Additional observed checks: 155

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
| `A1-LQ-001-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | locative, semantic_location |
| `A1-LQ-002-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | case, locative, semantic_location |
| `A1-LQ-003-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_REALIZE | exactly_two, number_morphology, semantic_quantity |
| `A1-LQ-004-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_REALIZE | exactly_two, number_morphology, semantic_quantity |
| `A1-NEG-001-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-NEG-002-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-NEG-003-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-NEG-004-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1-PO-001-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | case, participant_roles, semantic_possession, valency |
| `A1-PO-002-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | case, participant_roles, semantic_possession, valency |
| `A1-PO-003-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | case, participant_roles, semantic_possession, valency |
| `A1-PO-004-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | case, participant_roles, semantic_possession, valency |
| `A1-Q-001-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | reference, semantic_speech_act, yes_no_question |
| `A1-Q-002-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | reference, semantic_speech_act, yes_no_question |
| `A1-Q-003-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | reference, semantic_speech_act, wh_location |
| `A1-Q-004-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | reference, semantic_speech_act, wh_location |
| `A1.1-AC-101-EN-PL` | EN → PL | FAIL | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-102-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-103-EN-PL` | EN → PL | FAIL | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-104-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-105-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-106-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-107-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-108-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-109-EN-PL` | EN → PL | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-110-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-111-EN-PL` | EN → PL | FAIL | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-112-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-113-EN-PL` | EN → PL | FAIL | PASS | FAIL | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-114-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-115-EN-PL` | EN → PL | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-116-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-117-EN-PL` | EN → PL | FAIL | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-118-PL-EN` | PL → EN | FAIL | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-119-EN-PL` | EN → PL | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-120-PL-EN` | PL → EN | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-121-EN-PL` | EN → PL | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-122-PL-EN` | PL → EN | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-123-EN-PL` | EN → PL | FAIL | PASS | FAIL | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-124-PL-EN` | PL → EN | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-125-EN-PL` | EN → PL | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-126-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-127-EN-PL` | EN → PL | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-128-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-129-EN-PL` | EN → PL | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-130-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-131-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-132-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.2-AC-209-PL-EN` | PL → EN | PASS | FAIL | FAIL | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AC-210-EN-PL` | EN → PL | PASS | PASS | PASS | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AC-229-PL-EN` | PL → EN | PASS | FAIL | FAIL | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AC-230-EN-PL` | EN → PL | PASS | PASS | PASS | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-ADJ-219-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_REALIZE | adj, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-ADJ-220-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_REALIZE | adj, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-ADJ-231-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_REALIZE | adj, number_morphology, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-ADJ-232-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_REALIZE | adj, number_morphology, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-AMB-215-PL-EN` | PL → EN | PASS | FAIL | FAIL | amb, homographic_past_present, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AMB-216-EN-PL` | EN → PL | PASS | PASS | PASS | amb, homographic_past_present, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-DEF-223-PL-EN` | PL → EN | PASS | PASS | PASS | def, determiner_ambiguity, participant_roles, reference, semantic_event, valency |
| `A1.2-DEF-224-EN-PL` | EN → PL | PASS | PASS | PASS | def, determiner_ambiguity, participant_roles, reference, semantic_event, valency |
| `A1.2-NEG-211-PL-EN` | PL → EN | PASS | FAIL | FAIL | clause_negation, neg, number_morphology, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-212-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, neg, number_morphology, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-225-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-226-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-227-PL-EN` | PL → EN | FAIL | PASS | FAIL | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-228-EN-PL` | EN → PL | FAIL | FAIL | FAIL | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-ORDER-221-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_REALIZE | order, participant_roles, semantic_event, valency, word_order_ovs |
| `A1.2-ORDER-222-EN-PL` | EN → PL | PASS | PASS | PASS | order, participant_roles, semantic_event, valency, word_order_ovs |
| `A1.2-PO-201-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-202-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-203-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-204-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-205-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-206-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-207-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-PO-208-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, po, semantic_event, valency |
| `A1.2-TIME-213-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, semantic_event, time, valency |
| `A1.2-TIME-214-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, semantic_event, time, valency |
| `A1.2-TIME-217-PL-EN` | PL → EN | PASS | FAIL | FAIL | participant_roles, semantic_event, time, valency |
| `A1.2-TIME-218-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, semantic_event, time, valency |

## Additional observed checks

| Package | Check | Status |
|---|---|---|
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestA12ContrastGraphsValidateRolesCountsAndScope` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCorpusEventFixturesValidateAndPreserveAuthoredStructure` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCorpusMeaningAlternativesRequireEachAllowedReading` | PASS |
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
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRepresentableMeaning/quantity.exactly-two-cats` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRepresentableMeaning/question.classification.this` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRepresentableMeaning/question.classification.to` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRelationsAreNotMeaningStates/located_in` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRelationsAreNotMeaningStates/temporary_possession` | PASS |
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
| `github.com/clickbytes/golingua/tests/interlingua` | PASS | 41 |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | FAIL | 102 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.1` | FAIL | 34 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.2` | FAIL | 34 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | FAIL | 91 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | FAIL | 34 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.2` | FAIL | 34 |
| `github.com/clickbytes/golingua/tests/matrix` | PASS | 1 |
| `github.com/clickbytes/golingua/tests/translation` | PASS | 16 |
| `github.com/clickbytes/golingua/tests/translation/en-pl` | FAIL | 48 |
| `github.com/clickbytes/golingua/tests/translation/pl-en` | FAIL | 48 |
| `(package summary not observed)` | NOT RUN | 0 |

## Declared coverage

Declared tags are metadata attached to records; they do not prove that a construction executed successfully.

ac: 4 records | adj: 4 records | amb: 2 records | case: 5 records | clause_negation: 10 records | def: 2 records | determiner_ambiguity: 4 records | exactly_two: 2 records | hearing: 8 records | homographic_past_present: 2 records | locative: 2 records | looking: 8 records | neg: 6 records | number_morphology: 12 records | order: 2 records | participant_roles: 72 records | past: 8 records | po: 8 records | present: 24 records | receiving: 8 records | reference: 10 records | seeing: 8 records | semantic_classification: 2 records | semantic_event: 64 records | semantic_location: 2 records | semantic_polarity_scope: 10 records | semantic_possession: 4 records | semantic_predicate: 4 records | semantic_property_black: 6 records | semantic_quantity: 2 records | semantic_speech_act: 4 records | time: 4 records | valency: 72 records | wh_location: 2 records | word_order_ovs: 2 records | yes_no_question: 2 records

## Counts

PASS: 153 | FAIL: 37 | GAP_ANALYZE: 11 | GAP_REALIZE: 15 | GAP_I: 48 | SPEC_PENDING: 0 | INFRA: 0 | NOT RUN: 0

`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching record check result was observed. Additional checks retain their own observed statuses above. Consult the full raw log artifact for every run detail.

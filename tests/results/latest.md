# Latest language matrix result

> Observational report generated on push to `main`. Matrix outcomes do not gate the technical `go-ci` workflow.

- Source SHA: `386f0d98ca12bde7e53880109db4f634958321c6`
- Generated (UTC): `2026-10-28T11:06:28Z`
- Command: `go test ./tests/... -count=1 -v`
- Command exit code: `1` (NONZERO)
- Registered records: 1000; record checks: 3000
- Additional observed checks: 279

| Case ID | Direction | Analysis | Realization | Integration | Declared coverage tags |
|---|---|---|---|---|---|
| `A1-AC-001-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_roles, semantic_predicate, valency |
| `A1-AC-002-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, semantic_predicate, valency |
| `A1-AC-003-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_roles, semantic_predicate, valency |
| `A1-AC-004-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, semantic_predicate, valency |
| `A1-ID-001-EN-PL` | EN → PL | PASS | PASS | PASS | determiner_ambiguity, reference, semantic_classification |
| `A1-ID-002-PL-EN` | PL → EN | PASS | PASS | PASS | determiner_ambiguity, reference, semantic_classification |
| `A1-ID-003-EN-PL` | EN → PL | PASS | PASS | PASS | reference, semantic_property_black |
| `A1-ID-004-PL-EN` | PL → EN | PASS | PASS | PASS | reference, semantic_property_black |
| `A1-LQ-001-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | locative, semantic_location |
| `A1-LQ-002-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | case, locative, semantic_location |
| `A1-LQ-003-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exactly_two, number_morphology, semantic_quantity |
| `A1-LQ-004-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exactly_two, number_morphology, semantic_quantity |
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
| `A1.1-AC-101-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-102-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-103-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-104-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-105-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-106-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-107-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-108-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, present, seeing, semantic_event, valency |
| `A1.1-AC-109-EN-PL` | EN → PL | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-110-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-111-EN-PL` | EN → PL | FAIL | PASS | FAIL | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-112-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-113-EN-PL` | EN → PL | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-114-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-115-EN-PL` | EN → PL | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-116-PL-EN` | PL → EN | PASS | PASS | PASS | hearing, participant_roles, present, semantic_event, valency |
| `A1.1-AC-117-EN-PL` | EN → PL | FAIL | PASS | FAIL | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-118-PL-EN` | PL → EN | FAIL | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-119-EN-PL` | EN → PL | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-120-PL-EN` | PL → EN | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-121-EN-PL` | EN → PL | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-122-PL-EN` | PL → EN | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-123-EN-PL` | EN → PL | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-124-PL-EN` | PL → EN | PASS | PASS | PASS | looking, participant_roles, present, semantic_event, valency |
| `A1.1-AC-125-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-126-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-127-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-128-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-129-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-130-PL-EN` | PL → EN | FAIL | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-131-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.1-AC-132-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, past, receiving, semantic_event, valency |
| `A1.10-TIME-1001-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1002-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1003-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1004-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1005-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1006-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1007-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1008-PL-EN` | PL → EN | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1009-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1010-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1011-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1012-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1013-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1014-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1015-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1016-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1017-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1018-PL-EN` | PL → EN | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1019-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1020-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1021-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1022-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1023-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1024-PL-EN` | PL → EN | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1025-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1026-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1027-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1028-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1029-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1030-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1031-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1032-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1033-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1034-PL-EN` | PL → EN | PASS | GAP_REALIZE | GAP_REALIZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1035-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1036-PL-EN` | PL → EN | PASS | FAIL | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1037-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1038-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1039-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1040-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1041-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1042-PL-EN` | PL → EN | PASS | FAIL | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1043-EN-PL` | EN → PL | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1044-PL-EN` | PL → EN | PASS | PASS | PASS | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1045-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1046-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1047-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1048-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1049-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1050-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1051-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1052-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1053-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1054-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1055-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1056-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1057-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1058-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1059-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1060-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1061-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1062-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1063-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1064-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1065-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1066-PL-EN` | PL → EN | PASS | FAIL | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1067-EN-PL` | EN → PL | FAIL | PASS | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1068-PL-EN` | PL → EN | PASS | FAIL | FAIL | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1069-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1070-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1071-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1072-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1073-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1074-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1075-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1076-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1077-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1078-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1079-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1080-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1081-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1082-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1083-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1084-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1085-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1086-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1087-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1088-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1089-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1090-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1091-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1092-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1093-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1094-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1095-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1096-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1097-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1098-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1099-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1100-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1101-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1102-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1103-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1104-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1105-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1106-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1107-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1108-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1109-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1110-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1111-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.10-TIME-1112-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_multiplicity, participant_roles, semantic_time_frame |
| `A1.11-CASE-1101-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1102-PL-EN` | PL → EN | PASS | PASS | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1103-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1104-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1105-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1106-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1107-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1108-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1109-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1110-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1111-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1112-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1113-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1114-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1115-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1116-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1117-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1118-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1119-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1120-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1121-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1122-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1123-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1124-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1125-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1126-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1127-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1128-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1129-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1130-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1131-EN-PL` | EN → PL | PASS | PASS | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1132-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1133-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1134-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1135-EN-PL` | EN → PL | FAIL | PASS | FAIL | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1136-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1137-EN-PL` | EN → PL | PASS | GAP_REALIZE | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1138-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1139-EN-PL` | EN → PL | PASS | PASS | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1140-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1141-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1142-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1143-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1144-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1145-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1146-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1147-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1148-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1149-EN-PL` | EN → PL | PASS | PASS | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1150-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1151-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1152-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1153-EN-PL` | EN → PL | PASS | PASS | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1154-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1155-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1156-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1157-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1158-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1159-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1160-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1161-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1162-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1163-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1164-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1165-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1166-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1167-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1168-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1169-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1170-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1171-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1172-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1173-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1174-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1175-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1176-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1177-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1178-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1179-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1180-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1181-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1182-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1183-EN-PL` | EN → PL | PASS | PASS | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.11-CASE-1184-PL-EN` | PL → EN | PASS | FAIL | PASS | adjective_attribute, case_order, participant_roles, semantic_predicate |
| `A1.12-AMB-1201-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1202-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1203-EN-PL` | EN → PL | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1204-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1205-EN-PL` | EN → PL | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1206-PL-EN` | PL → EN | FAIL | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1207-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1208-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1209-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1210-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1211-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1212-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1213-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1214-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1215-EN-PL` | EN → PL | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1216-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1217-EN-PL` | EN → PL | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1218-PL-EN` | PL → EN | FAIL | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1219-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1220-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1221-EN-PL` | EN → PL | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1222-PL-EN` | PL → EN | FAIL | FAIL | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1223-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1224-PL-EN` | PL → EN | PASS | FAIL | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1225-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1226-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1227-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1228-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1229-EN-PL` | EN → PL | PASS | GAP_REALIZE | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1230-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1231-EN-PL` | EN → PL | PASS | GAP_REALIZE | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1232-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1233-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1234-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1235-EN-PL` | EN → PL | PASS | GAP_REALIZE | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1236-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1237-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1238-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1239-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1240-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1241-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1242-PL-EN` | PL → EN | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1243-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1244-PL-EN` | PL → EN | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1245-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1246-PL-EN` | PL → EN | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1247-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1248-PL-EN` | PL → EN | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1249-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1250-PL-EN` | PL → EN | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1251-EN-PL` | EN → PL | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1252-PL-EN` | PL → EN | FAIL | PASS | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1253-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1254-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1255-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1256-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1257-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1258-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1259-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1260-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1261-EN-PL` | EN → PL | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1262-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1263-EN-PL` | EN → PL | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1264-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1265-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1266-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1267-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1268-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1269-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1270-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1271-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1272-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1273-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1274-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1275-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1276-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1277-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1278-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1279-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1280-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1281-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1282-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1283-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1284-PL-EN` | PL → EN | PASS | PASS | PASS | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1285-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.12-AMB-1286-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | predicate_roles, reference_binding, semantic_ambiguity |
| `A1.13-COMP-1301-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1302-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1303-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1304-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1305-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1306-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1307-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1308-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1309-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1310-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1311-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1312-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1313-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1314-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1315-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1316-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1317-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1318-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1319-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1320-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1321-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1322-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1323-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1324-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1325-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1326-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1327-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1328-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1329-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1330-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1331-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1332-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1333-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1334-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1335-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1336-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1337-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1338-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1339-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1340-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1341-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1342-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1343-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1344-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1345-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1346-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1347-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1348-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1349-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1350-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1351-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1352-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1353-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1354-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1355-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1356-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1357-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1358-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1359-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1360-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1361-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1362-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1363-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1364-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1365-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.13-COMP-1366-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.2-AC-209-PL-EN` | PL → EN | PASS | PASS | PASS | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AC-210-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AC-229-PL-EN` | PL → EN | PASS | PASS | PASS | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AC-230-EN-PL` | EN → PL | PASS | PASS | PASS | ac, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-ADJ-219-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adj, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-ADJ-220-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adj, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-ADJ-231-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adj, number_morphology, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-ADJ-232-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | adj, number_morphology, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.2-AMB-215-PL-EN` | PL → EN | PASS | PASS | PASS | amb, homographic_past_present, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-AMB-216-EN-PL` | EN → PL | PASS | PASS | PASS | amb, homographic_past_present, number_morphology, participant_roles, semantic_event, valency |
| `A1.2-DEF-223-PL-EN` | PL → EN | PASS | PASS | PASS | def, determiner_ambiguity, participant_roles, reference, semantic_event, valency |
| `A1.2-DEF-224-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | def, determiner_ambiguity, participant_roles, reference, semantic_event, valency |
| `A1.2-NEG-211-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, neg, number_morphology, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-212-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, neg, number_morphology, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-225-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-226-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-227-PL-EN` | PL → EN | FAIL | PASS | PASS | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-NEG-228-EN-PL` | EN → PL | FAIL | PASS | FAIL | clause_negation, neg, participant_roles, semantic_event, semantic_polarity_scope, valency |
| `A1.2-ORDER-221-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | order, participant_roles, semantic_event, valency, word_order_ovs |
| `A1.2-ORDER-222-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | order, participant_roles, semantic_event, valency, word_order_ovs |
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
| `A1.2-TIME-217-PL-EN` | PL → EN | PASS | PASS | PASS | participant_roles, semantic_event, time, valency |
| `A1.2-TIME-218-EN-PL` | EN → PL | PASS | PASS | PASS | participant_roles, semantic_event, time, valency |
| `A1.3-CO-301-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-302-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-303-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-304-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-305-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, distinct_referents, participant_roles, semantic_event, valency |
| `A1.3-CO-306-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, distinct_referents, participant_roles, semantic_event, valency |
| `A1.3-CO-307-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-308-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-309-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-310-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-311-PL-EN` | PL → EN | PASS | PASS | PASS | composition, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-312-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | composition, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-313-PL-EN` | PL → EN | PASS | PASS | PASS | composition, distinct_referents, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-314-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | composition, distinct_referents, participant_roles, semantic_event, semantic_property_black, valency |
| `A1.3-CO-315-PL-EN` | PL → EN | PASS | PASS | PASS | composition, participant_roles, semantic_event, valency, word_order_ovs |
| `A1.3-CO-316-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | composition, participant_roles, semantic_event, valency, word_order_ovs |
| `A1.3-CO-317-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | clause_negation, composition, participant_roles, semantic_event, semantic_polarity_scope, semantic_property_black |
| `A1.3-CO-318-EN-PL` | EN → PL | PASS | GAP_REALIZE | FAIL | clause_negation, composition, participant_roles, semantic_event, semantic_polarity_scope, semantic_property_black |
| `A1.3-CO-319-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, composition, participant_roles, semantic_event, semantic_polarity_scope, semantic_property_black |
| `A1.3-CO-320-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, composition, participant_roles, semantic_event, semantic_polarity_scope, semantic_property_black |
| `A1.3-CO-321-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-322-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-323-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.3-CO-324-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | composition, participant_roles, semantic_event, shared_referent, valency |
| `A1.4-AC-401-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-402-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-403-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-404-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-405-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-406-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-407-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-408-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-409-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-410-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-411-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-412-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-413-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-414-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-415-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-416-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-417-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-418-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-419-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-420-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-421-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-422-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-423-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-424-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-425-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-426-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-427-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-428-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-429-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-430-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-431-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-432-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-433-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-434-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-435-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-436-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-437-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-438-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-439-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-440-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-441-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-442-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-443-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-444-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-445-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-446-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-447-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-448-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-449-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-450-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-451-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-452-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-453-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-454-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-455-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-456-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-457-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-458-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-459-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-460-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-461-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-462-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-463-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-464-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-465-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-466-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-467-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-468-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-469-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-470-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-471-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-472-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-473-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-474-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-475-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-476-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-477-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-478-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-479-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-480-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-481-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-482-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-483-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-484-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-485-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-486-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-487-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-488-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-489-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-490-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-491-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-492-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-493-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.4-AC-494-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | participant_roles, semantic_event, valency |
| `A1.5-ID-501-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-502-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-503-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-504-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-505-EN-PL` | EN → PL | FAIL | PASS | FAIL | deictic_reference, reference, semantic_classification |
| `A1.5-ID-506-PL-EN` | PL → EN | FAIL | PASS | FAIL | deictic_reference, reference, semantic_classification |
| `A1.5-ID-507-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-508-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-509-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-510-PL-EN` | PL → EN | PASS | GAP_REALIZE | GAP_REALIZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-511-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-512-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-513-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-514-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-515-EN-PL` | EN → PL | GAP_ANALYZE | PASS | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-516-PL-EN` | PL → EN | PASS | GAP_REALIZE | GAP_REALIZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-517-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-518-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-519-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-520-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-521-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-522-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-523-EN-PL` | EN → PL | FAIL | PASS | FAIL | deictic_reference, reference, semantic_classification |
| `A1.5-ID-524-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-525-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-526-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-527-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-528-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-529-EN-PL` | EN → PL | GAP_ANALYZE | PASS | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-530-PL-EN` | PL → EN | PASS | GAP_REALIZE | GAP_REALIZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-531-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-532-PL-EN` | PL → EN | FAIL | GAP_REALIZE | GAP_REALIZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-533-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-534-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-535-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-536-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_classification |
| `A1.5-ID-537-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-538-PL-EN` | PL → EN | FAIL | PASS | FAIL | deictic_reference, reference, semantic_classification |
| `A1.5-ID-539-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-540-PL-EN` | PL → EN | PASS | GAP_REALIZE | GAP_REALIZE | deictic_reference, reference, semantic_classification |
| `A1.5-ID-541-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_classification |
| `A1.5-ID-542-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_classification |
| `A1.5-ID-543-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_classification |
| `A1.5-ID-544-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_classification |
| `A1.5-ID-545-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-546-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-547-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-548-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-549-EN-PL` | EN → PL | GAP_ANALYZE | PASS | GAP_ANALYZE | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-550-PL-EN` | PL → EN | PASS | GAP_REALIZE | FAIL | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-551-EN-PL` | EN → PL | PASS | GAP_REALIZE | GAP_REALIZE | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-552-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-553-EN-PL` | EN → PL | FAIL | GAP_REALIZE | FAIL | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-554-PL-EN` | PL → EN | GAP_ANALYZE | PASS | GAP_ANALYZE | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-555-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-556-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-557-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-558-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-559-EN-PL` | EN → PL | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-560-PL-EN` | PL → EN | PASS | PASS | PASS | deictic_reference, reference, semantic_property_black |
| `A1.5-ID-561-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_property_black |
| `A1.5-ID-562-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_property_black |
| `A1.5-ID-563-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_property_black |
| `A1.5-ID-564-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_property_black |
| `A1.5-ID-565-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_classification |
| `A1.5-ID-566-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_classification |
| `A1.5-ID-567-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_property_black |
| `A1.5-ID-568-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | deictic_reference, number_morphology, reference, semantic_property_black |
| `A1.5-ID-569-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-570-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-571-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-572-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-573-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-574-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-575-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.5-ID-576-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | deictic_reference, reference, semantic_color_gap |
| `A1.6-PO-601-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-602-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-603-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-604-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-605-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-606-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-607-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-608-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-609-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-610-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-611-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-612-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-613-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-614-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-615-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, participant_roles, possession, semantic_polarity_scope, semantic_relation |
| `A1.6-PO-616-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, participant_roles, possession, semantic_polarity_scope, semantic_relation |
| `A1.6-PO-617-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, number_morphology, participant_roles, possession, scope_ambiguity, semantic_polarity_scope, semantic_relation |
| `A1.6-PO-618-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, number_morphology, participant_roles, possession, scope_ambiguity, semantic_polarity_scope, semantic_relation |
| `A1.6-PO-619-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, number_morphology, participant_roles, possession, scope_ambiguity, semantic_polarity_scope, semantic_relation |
| `A1.6-PO-620-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, number_morphology, participant_roles, possession, scope_ambiguity, semantic_polarity_scope, semantic_relation |
| `A1.6-PO-621-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-622-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-623-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-624-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, scope_ambiguity, semantic_relation |
| `A1.6-PO-625-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-626-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-627-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-628-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-629-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-630-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-631-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-632-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-633-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-634-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-635-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-636-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-637-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-638-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-639-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-640-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-641-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-642-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-643-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-644-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-645-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-646-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-647-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-648-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-649-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, possessive_determiner, reference, scope_ambiguity, semantic_relation |
| `A1.6-PO-650-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, possessive_determiner, reference, scope_ambiguity, semantic_relation |
| `A1.6-PO-651-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, possessive_determiner, reference, scope_ambiguity, semantic_relation |
| `A1.6-PO-652-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, possessive_determiner, reference, scope_ambiguity, semantic_relation |
| `A1.6-PO-653-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-654-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-655-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-656-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, possessive_determiner, reference, semantic_relation |
| `A1.6-PO-657-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, possessive_determiner, reference, scope_ambiguity, semantic_relation |
| `A1.6-PO-658-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, possession, possessive_determiner, reference, scope_ambiguity, semantic_relation |
| `A1.6-PO-659-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-660-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-661-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-662-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, possession, semantic_relation |
| `A1.6-PO-663-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | kinship_social_relation, participant_roles, possession, reference, semantic_relation |
| `A1.6-PO-664-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | kinship_social_relation, participant_roles, possession, reference, semantic_relation |
| `A1.6-PO-665-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | kinship_social_relation, participant_roles, possession, reference, semantic_relation |
| `A1.6-PO-666-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | kinship_social_relation, participant_roles, possession, reference, semantic_relation |
| `A1.6-PO-667-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | kinship_social_relation, participant_roles, possession, reference, semantic_relation |
| `A1.6-PO-668-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | kinship_social_relation, participant_roles, possession, reference, semantic_relation |
| `A1.7-LQ-701-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-702-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-703-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-704-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-705-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-706-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-707-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-708-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-709-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-710-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-711-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-712-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-713-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-714-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-715-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-716-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-717-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-718-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-719-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-720-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-721-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-722-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-723-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-724-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-725-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-726-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-727-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-728-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-729-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-730-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-731-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-732-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-733-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-734-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-735-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-736-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, spatial_order, temporal_reference |
| `A1.7-LQ-737-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-738-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-739-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-740-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-741-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-742-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | number_morphology, participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-743-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-744-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-745-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-746-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-747-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-748-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | participant_roles, semantic_spatial_relation, temporal_reference |
| `A1.7-LQ-749-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-750-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-751-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-752-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-753-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-754-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-755-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-756-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-757-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-758-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-759-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-760-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-761-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-762-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-763-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-764-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-765-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-766-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-767-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-768-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-769-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-770-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-771-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-772-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-773-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-774-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-775-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-776-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-777-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-778-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-779-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-780-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-781-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-782-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-783-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-784-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-785-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-786-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-787-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-788-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-789-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-790-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-791-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-792-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-793-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-794-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-795-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.7-LQ-796-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | exact_quantity, semantic_existential, semantic_quantity, temporal_reference |
| `A1.8-NEG-801-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-802-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-803-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-804-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-805-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-806-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-807-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-808-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-809-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-810-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-811-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-812-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-813-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-814-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-815-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-816-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-817-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-818-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-819-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-820-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-821-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-822-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-823-EN-PL` | EN → PL | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-824-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-825-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-826-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-827-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-828-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-829-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-830-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-831-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-832-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-833-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-834-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-835-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-836-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-837-EN-PL` | EN → PL | GAP_ANALYZE | PASS | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-838-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-839-EN-PL` | EN → PL | FAIL | PASS | FAIL | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-840-PL-EN` | PL → EN | PASS | PASS | PASS | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-841-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-842-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-843-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-844-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-845-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-846-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-847-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-848-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-849-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-850-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-851-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-852-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-853-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-854-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-855-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-856-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-857-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-858-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-859-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-860-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-861-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-862-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-863-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-864-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-865-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-866-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-867-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-868-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-869-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-870-PL-EN` | PL → EN | FAIL | GAP_REALIZE | FAIL | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-871-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-872-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-873-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-874-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-875-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-876-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-877-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-878-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-879-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-880-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-881-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-882-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-883-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-884-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-885-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-886-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-887-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-888-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-889-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.8-NEG-890-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | clause_negation, semantic_polarity_scope |
| `A1.9-Q-1000-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-1001-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-1002-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-1003-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-1004-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-1005-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-1006-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-1007-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | polarity_scope, semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-1008-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | polarity_scope, semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-1009-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | polarity_scope, semantic_speech_act, yes_no_question |
| `A1.9-Q-1010-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | polarity_scope, semantic_speech_act, yes_no_question |
| `A1.9-Q-1011-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | polarity_scope, semantic_speech_act, yes_no_question |
| `A1.9-Q-1012-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | polarity_scope, semantic_speech_act, yes_no_question |
| `A1.9-Q-1013-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-1014-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-1015-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-1016-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-901-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-902-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-903-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-904-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-905-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-906-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-907-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-908-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-909-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-910-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-911-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-912-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-913-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-914-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-915-EN-PL` | EN → PL | GAP_ANALYZE | FAIL | GAP_ANALYZE | polarity_scope, semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-916-PL-EN` | PL → EN | GAP_ANALYZE | FAIL | GAP_ANALYZE | polarity_scope, semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-917-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | polarity_scope, semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-918-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | polarity_scope, semantic_event, semantic_speech_act, yes_no_question |
| `A1.9-Q-919-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | polarity_scope, semantic_speech_act, yes_no_question |
| `A1.9-Q-920-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | polarity_scope, semantic_speech_act, yes_no_question |
| `A1.9-Q-921-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-922-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-923-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-924-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-925-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-926-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-927-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-928-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-929-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-930-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-931-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-932-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-933-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-934-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-935-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-936-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-937-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-938-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-939-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-940-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-941-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-942-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-943-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-944-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-945-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-946-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-947-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-948-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-949-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-950-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-951-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-952-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-953-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-954-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-955-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-956-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-957-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-958-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-959-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-960-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-961-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-962-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-963-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-964-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-965-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-966-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-967-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-968-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-969-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-970-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-971-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-972-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-973-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-974-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-975-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-976-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-977-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-978-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-979-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, yes_no_question |
| `A1.9-Q-980-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, yes_no_question |
| `A1.9-Q-981-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, yes_no_question |
| `A1.9-Q-982-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, yes_no_question |
| `A1.9-Q-983-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-984-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-985-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-986-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-987-EN-PL` | EN → PL | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-988-PL-EN` | PL → EN | GAP_I | GAP_I | GAP_I | semantic_speech_act, wh_binding, yes_no_question |
| `A1.9-Q-989-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-990-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-991-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-992-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-993-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-994-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_classification, semantic_speech_act, yes_no_question |
| `A1.9-Q-995-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-996-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-997-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-998-PL-EN` | PL → EN | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |
| `A1.9-Q-999-EN-PL` | EN → PL | GAP_ANALYZE | GAP_REALIZE | GAP_ANALYZE | semantic_property, semantic_speech_act, yes_no_question |

## Boundary refusal checks

These malformed or unsupported inputs are separate from normal registered translation records and do not contribute to record PASS totals.

- Total: 32; PASS: 32

| Category | Package | Check | Status |
|---|---|---|---|
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-001` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-002` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-003` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-004` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-005` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-006` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-007` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/en/A1.1` | `TestRefusalAnalysis/REF-EN-008` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-001` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-002` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-003` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-004` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-005` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-006` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-007` | PASS |
| native analysis refusal | `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | `TestRefusalAnalysis/REF-PL-008` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-001` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-002` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-003` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-004` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-005` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-006` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-007` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-EN-008` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-001` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-002` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-003` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-004` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-005` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-006` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-007` | PASS |
| translation refusal | `github.com/clickbytes/golingua/tests/translation` | `TestRefusalTranslation/REF-PL-008` | PASS |

## Additional observed checks

| Package | Check | Status |
|---|---|---|
| `github.com/clickbytes/golingua/tests/interlingua` | `TestActionExpansionFixturesPreserveAuthoredRolesAndMultiplicity` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestQuestionExpansionFixturesValidate` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestQuestionExpansionDoesNotInventWHVariableFixtures` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestAmbiguityExpansionCatalogValidatesAndHasNoPerSourceAliases` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestAmbiguityExpansionReadFramesAndChildVariants` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestAmbiguityExpansionPolishSyncreticRoleSwap` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestAmbiguityExpansionAnchorsQuantitiesAndDistinctMentions` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCaseOrderExpansionIndependentGraphsValidate` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionBatchFixturesValidateAndPreserveReferents` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionExpansionCatalogueValidatesAndHasUniqueVariants` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionExpansionPreservesFramesRolesNegationAndAttributes` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionExpansionRequestActsAndParticipantAnchors` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCompositionMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestA12ContrastGraphsValidateRolesCountsAndScope` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCorpusEventFixturesValidateAndPreserveAuthoredStructure` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCorpusMeaningAlternativesRequireEachAllowedReading` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestIdentificationExpansionFixturesAndMissingColorGap` | PASS |
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
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/referent_multiplicity` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestMeaningAttributes/temporal_frames` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestNegationExpansionFixturesValidateAndRetainScope` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestNegationExpansionGapLabelsHaveNoFixture` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestNegationExpansionPluralOnlyDoorCountVariants` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestNoCatSentenceNeedsParticipantQuantification` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestNegation` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRepresentableMeaning/quantity.exactly-two-cats.component` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRepresentableMeaning/question.classification.this` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRepresentableMeaning/question.classification.to` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRelationsAreNotMeaningStates/located_in` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPendingDomainRelationsAreNotMeaningStates/temporary_possession` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPossessionAndKinshipRelationsAreNotMeaningStates/temporary_possession` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPossessionAndKinshipRelationsAreNotMeaningStates/legal_ownership` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPossessionAndKinshipRelationsAreNotMeaningStates/siblinghood` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestPossessionAndKinshipRelationsAreNotMeaningStates/friendship_relationship` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestQuantifierAndReferentCannotCarryExistenceTime/quantifier` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestQuantifierAndReferentCannotCarryExistenceTime/referent` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestCoreExistenceCannotBindCountedReferent` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestWomanConceptIsNotAvailableForCountContracts` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestSeeingMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestTimeExpansionFixturesValidateAndPreserveSourceAttributes` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestTimeExpansionPersonAnchorsAndLanguageSpecificCounts` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestTimeExpansionChildAndDoorDistinctionsAndExplicitGaps` | PASS |
| `github.com/clickbytes/golingua/tests/interlingua` | `TestTimeExpansionEnumeratedVariantsAreUniqueAndReadIsAmbiguous` | PASS |
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
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/offspring_modifiers_and_article_licensing` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/pending_modifier_mappings_do_not_silently_compose/black_equivalent_is_pending` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionBoundaries/pending_modifier_mappings_do_not_silently_compose/competing_pending_equivalent_exists` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-cat-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-dog-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-bird-sees-dog` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/composition.black-cat-sees-cat` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/indefinite_black_subject_and_object` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestCompositionAnalysis/analyzed_source_candidate_realizes_without_ambiguity` | FAIL |
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
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersIndependentSeeingAndReading/a1.12.amb.007.pl` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersIndependentSeeingAndReading/a1.12.amb.003.pl` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersPreserveActualEnglishSource/The_teachers_see_the_birds.` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersPreserveActualEnglishSource/Teachers_see_birds.` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersPreserveActualEnglishSource/A_mother_reads_a_letter.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersSeparateNominalAndPersonalReadingSources/I` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersSeparateNominalAndPersonalReadingSources/We` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersSeparateNominalAndPersonalReadingSources/You` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersNegativeIndependentChoices/a1.8.neg.019` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersNegativeIndependentChoices/a1.8.neg.020` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersParticipantsAndThisControl` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersRejectMalformedActualSources/Teacher_sees_the_cat.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersRejectMalformedActualSources/The_teacher_sees_dog.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersRejectMalformedActualSources/The_teacher_sees_a_dogs.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersRejectMalformedActualSources/The_teacher_sees_an_dog.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | `TestPluralDeterminersRejectMalformedActualSources/The_teacher_sees_a_apple.` | PASS |
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
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-cat-sees-dog` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-dog-sees-cat` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-bird-sees-dog` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCompositionRealization/composition.black-cat-sees-cat` | FAIL |
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
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegativeCaseRealizationContrasts/negative-seeing-book` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegativeCaseRealizationContrasts/negative-hearing-birds` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegativeCaseRealizationContrasts/negative-looking-na-accusative` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegativeSeeingBookAnalysisKeepsSingularStimulus` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegationContextPolarity/negative-meaning-positive-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNegationContextPolarity/positive-meaning-negative-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNeuterBlackAttributionUsesTo/a1.5.identification.029.en` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNeuterBlackAttributionUsesTo/a1.5.identification.030.en` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCountNounReferentsInOrdinaryClauses/Nauczyciel_widzi_okno.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCountNounReferentsInOrdinaryClauses/Nauczyciel_widzi_okna.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCountNounReferentsInOrdinaryClauses/Nauczyciel_widzi_jabłko.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCountNounReferentsInOrdinaryClauses/Nauczyciel_widzi_jabłka.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCountNounReferentsInOrdinaryClauses/Nauczyciel_czyta_list.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestCountNounReferentsInOrdinaryClauses/Nauczyciel_czyta_listy.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestMotherReadLetterPastMeaning` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestNeuterAttributionSourceAndTranslation` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Widzę_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Widzimy_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Widzisz_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Widzicie_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Ja_widzę_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/My_widzimy_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Ty_widzisz_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalSourcePreservesParticipantAnchors/Wy_widzicie_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalKnownSpeakerHasOnePolishTarget` | FAIL |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestEnglishYouFixturesKeepCountUnknownAndPolishTargetsRemainIndependent` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPersonalIndependentPastReadingUsesGenderEvidenceAssumptions` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPolishPastPersonalSourceGenderRestrictsItsTarget` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestParticipantAnchorMisalignmentFailsClosedForTargets/en/conflicting-same-actor-anchors` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestParticipantAnchorMisalignmentFailsClosedForTargets/en/stale-first-paired-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestParticipantAnchorMisalignmentFailsClosedForTargets/en/participant-anchor-on-object` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestParticipantAnchorMisalignmentFailsClosedForTargets/pl/conflicting-same-actor-anchors` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestParticipantAnchorMisalignmentFailsClosedForTargets/pl/stale-first-paired-context` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestParticipantAnchorMisalignmentFailsClosedForTargets/pl/participant-anchor-on-object` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPolishImplicitRoleEvidenceRejectsDuplicateAndStaleProofs/duplicate` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPolishImplicitRoleEvidenceRejectsDuplicateAndStaleProofs/stale-derivation` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingIndependentIToPolish/a1.10.time.017.en` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingIndependentIToPolish/a1.10.time.018.en` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingPolishSourceAndAgreement/Nauczyciel_widział_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingPolishSourceAndAgreement/Nauczyciele_widzieli_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingPolishSourceAndAgreement/Matka_widziała_psa.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingPolishSourceAndAgreement/Psy_widziały_kota.` | PASS |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | `TestPastSeeingEnglishSourceAndCrossLanguageTranslation` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestRequiredSurfaceContracts` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestRequiredTextGroups` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestContextMatchesBindsAnchorsToGraphRoles` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestContextMatchesRejectsMissingExtraWrongAndResolvedMentions/missing` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestContextMatchesRejectsMissingExtraWrongAndResolvedMentions/extra` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestContextMatchesRejectsMissingExtraWrongAndResolvedMentions/wrong-anchor` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestContextMatchesRejectsMissingExtraWrongAndResolvedMentions/resolved-id` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestContextMatchesIgnoresOnlyUnanchoredUnresolvedMentions` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestMatchAllowedCandidateUsesMatchingVariantContext` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestMatchAllowedCandidatePropagatesInvalidExpectedContextAndBudget` | PASS |
| `github.com/clickbytes/golingua/tests/matrix` | `TestQuestionFixtureLoaderRoutesOnlyAnchoredContexts` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestAmbiguityCatalogueLinks` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestInventory` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestInventoryIdentifierSupportsBatches` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestA18NegationLinksHaveMatchingMeaningContracts` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestAmbiguityAssumptionsFollowRoles` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-source-en-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-source-en-02` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-source-en-03` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-negative-source-en-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-source-pl-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-source-pl-02` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-source-pl-03` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-negative-source-pl-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-target-pl-01` | GAP_REALIZE |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-target-pl-02` | GAP_REALIZE |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-target-pl-03` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-negative-target-pl-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-target-en-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-target-en-02` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-target-en-03` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestChildAgeAndOffspringRoutes/R6-child-negative-target-en-01` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-dog_EN-to-PL` | FAIL |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-dog_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-dog-sees-cat_EN-to-PL` | FAIL |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-dog-sees-cat_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-bird-sees-dog_EN-to-PL` | FAIL |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-bird-sees-dog_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-cat_EN-to-PL` | FAIL |
| `github.com/clickbytes/golingua/tests/translation` | `TestCompositionTranslation/composition.black-cat-sees-cat_PL-to-EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/English_present_to_Polish` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/English_past_to_Polish` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/Polish_plural_present_to_English` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMeaningAttributesReadingTranslation/Polish_plural_past_to_English` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/role_swap_PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/role_swap_EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/number_PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/number_EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/tense_PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/tense_EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/polarity_PL` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/polarity_EN` | PASS |
| `github.com/clickbytes/golingua/tests/translation` | `TestMetamorphicProperties/SVO_to_OVS_PL` | GAP_ANALYZE |

## Package summaries

Packages with no observed test run are reported as infrastructure/build failures when Go returned `FAIL`; summaries do not replace per-check outcomes.

| Package | Go summary | Observed test names |
|---|---|---|
| `github.com/clickbytes/golingua/tests/interlingua` | PASS | 72 |
| `github.com/clickbytes/golingua/tests/languages/en/A1` | FAIL | 123 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.1` | FAIL | 43 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.10` | FAIL | 116 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.11` | FAIL | 86 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.12` | FAIL | 88 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.13` | FAIL | 68 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.2` | FAIL | 34 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.3` | FAIL | 26 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.4` | FAIL | 96 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.5` | FAIL | 78 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.6` | FAIL | 70 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.7` | FAIL | 98 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.8` | FAIL | 92 |
| `github.com/clickbytes/golingua/tests/languages/en/A1.9` | FAIL | 118 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1` | FAIL | 140 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.1` | FAIL | 43 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.10` | FAIL | 116 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.11` | FAIL | 86 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.12` | FAIL | 88 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.13` | FAIL | 68 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.2` | FAIL | 34 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.3` | FAIL | 26 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.4` | FAIL | 96 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.5` | FAIL | 78 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.6` | FAIL | 70 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.7` | FAIL | 98 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.8` | FAIL | 92 |
| `github.com/clickbytes/golingua/tests/languages/pl/A1.9` | FAIL | 118 |
| `github.com/clickbytes/golingua/tests/matrix` | PASS | 12 |
| `github.com/clickbytes/golingua/tests/translation` | FAIL | 63 |
| `github.com/clickbytes/golingua/tests/translation/en-pl` | FAIL | 504 |
| `github.com/clickbytes/golingua/tests/translation/pl-en` | FAIL | 504 |
| `(package summary not observed)` | NOT RUN | 0 |

## Declared coverage

Declared tags are authored metadata; they do not prove construction execution or correctness.

ac: 4 records | adj: 4 records | adjective_attribute: 84 records | amb: 2 records | case: 5 records | case_order: 84 records | clause_negation: 110 records | composition: 90 records | def: 2 records | deictic_reference: 76 records | determiner_ambiguity: 4 records | distinct_referents: 4 records | exact_quantity: 48 records | exactly_two: 2 records | hearing: 8 records | homographic_past_present: 2 records | kinship_social_relation: 6 records | locative: 2 records | looking: 8 records | neg: 6 records | number_morphology: 50 records | order: 2 records | participant_multiplicity: 112 records | participant_roles: 568 records | past: 8 records | po: 8 records | polarity_scope: 12 records | possession: 68 records | possessive_determiner: 18 records | predicate_roles: 86 records | present: 24 records | receiving: 8 records | reference: 110 records | reference_binding: 86 records | scope_ambiguity: 20 records | seeing: 8 records | semantic_ambiguity: 86 records | semantic_classification: 58 records | semantic_color_gap: 8 records | semantic_event: 264 records | semantic_existential: 48 records | semantic_location: 2 records | semantic_polarity_scope: 110 records | semantic_possession: 4 records | semantic_predicate: 88 records | semantic_property: 14 records | semantic_property_black: 40 records | semantic_quantity: 50 records | semantic_relation: 68 records | semantic_spatial_relation: 48 records | semantic_speech_act: 120 records | semantic_time_frame: 112 records | shared_referent: 74 records | spatial_order: 18 records | temporal_reference: 96 records | time: 4 records | valency: 252 records | wh_binding: 66 records | wh_location: 2 records | word_order_ovs: 4 records | yes_no_question: 118 records

## Pair and group balance

Pair/group counts describe sample distribution. Coverage targets are planning goals, not correctness gates.

- Declared unique pairs: 500; target total: 500

| Group | Observed pairs | Planning target |
|---|---:|---:|
| actions | 65 | 65 |
| ambiguity_reference | 45 | 45 |
| case_adjective_order | 45 | 45 |
| composition_boundary | 45 | 45 |
| identification | 40 | 40 |
| location_quantity | 50 | 50 |
| negation | 50 | 50 |
| possession | 40 | 40 |
| questions | 60 | 60 |
| time_agreement | 60 | 60 |

### Direction outcomes by group and stage

| Group | Direction | Stage | Check outcome | Records |
|---|---|---|---|---:|
| actions | en → pl | analysis | FAIL | 8 |
| actions | en → pl | analysis | GAP_ANALYZE | 47 |
| actions | en → pl | analysis | PASS | 10 |
| actions | en → pl | integration | FAIL | 11 |
| actions | en → pl | integration | GAP_ANALYZE | 47 |
| actions | en → pl | integration | PASS | 7 |
| actions | en → pl | realization | GAP_REALIZE | 53 |
| actions | en → pl | realization | PASS | 12 |
| actions | pl → en | analysis | FAIL | 6 |
| actions | pl → en | analysis | GAP_ANALYZE | 47 |
| actions | pl → en | analysis | PASS | 12 |
| actions | pl → en | integration | GAP_ANALYZE | 47 |
| actions | pl → en | integration | PASS | 18 |
| actions | pl → en | realization | GAP_REALIZE | 47 |
| actions | pl → en | realization | PASS | 18 |
| ambiguity_reference | en → pl | analysis | FAIL | 5 |
| ambiguity_reference | en → pl | analysis | GAP_ANALYZE | 13 |
| ambiguity_reference | en → pl | analysis | GAP_I | 7 |
| ambiguity_reference | en → pl | analysis | PASS | 20 |
| ambiguity_reference | en → pl | integration | FAIL | 14 |
| ambiguity_reference | en → pl | integration | GAP_ANALYZE | 13 |
| ambiguity_reference | en → pl | integration | GAP_I | 7 |
| ambiguity_reference | en → pl | integration | PASS | 11 |
| ambiguity_reference | en → pl | realization | FAIL | 2 |
| ambiguity_reference | en → pl | realization | GAP_I | 7 |
| ambiguity_reference | en → pl | realization | GAP_REALIZE | 23 |
| ambiguity_reference | en → pl | realization | PASS | 13 |
| ambiguity_reference | pl → en | analysis | FAIL | 9 |
| ambiguity_reference | pl → en | analysis | GAP_ANALYZE | 14 |
| ambiguity_reference | pl → en | analysis | GAP_I | 7 |
| ambiguity_reference | pl → en | analysis | PASS | 15 |
| ambiguity_reference | pl → en | integration | FAIL | 8 |
| ambiguity_reference | pl → en | integration | GAP_ANALYZE | 14 |
| ambiguity_reference | pl → en | integration | GAP_I | 7 |
| ambiguity_reference | pl → en | integration | PASS | 16 |
| ambiguity_reference | pl → en | realization | FAIL | 8 |
| ambiguity_reference | pl → en | realization | GAP_I | 7 |
| ambiguity_reference | pl → en | realization | GAP_REALIZE | 7 |
| ambiguity_reference | pl → en | realization | PASS | 23 |
| case_adjective_order | en → pl | analysis | FAIL | 4 |
| case_adjective_order | en → pl | analysis | GAP_ANALYZE | 25 |
| case_adjective_order | en → pl | analysis | GAP_I | 7 |
| case_adjective_order | en → pl | analysis | PASS | 9 |
| case_adjective_order | en → pl | integration | FAIL | 7 |
| case_adjective_order | en → pl | integration | GAP_ANALYZE | 25 |
| case_adjective_order | en → pl | integration | GAP_I | 7 |
| case_adjective_order | en → pl | integration | PASS | 6 |
| case_adjective_order | en → pl | realization | GAP_I | 7 |
| case_adjective_order | en → pl | realization | GAP_REALIZE | 32 |
| case_adjective_order | en → pl | realization | PASS | 6 |
| case_adjective_order | pl → en | analysis | GAP_ANALYZE | 36 |
| case_adjective_order | pl → en | analysis | GAP_I | 7 |
| case_adjective_order | pl → en | analysis | PASS | 2 |
| case_adjective_order | pl → en | integration | GAP_ANALYZE | 36 |
| case_adjective_order | pl → en | integration | GAP_I | 7 |
| case_adjective_order | pl → en | integration | PASS | 2 |
| case_adjective_order | pl → en | realization | FAIL | 3 |
| case_adjective_order | pl → en | realization | GAP_I | 7 |
| case_adjective_order | pl → en | realization | GAP_REALIZE | 25 |
| case_adjective_order | pl → en | realization | PASS | 10 |
| composition_boundary | en → pl | analysis | GAP_ANALYZE | 28 |
| composition_boundary | en → pl | analysis | GAP_I | 13 |
| composition_boundary | en → pl | analysis | PASS | 4 |
| composition_boundary | en → pl | integration | FAIL | 4 |
| composition_boundary | en → pl | integration | GAP_ANALYZE | 28 |
| composition_boundary | en → pl | integration | GAP_I | 13 |
| composition_boundary | en → pl | realization | GAP_I | 13 |
| composition_boundary | en → pl | realization | GAP_REALIZE | 32 |
| composition_boundary | pl → en | analysis | GAP_ANALYZE | 29 |
| composition_boundary | pl → en | analysis | GAP_I | 13 |
| composition_boundary | pl → en | analysis | PASS | 3 |
| composition_boundary | pl → en | integration | GAP_ANALYZE | 29 |
| composition_boundary | pl → en | integration | GAP_I | 13 |
| composition_boundary | pl → en | integration | PASS | 3 |
| composition_boundary | pl → en | realization | FAIL | 1 |
| composition_boundary | pl → en | realization | GAP_I | 13 |
| composition_boundary | pl → en | realization | GAP_REALIZE | 27 |
| composition_boundary | pl → en | realization | PASS | 4 |
| identification | en → pl | analysis | FAIL | 3 |
| identification | en → pl | analysis | GAP_ANALYZE | 17 |
| identification | en → pl | analysis | GAP_I | 4 |
| identification | en → pl | analysis | PASS | 16 |
| identification | en → pl | integration | FAIL | 3 |
| identification | en → pl | integration | GAP_ANALYZE | 17 |
| identification | en → pl | integration | GAP_I | 4 |
| identification | en → pl | integration | GAP_REALIZE | 1 |
| identification | en → pl | integration | PASS | 15 |
| identification | en → pl | realization | GAP_I | 4 |
| identification | en → pl | realization | GAP_REALIZE | 16 |
| identification | en → pl | realization | PASS | 20 |
| identification | pl → en | analysis | FAIL | 3 |
| identification | pl → en | analysis | GAP_ANALYZE | 12 |
| identification | pl → en | analysis | GAP_I | 4 |
| identification | pl → en | analysis | PASS | 21 |
| identification | pl → en | integration | FAIL | 3 |
| identification | pl → en | integration | GAP_ANALYZE | 12 |
| identification | pl → en | integration | GAP_I | 4 |
| identification | pl → en | integration | GAP_REALIZE | 5 |
| identification | pl → en | integration | PASS | 16 |
| identification | pl → en | realization | GAP_I | 4 |
| identification | pl → en | realization | GAP_REALIZE | 16 |
| identification | pl → en | realization | PASS | 20 |
| location_quantity | en → pl | analysis | GAP_I | 50 |
| location_quantity | en → pl | integration | GAP_I | 50 |
| location_quantity | en → pl | realization | GAP_I | 50 |
| location_quantity | pl → en | analysis | GAP_I | 50 |
| location_quantity | pl → en | integration | GAP_I | 50 |
| location_quantity | pl → en | realization | GAP_I | 50 |
| negation | en → pl | analysis | FAIL | 2 |
| negation | en → pl | analysis | GAP_ANALYZE | 33 |
| negation | en → pl | analysis | GAP_I | 9 |
| negation | en → pl | analysis | PASS | 6 |
| negation | en → pl | integration | FAIL | 2 |
| negation | en → pl | integration | GAP_ANALYZE | 33 |
| negation | en → pl | integration | GAP_I | 9 |
| negation | en → pl | integration | PASS | 6 |
| negation | en → pl | realization | GAP_I | 9 |
| negation | en → pl | realization | GAP_REALIZE | 32 |
| negation | en → pl | realization | PASS | 9 |
| negation | pl → en | analysis | FAIL | 2 |
| negation | pl → en | analysis | GAP_ANALYZE | 31 |
| negation | pl → en | analysis | GAP_I | 9 |
| negation | pl → en | analysis | PASS | 8 |
| negation | pl → en | integration | FAIL | 1 |
| negation | pl → en | integration | GAP_ANALYZE | 31 |
| negation | pl → en | integration | GAP_I | 9 |
| negation | pl → en | integration | PASS | 9 |
| negation | pl → en | realization | GAP_I | 9 |
| negation | pl → en | realization | GAP_REALIZE | 32 |
| negation | pl → en | realization | PASS | 9 |
| possession | en → pl | analysis | GAP_I | 40 |
| possession | en → pl | integration | GAP_I | 40 |
| possession | en → pl | realization | GAP_I | 40 |
| possession | pl → en | analysis | GAP_I | 40 |
| possession | pl → en | integration | GAP_I | 40 |
| possession | pl → en | realization | GAP_I | 40 |
| questions | en → pl | analysis | GAP_ANALYZE | 18 |
| questions | en → pl | analysis | GAP_I | 42 |
| questions | en → pl | integration | GAP_ANALYZE | 18 |
| questions | en → pl | integration | GAP_I | 42 |
| questions | en → pl | realization | FAIL | 1 |
| questions | en → pl | realization | GAP_I | 42 |
| questions | en → pl | realization | GAP_REALIZE | 17 |
| questions | pl → en | analysis | GAP_ANALYZE | 18 |
| questions | pl → en | analysis | GAP_I | 42 |
| questions | pl → en | integration | GAP_ANALYZE | 18 |
| questions | pl → en | integration | GAP_I | 42 |
| questions | pl → en | realization | FAIL | 1 |
| questions | pl → en | realization | GAP_I | 42 |
| questions | pl → en | realization | GAP_REALIZE | 17 |
| time_agreement | en → pl | analysis | FAIL | 5 |
| time_agreement | en → pl | analysis | GAP_ANALYZE | 30 |
| time_agreement | en → pl | analysis | GAP_I | 8 |
| time_agreement | en → pl | analysis | PASS | 17 |
| time_agreement | en → pl | integration | FAIL | 10 |
| time_agreement | en → pl | integration | GAP_ANALYZE | 30 |
| time_agreement | en → pl | integration | GAP_I | 8 |
| time_agreement | en → pl | integration | PASS | 12 |
| time_agreement | en → pl | realization | GAP_I | 8 |
| time_agreement | en → pl | realization | GAP_REALIZE | 35 |
| time_agreement | en → pl | realization | PASS | 17 |
| time_agreement | pl → en | analysis | FAIL | 3 |
| time_agreement | pl → en | analysis | GAP_ANALYZE | 30 |
| time_agreement | pl → en | analysis | GAP_I | 8 |
| time_agreement | pl → en | analysis | PASS | 19 |
| time_agreement | pl → en | integration | FAIL | 7 |
| time_agreement | pl → en | integration | GAP_ANALYZE | 30 |
| time_agreement | pl → en | integration | GAP_I | 8 |
| time_agreement | pl → en | integration | GAP_REALIZE | 1 |
| time_agreement | pl → en | integration | PASS | 14 |
| time_agreement | pl → en | realization | FAIL | 4 |
| time_agreement | pl → en | realization | GAP_I | 8 |
| time_agreement | pl → en | realization | GAP_REALIZE | 31 |
| time_agreement | pl → en | realization | PASS | 17 |

## Inventory and pack metadata

- Inventory records parsed: 1000 / 1000 declared directions.
- Snapshot identity records: 1.
- Sense catalogue associations parsed: 2031 (not morphology execution coverage).
- Declared Meaning ID references: 1094 allowed; 1094 required.
- Inventory fields preserve declared tags, source semantics, allowed/required Meaning IDs, and owner status separately from runtime observations.
- Inventory source SHA `386f0d98ca12bde7e53880109db4f634958321c6`: 1000 directions.

### Declared source requirements

Aggregates below summarize authored source requirements; they are not observed analyzer output.

| Required predicate | Directions |
|---|---:|
| `(unspecified)` | 456 |
| `accompaniment` | 2 |
| `belonging_association` | 4 |
| `between_position` | 4 |
| `black` | 4 |
| `blue` | 2 |
| `class_membership` | 2 |
| `closing_object` | 2 |
| `conversation` | 2 |
| `crying_emotionally` | 2 |
| `cutting_object` | 2 |
| `drinking` | 12 |
| `drinking+eating` | 2 |
| `drinking+reading` | 2 |
| `eating` | 12 |
| `finite_present_existence_with_exact_quantity` | 50 |
| `friendship_relationship` | 2 |
| `front_back_order` | 8 |
| `giving` | 24 |
| `green` | 2 |
| `have_unspecified` | 54 |
| `hearing` | 40 |
| `helping` | 6 |
| `interior_location` | 10 |
| `legal_ownership` | 4 |
| `located_in` | 6 |
| `looking` | 2 |
| `negative_absence_or_have` | 1 |
| `negative_existence` | 1 |
| `on_support` | 12 |
| `open_state` | 6 |
| `opening_object` | 4 |
| `opening_object+closing_object` | 2 |
| `origin_from` | 2 |
| `possessive_association` | 18 |
| `proximity_relation` | 4 |
| `reading` | 28 |
| `reading+drinking` | 2 |
| `reading+hearing` | 2 |
| `reading+looking` | 2 |
| `reading+seeing` | 2 |
| `red` | 2 |
| `reflexive_possessive_association` | 4 |
| `running` | 2 |
| `sadness` | 6 |
| `seeing` | 102 |
| `seeing+hearing` | 20 |
| `siblinghood` | 4 |
| `sitting_change` | 2 |
| `sleeping` | 16 |
| `sleeping+reading` | 2 |
| `sleeping+running` | 4 |
| `spatial_adjacency` | 4 |
| `speaking+sleeping` | 2 |
| `thanking` | 2 |
| `under` | 4 |
| `using_object` | 2 |
| `vertical_order` | 6 |
| `walking` | 2 |
| `white` | 2 |
| `writing_content` | 4 |
| `writing_instrument` | 2 |
| `writing_message` | 2 |

Declared role mentions: `agent` 16, `asset` 4, `associate` 4, `associated` 22, `author` 4, `back` 8, `bearer` 24, `beneficiary` 6, `choice_restriction` 2, `companion` 2, `content` 2, `destination` 2, `document` 2, `drinker` 16, `eater` 12, `entity` 2, `existence` 50, `experiencer` 160, `figure` 4, `food` 10, `friend_a` 2, `friend_b` 2, `front` 8, `giver` 20, `helper` 6, `higher` 6, `holder` 50, `instrument` 4, `item` 4, `landmark` 32, `landmark_a` 4, `landmark_b` 4, `liquid` 12, `located` 34, `location` 10, `lower` 6, `manner` 2, `material` 38, `member` 2, `object` 56, `occupant` 6, `owner` 4, `partner` 2, `person_variable` 6, `possessor` 18, `quantified_set` 50, `reader` 38, `reason` 2, `recipient` 24, `runner` 4, `sibling_a` 4, `sibling_b` 4, `sleeper` 22, `source` 2, `speaker` 4, `stimulus` 160, `support` 2, `supported` 2, `target` 4, `thanker` 2, `theme` 20, `time` 4, `tool` 2, `user` 2, `walker` 2, `writer` 2.
- Pack identity: `{"dataset_id": "golingua-authored", "dataset_version": "2026-10-23", "digest": "e62fdf8903c59fd27690aa46e58b00d199a10a0e1108869e3aa8d47a8f7196f0", "format_version": 1}`; inventory source SHA: `386f0d98ca12bde7e53880109db4f634958321c6`.

Catalogue sense associations are not proof that a paradigm form or morphology procedure ran. Aggregates below describe declared pack associations only.

| Association | Language | POS | Paradigm | Senses |
|---|---|---|---|---:|
| associated | en | adjective | en_identification_black_positive | 1 |
| associated | en | determiner | en_determiner_definite | 1 |
| associated | en | determiner | en_determiner_indefinite | 2 |
| associated | en | noun | en_noun_consonant_y | 19 |
| associated | en | noun | en_noun_principal_parts | 9 |
| associated | en | noun | en_noun_regular | 52 |
| associated | en | noun | en_noun_sibilant | 13 |
| associated | en | noun | en_noun_singular_attested | 3 |
| associated | en | noun | en_noun_zero_plural | 3 |
| associated | en | noun | not-declared | 613 |
| associated | en | preposition | not-declared | 23 |
| associated | en | pronoun | en_personal_pronoun_i | 1 |
| associated | en | pronoun | en_personal_pronoun_we | 1 |
| associated | en | pronoun | en_personal_pronoun_you | 1 |
| associated | en | verb | en_identification_be_present_3sg | 1 |
| associated | en | verb | en_verb_finite_consonant_y | 1 |
| associated | en | verb | en_verb_finite_principal_parts | 6 |
| associated | en | verb | en_verb_finite_regular | 12 |
| associated | en | verb | en_verb_finite_sibilant | 5 |
| associated | en | verb | en_verb_finite_silent_e | 13 |
| associated | en | verb | not-declared | 238 |
| associated | pl | adjective | pl_adjective_dorosly_attested | 1 |
| associated | pl | adjective | pl_identification_czarny_masculine_animal_nominative | 1 |
| associated | pl | foreign | not-declared | 1 |
| associated | pl | noun | not-declared | 690 |
| associated | pl | noun | pl_noun_attested_feminine | 8 |
| associated | pl | noun | pl_noun_attested_masculine_animal | 1 |
| associated | pl | noun | pl_noun_attested_masculine_inanimate | 7 |
| associated | pl | noun | pl_noun_attested_masculine_personal | 5 |
| associated | pl | noun | pl_noun_attested_neuter | 6 |
| associated | pl | noun | pl_noun_attested_pl_kon_n | 1 |
| associated | pl | noun | pl_noun_attested_pl_kot_n | 1 |
| associated | pl | noun | pl_noun_attested_pl_miasto_n | 1 |
| associated | pl | noun | pl_noun_attested_pl_mieszkanie_n | 1 |
| associated | pl | noun | pl_noun_attested_pl_ptak_n | 1 |
| associated | pl | noun | pl_noun_attested_pl_rzeka_n | 1 |
| associated | pl | noun | pl_noun_attested_pl_siostra_n | 2 |
| associated | pl | noun | pl_noun_drzewo_attested | 1 |
| associated | pl | noun | pl_noun_feminine_a_partial | 4 |
| associated | pl | noun | pl_noun_plural_only_attested | 1 |
| associated | pl | noun_proper | not-declared | 1 |
| associated | pl | pronoun | pl_personal_pl.ja.pron | 1 |
| associated | pl | pronoun | pl_personal_pl.my.pron | 1 |
| associated | pl | pronoun | pl_personal_pl.ty.pron | 1 |
| associated | pl | pronoun | pl_personal_pl.wy.pron | 1 |
| associated | pl | verb | not-declared | 237 |
| associated | pl | verb | pl_identification_byc_present_3sg | 1 |
| associated | pl | verb | pl_verb_czytac_finite | 1 |
| associated | pl | verb | pl_verb_perfective_finite | 2 |
| associated | pl | verb | pl_verb_present_i | 1 |
| associated | pl | verb | pl_verb_present_y | 2 |
| associated | pl | verb_reflexive | not-declared | 30 |

## Runtime adapter observations

Runtime calls are separate from check outcomes. A false adapter marker is runtime NOT RUN even when its corresponding GAPI check has a recorded result. Check NOT RUN remains in the record table above.

| Stage | Adapter executed | Calls |
|---|---|---:|
| context_source_analysis | true | 152 |
| independent_realization | false | 362 |
| independent_realization | true | 762 |
| integration | true | 1000 |
| native_analysis | true | 1012 |
| source_gap_probe | true | 348 |

### Independent realization runtime NOT RUN

- Calls not executed: 362

| Reason | Affected record IDs | Unexecuted calls |
|---|---|---:|
| A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | `A1.9-Q-1009-EN-PL:realize`, `A1.9-Q-1010-PL-EN:realize`, `A1.9-Q-1011-EN-PL:realize`, `A1.9-Q-1012-PL-EN:realize`, `A1.9-Q-921-EN-PL:realize`, `A1.9-Q-922-PL-EN:realize`, `A1.9-Q-923-EN-PL:realize`, `A1.9-Q-924-PL-EN:realize`, `A1.9-Q-925-EN-PL:realize`, `A1.9-Q-926-PL-EN:realize`, `A1.9-Q-927-EN-PL:realize`, `A1.9-Q-928-PL-EN:realize`, `A1.9-Q-929-EN-PL:realize`, `A1.9-Q-930-PL-EN:realize`, `A1.9-Q-931-EN-PL:realize`, `A1.9-Q-932-PL-EN:realize`, `A1.9-Q-933-EN-PL:realize`, `A1.9-Q-934-PL-EN:realize`, `A1.9-Q-935-EN-PL:realize`, `A1.9-Q-936-PL-EN:realize`, `A1.9-Q-937-EN-PL:realize`, `A1.9-Q-938-PL-EN:realize`, `A1.9-Q-939-EN-PL:realize`, `A1.9-Q-940-PL-EN:realize`, `A1.9-Q-941-EN-PL:realize`, `A1.9-Q-942-PL-EN:realize`, `A1.9-Q-943-EN-PL:realize`, `A1.9-Q-944-PL-EN:realize`, `A1.9-Q-945-EN-PL:realize`, `A1.9-Q-946-PL-EN:realize`, `A1.9-Q-949-EN-PL:realize`, `A1.9-Q-950-PL-EN:realize`, `A1.9-Q-975-EN-PL:realize`, `A1.9-Q-976-PL-EN:realize`, `A1.9-Q-977-EN-PL:realize`, `A1.9-Q-978-PL-EN:realize`, `A1.9-Q-983-EN-PL:realize`, `A1.9-Q-984-PL-EN:realize`, `A1.9-Q-985-EN-PL:realize`, `A1.9-Q-986-PL-EN:realize`, `A1.9-Q-987-EN-PL:realize`, `A1.9-Q-988-PL-EN:realize` | 42 |
| CORE cutting_object omits the instrument role; knife_utensil is only a referent concept. | `A1.11-CASE-1175-EN-PL:realize`, `A1.11-CASE-1176-PL-EN:realize` | 2 |
| CORE existence has no participant role to bind the quantified cats, and TemporalFrame is forbidden on a quantifier or referent. The supported exactly-two-cats component alone does not assert finite present existence. | `A1-LQ-003-EN-PL:realize`, `A1-LQ-004-PL-EN:realize` | 2 |
| CORE friendship_relationship is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles friend_a, friend_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. | `A1.6-PO-667-EN-PL:realize` | 1 |
| CORE friendship_relationship is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles friend_a, friend_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. Polish friend noun morphology does not establish the friend’s personal gender. | `A1.6-PO-668-PL-EN:realize` | 1 |
| CORE has black but no blue property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | `A1.5-ID-573-EN-PL:realize`, `A1.5-ID-574-PL-EN:realize` | 2 |
| CORE has black but no green property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | `A1.5-ID-575-EN-PL:realize`, `A1.5-ID-576-PL-EN:realize` | 2 |
| CORE has black but no red property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | `A1.5-ID-569-EN-PL:realize`, `A1.5-ID-570-PL-EN:realize` | 2 |
| CORE has black but no white property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | `A1.5-ID-571-EN-PL:realize`, `A1.5-ID-572-PL-EN:realize` | 2 |
| CORE has no generic BELONG TO association predicate with associate/item roles. It is not equivalent to part-whole, temporary possession, or legal ownership; candidate possession concepts are KindRelation and cannot express the complete sentence as Meaning. No partial graph is supplied. | `A1.6-PO-633-EN-PL:realize`, `A1.6-PO-634-PL-EN:realize`, `A1.6-PO-635-EN-PL:realize`, `A1.6-PO-636-PL-EN:realize` | 4 |
| CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. | `A1.6-PO-615-EN-PL:realize`, `A1.6-PO-616-PL-EN:realize` | 2 |
| CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | `A1.6-PO-617-EN-PL:realize`, `A1.6-PO-618-PL-EN:realize`, `A1.6-PO-619-EN-PL:realize`, `A1.6-PO-620-PL-EN:realize` | 4 |
| CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. | `A1.6-PO-601-EN-PL:realize`, `A1.6-PO-602-PL-EN:realize`, `A1.6-PO-603-EN-PL:realize`, `A1.6-PO-604-PL-EN:realize`, `A1.6-PO-605-EN-PL:realize`, `A1.6-PO-606-PL-EN:realize`, `A1.6-PO-607-EN-PL:realize`, `A1.6-PO-608-PL-EN:realize`, `A1.6-PO-621-EN-PL:realize`, `A1.6-PO-622-PL-EN:realize`, `A1.6-PO-625-EN-PL:realize`, `A1.6-PO-626-PL-EN:realize`, `A1.6-PO-627-EN-PL:realize`, `A1.6-PO-628-PL-EN:realize`, `A1.6-PO-629-EN-PL:realize`, `A1.6-PO-630-PL-EN:realize`, `A1.6-PO-631-EN-PL:realize`, `A1.6-PO-632-PL-EN:realize` | 18 |
| CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | `A1.6-PO-609-EN-PL:realize`, `A1.6-PO-610-PL-EN:realize`, `A1.6-PO-611-EN-PL:realize`, `A1.6-PO-612-PL-EN:realize`, `A1.6-PO-613-EN-PL:realize`, `A1.6-PO-614-PL-EN:realize`, `A1.6-PO-623-EN-PL:realize`, `A1.6-PO-624-PL-EN:realize` | 8 |
| CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. | `A1.6-PO-641-EN-PL:realize`, `A1.6-PO-642-PL-EN:realize`, `A1.6-PO-643-EN-PL:realize`, `A1.6-PO-644-PL-EN:realize`, `A1.6-PO-645-EN-PL:realize`, `A1.6-PO-647-EN-PL:realize`, `A1.6-PO-648-PL-EN:realize`, `A1.6-PO-649-EN-PL:realize`, `A1.6-PO-650-PL-EN:realize`, `A1.6-PO-651-EN-PL:realize`, `A1.6-PO-652-PL-EN:realize`, `A1.6-PO-653-EN-PL:realize`, `A1.6-PO-654-PL-EN:realize`, `A1.6-PO-655-EN-PL:realize`, `A1.6-PO-656-PL-EN:realize` | 15 |
| CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | `A1.6-PO-657-EN-PL:realize`, `A1.6-PO-658-PL-EN:realize` | 2 |
| CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. Polish jego permits masculine or neuter reference. | `A1.6-PO-646-PL-EN:realize` | 1 |
| CORE has no generic reflexive-possessive association predicate with holder/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for reflexive association. No partial graph is supplied. | `A1.6-PO-659-EN-PL:realize`, `A1.6-PO-660-PL-EN:realize`, `A1.6-PO-661-EN-PL:realize`, `A1.6-PO-662-PL-EN:realize` | 4 |
| CORE has no posture-change event for moving into sitting; the existing sitting posture state is not equivalent. | `A1.13-COMP-1353-EN-PL:realize`, `A1.13-COMP-1354-PL-EN:realize` | 2 |
| CORE has no using_object predicate with user/tool roles; the tool concept alone cannot encode the action. | `A1.11-CASE-1173-EN-PL:realize`, `A1.11-CASE-1174-PL-EN:realize` | 2 |
| CORE has opening_object EVENT; open-state is a distinct untimed state and is not encoded. | `A1.11-CASE-1127-EN-PL:realize`, `A1.11-CASE-1128-PL-EN:realize` | 2 |
| CORE has supports_physical, but no on-support assertion with figure/landmark roles in Meaning. | `A1.11-CASE-1103-EN-PL:realize`, `A1.11-CASE-1104-PL-EN:realize` | 2 |
| CORE legal_ownership is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles owner, asset; no partial graph is supplied. The wording explicitly licenses legal ownership, but the relation kind remains unavailable as a Meaning predicate. | `A1.6-PO-637-EN-PL:realize`, `A1.6-PO-638-PL-EN:realize`, `A1.6-PO-639-EN-PL:realize`, `A1.6-PO-640-PL-EN:realize` | 4 |
| CORE siblinghood is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles sibling_a, sibling_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. | `A1.6-PO-663-EN-PL:realize`, `A1.6-PO-664-PL-EN:realize`, `A1.6-PO-665-EN-PL:realize`, `A1.6-PO-666-PL-EN:realize` | 4 |
| CORE speaking has no content role to bind the embedded sleeping proposition; adding a bare sleep root would assert it. | `A1.13-COMP-1359-EN-PL:realize`, `A1.13-COMP-1360-PL-EN:realize` | 2 |
| Full negative writing-content proposition requires content binding on writing_message; writing_document would assert a durable document instead. | `A1.8-NEG-809-EN-PL:realize`, `A1.8-NEG-810-PL-EN:realize` | 2 |
| Generic HAVE does not distinguish temporary possession from legal ownership and implies neither legal title nor duration; CORE temporary_possession is KindRelation, not a Meaning State/Event predicate, and neutral Meaning has no holder/object possession roles. | `A1-PO-001-EN-PL:realize`, `A1-PO-002-PL-EN:realize`, `A1-PO-003-EN-PL:realize`, `A1-PO-004-PL-EN:realize`, `A1.2-PO-201-PL-EN:realize`, `A1.2-PO-202-EN-PL:realize`, `A1.2-PO-203-PL-EN:realize`, `A1.2-PO-204-EN-PL:realize`, `A1.2-PO-205-PL-EN:realize`, `A1.2-PO-206-EN-PL:realize`, `A1.2-PO-207-PL-EN:realize`, `A1.2-PO-208-EN-PL:realize` | 12 |
| Meaning cannot bind the child specifically to the stated mother; child_of_person alone does not encode which parent. | `A1.12-AMB-1285-EN-PL:realize`, `A1.12-AMB-1286-PL-EN:realize` | 2 |
| Meaning cannot bind “on the table” specifically to cat versus seeing event/observer without a location/attachment relation. | `A1.12-AMB-1277-EN-PL:realize`, `A1.12-AMB-1278-PL-EN:realize` | 2 |
| Meaning cannot bind “with a book” specifically to teacher versus observer without asserting ownership. | `A1.12-AMB-1279-EN-PL:realize`, `A1.12-AMB-1280-PL-EN:realize` | 2 |
| Meaning cannot bind “with the child” specifically to mother versus observer without an accompaniment/attachment relation. | `A1.12-AMB-1275-EN-PL:realize`, `A1.12-AMB-1276-PL-EN:realize` | 2 |
| Meaning cannot represent finite negative absence together with the source-language HAVE/pro-drop alternative without conflation. | `A1.12-AMB-1274-PL-EN:realize` | 1 |
| Meaning has no explicit quantifier-variable binding/scope for a shared teacher with every dog. | `A1.12-AMB-1267-EN-PL:realize`, `A1.12-AMB-1268-PL-EN:realize` | 2 |
| Meaning has no explicit quantifier-variable binding/scope for every teacher to a dog that may vary per teacher. | `A1.12-AMB-1265-EN-PL:realize`, `A1.12-AMB-1266-PL-EN:realize` | 2 |
| Meaning lacks a full present finite absence predicate tying existence and its discourse domain to zero cat referents; the English source licenses absence only. | `A1.12-AMB-1273-EN-PL:realize` | 1 |
| Meaning v1 cannot preserve both epistemic possibility and permission readings (and the Polish ability reading) without asserting actual sleep. | `A1.13-COMP-1357-EN-PL:realize`, `A1.13-COMP-1358-PL-EN:realize` | 2 |
| Meaning v1 has no causal relation binding the two propositions. | `A1.13-COMP-1331-EN-PL:realize`, `A1.13-COMP-1332-PL-EN:realize` | 2 |
| Meaning v1 has no contrast relation; a plain conjunction erases the selected contrast. | `A1.13-COMP-1363-EN-PL:realize`, `A1.13-COMP-1364-PL-EN:realize` | 2 |
| Meaning v1 has no disjunction operator for alternative propositions; a root conjunction would assert the wrong relation. | `A1.13-COMP-1327-EN-PL:realize`, `A1.13-COMP-1328-PL-EN:realize` | 2 |
| Meaning v1 has no heterogeneous coordinated-group representation that preserves the selected collective/distributive ambiguity; separate atomic predicates would not satisfy the group reading. | `A1.13-COMP-1323-EN-PL:realize`, `A1.13-COMP-1324-PL-EN:realize` | 2 |
| Meaning v1 has no ordered temporal relation between events; discourse order is not event sequence. | `A1.13-COMP-1333-EN-PL:realize`, `A1.13-COMP-1334-PL-EN:realize` | 2 |
| Meaning v1 hypothetical scope cannot encode the selected conditional implication between two propositions. | `A1.13-COMP-1329-EN-PL:realize`, `A1.13-COMP-1330-PL-EN:realize` | 2 |
| Meaning v1 lacks a heterogeneous group and unresolved collective/distributive event representation; separate atomic predicates would not satisfy the selected subject-group reading. | `A1.13-COMP-1325-EN-PL:realize`, `A1.13-COMP-1326-PL-EN:realize` | 2 |
| Meaning v1 request act does not preserve explicit please/politeness; neither a bare request nor one child sense represents the selected politeness and child-sense ambiguity. | `A1.13-COMP-1349-EN-PL:realize`, `A1.13-COMP-1350-PL-EN:realize` | 2 |
| Meaning v1 request act does not preserve explicit please/politeness; request force alone does not meet the selected full meaning. | `A1.13-COMP-1343-EN-PL:realize`, `A1.13-COMP-1344-PL-EN:realize` | 2 |
| Meaning v1 temporal_frame records reference time per event but cannot encode overlap between events. | `A1.13-COMP-1335-EN-PL:realize`, `A1.13-COMP-1336-PL-EN:realize` | 2 |
| Negation over the existentially bound cat participant is required: the actor must be quantified, while negating one event does not quantify its drinker. Meaning has no typed bound-variable link from a quantifier to that event role. | `A1.8-NEG-887-EN-PL:realize`, `A1.8-NEG-888-PL-EN:realize` | 2 |
| No accompaniment relation/predicate is licensed; another social or ownership assertion would change meaning. | `A1.11-CASE-1181-EN-PL:realize`, `A1.11-CASE-1182-PL-EN:realize` | 4 |
| No reciprocal conversation predicate is licensed; speaking would assert a different one-way event. | `A1.11-CASE-1179-EN-PL:realize`, `A1.11-CASE-1180-PL-EN:realize` | 2 |
| No writing-with-instrument predicate/role is licensed. writing_document would add durable-document meaning. | `A1.11-CASE-1177-EN-PL:realize`, `A1.11-CASE-1178-PL-EN:realize` | 2 |
| Polish drzwi is plural-only but does not establish numeric multiplicity; preserve unknown count. The full open-state predicate with a bearer role is absent from CORE. | `A1.9-Q-1002-PL-EN:realize` | 1 |
| The English source explicitly presents one door; the required open-state predicate with a bearer role is absent from CORE. | `A1.9-Q-1001-EN-PL:realize` | 1 |
| The full location proposition requires a located-in relation plus an explicit unknown location variable; CORE located_in is KindRelation, not a state predicate. | `A1.9-Q-953-EN-PL:realize`, `A1.9-Q-954-PL-EN:realize`, `A1.9-Q-955-EN-PL:realize`, `A1.9-Q-956-PL-EN:realize` | 4 |
| The full possession proposition is required; available CORE possession predicates encode a specific relation and cannot express unspecified control or possession without that inference. | `A1.9-Q-979-EN-PL:realize`, `A1.9-Q-980-PL-EN:realize`, `A1.9-Q-981-EN-PL:realize`, `A1.9-Q-982-PL-EN:realize` | 4 |
| The full sentence needs a present-time existence assertion linked to an exactly-1 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→book(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-795-EN-PL:realize`, `A1.7-LQ-796-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-1 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→cat(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-749-EN-PL:realize`, `A1.7-LQ-750-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-1 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→window(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-783-EN-PL:realize`, `A1.7-LQ-784-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-10 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(10)→cat(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-757-EN-PL:realize`, `A1.7-LQ-758-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→apple_fruit(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-787-EN-PL:realize`, `A1.7-LQ-788-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→bird(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-771-EN-PL:realize`, `A1.7-LQ-772-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→book(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-763-EN-PL:realize`, `A1.7-LQ-764-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→car(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-767-EN-PL:realize`, `A1.7-LQ-768-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→dog(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-759-EN-PL:realize`, `A1.7-LQ-760-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted door set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→door(many) component alone asserts cardinality, not finite present existence. The collective Polish numeral dwoje takes singular finite agreement; the counted units are doors, not pairs of doors. | `A1.7-LQ-791-EN-PL:realize`, `A1.7-LQ-792-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→teacher(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-775-EN-PL:realize`, `A1.7-LQ-776-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-2 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology. | `A1.7-LQ-779-EN-PL:realize`, `A1.7-LQ-780-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-3 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(3)→cat(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-751-EN-PL:realize`, `A1.7-LQ-752-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-4 counted bag set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(4)→bag(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-793-EN-PL:realize`, `A1.7-LQ-794-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-4 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(4)→cat(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-753-EN-PL:realize`, `A1.7-LQ-754-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→apple_fruit(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-789-EN-PL:realize`, `A1.7-LQ-790-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→bird(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-773-EN-PL:realize`, `A1.7-LQ-774-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→book(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-765-EN-PL:realize`, `A1.7-LQ-766-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→car(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-769-EN-PL:realize`, `A1.7-LQ-770-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→cat(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-755-EN-PL:realize`, `A1.7-LQ-756-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→dog(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-761-EN-PL:realize`, `A1.7-LQ-762-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→teacher(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-777-EN-PL:realize`, `A1.7-LQ-778-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→window(many) component alone asserts cardinality, not finite present existence. | `A1.7-LQ-785-EN-PL:realize`, `A1.7-LQ-786-PL-EN:realize` | 2 |
| The full sentence needs a present-time existence assertion linked to an exactly-5 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology. | `A1.7-LQ-781-EN-PL:realize`, `A1.7-LQ-782-PL-EN:realize` | 2 |
| The full sentence requires a home-location relation scoped under not-every; CORE location relation is KindRelation and cannot be a Meaning predicate. | `A1.8-NEG-885-EN-PL:realize`, `A1.8-NEG-886-PL-EN:realize` | 2 |
| The full sentence requires the present-time spatial relation between_position with roles located, landmark_a, landmark_b; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve both distinct landmark anchors; their order is not asserted. | `A1.7-LQ-729-EN-PL:realize`, `A1.7-LQ-730-PL-EN:realize`, `A1.7-LQ-731-EN-PL:realize`, `A1.7-LQ-732-PL-EN:realize` | 4 |
| The full sentence requires the present-time spatial relation front_back_order with roles front, back; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Front/back viewpoint remains unresolved; do not invent an observer position. | `A1.7-LQ-717-EN-PL:realize`, `A1.7-LQ-718-PL-EN:realize`, `A1.7-LQ-719-EN-PL:realize`, `A1.7-LQ-720-PL-EN:realize`, `A1.7-LQ-733-EN-PL:realize`, `A1.7-LQ-734-PL-EN:realize`, `A1.7-LQ-735-EN-PL:realize`, `A1.7-LQ-736-PL-EN:realize` | 8 |
| The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. | `A1.7-LQ-707-EN-PL:realize`, `A1.7-LQ-708-PL-EN:realize`, `A1.7-LQ-709-EN-PL:realize`, `A1.7-LQ-710-PL-EN:realize`, `A1.7-LQ-743-EN-PL:realize`, `A1.7-LQ-744-PL-EN:realize`, `A1.7-LQ-745-EN-PL:realize`, `A1.7-LQ-746-PL-EN:realize` | 8 |
| The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. Preserve grammatical plurality but do not infer an exact count. | `A1.7-LQ-739-EN-PL:realize`, `A1.7-LQ-740-PL-EN:realize` | 2 |
| The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. | `A1.7-LQ-701-EN-PL:realize`, `A1.7-LQ-702-PL-EN:realize`, `A1.7-LQ-705-EN-PL:realize`, `A1.7-LQ-706-PL-EN:realize`, `A1.7-LQ-721-EN-PL:realize`, `A1.7-LQ-722-PL-EN:realize` | 6 |
| The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. Preserve grammatical plurality but do not infer an exact count. | `A1.7-LQ-737-EN-PL:realize`, `A1.7-LQ-738-PL-EN:realize` | 2 |
| The full sentence requires the present-time spatial relation proximity_relation with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Nearness has no numeric distance threshold. | `A1.7-LQ-713-EN-PL:realize`, `A1.7-LQ-714-PL-EN:realize`, `A1.7-LQ-747-EN-PL:realize`, `A1.7-LQ-748-PL-EN:realize` | 4 |
| The full sentence requires the present-time spatial relation spatial_adjacency with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Adjacency does not mean an arbitrary proximity threshold. | `A1.7-LQ-711-EN-PL:realize`, `A1.7-LQ-712-PL-EN:realize`, `A1.7-LQ-723-EN-PL:realize`, `A1.7-LQ-724-PL-EN:realize` | 4 |
| The full sentence requires the present-time spatial relation under with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. | `A1.7-LQ-703-EN-PL:realize`, `A1.7-LQ-704-PL-EN:realize` | 2 |
| The full sentence requires the present-time spatial relation under with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve grammatical plurality but do not infer an exact count. | `A1.7-LQ-741-EN-PL:realize`, `A1.7-LQ-742-PL-EN:realize` | 2 |
| The full sentence requires the present-time spatial relation vertical_order with roles higher, lower; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve relative higher/lower order without inferring contact or posture. | `A1.7-LQ-715-EN-PL:realize`, `A1.7-LQ-716-PL-EN:realize`, `A1.7-LQ-725-EN-PL:realize`, `A1.7-LQ-726-PL-EN:realize`, `A1.7-LQ-727-EN-PL:realize`, `A1.7-LQ-728-PL-EN:realize` | 6 |
| The full sleeping-in-the-house proposition requires located_in(occupant, location); CORE located_in is KindRelation and cannot be a Meaning predicate. | `A1.9-Q-1005-EN-PL:realize`, `A1.9-Q-1006-PL-EN:realize` | 2 |
| The full writing-content event needs a content participant, and the WH writer needs an explicit unknown-variable binding; neither is supplied by a partial writing-message event. | `A1.9-Q-947-EN-PL:realize`, `A1.9-Q-948-PL-EN:realize` | 2 |
| The input explicitly licenses walking; the WH destination remains unrepresented because Meaning has no typed unknown-variable binding. | `A1.9-Q-959-EN-PL:realize`, `A1.9-Q-960-PL-EN:realize` | 2 |
| The negative sadness state requires a bearer role, but CORE sadness has no compatible bearer role; no partial graph substitutes for the complete proposition. | `A1.8-NEG-855-EN-PL:realize`, `A1.8-NEG-856-PL-EN:realize`, `A1.8-NEG-857-EN-PL:realize`, `A1.8-NEG-858-PL-EN:realize`, `A1.8-NEG-859-EN-PL:realize`, `A1.8-NEG-860-PL-EN:realize` | 6 |
| The question requires an explicit unknown time variable bound to the event; a past frame alone does not encode a WH-time query. | `A1.9-Q-971-EN-PL:realize`, `A1.9-Q-972-PL-EN:realize` | 2 |
| The question requires an explicit unknown time variable bound to the event; a present frame alone does not encode a WH-time query. | `A1.9-Q-973-EN-PL:realize`, `A1.9-Q-974-PL-EN:realize` | 2 |
| The question requires an unknown choice variable restricted to book referents and bound to the black bearer role; Question has no such variable. | `A1.9-Q-1013-EN-PL:realize`, `A1.9-Q-1014-PL-EN:realize` | 2 |
| The question requires both a locative participant on the sleeping event and an explicit unknown location variable; the event has no location role. | `A1.9-Q-951-EN-PL:realize`, `A1.9-Q-952-PL-EN:realize` | 2 |
| The requested amount requires a typed quantity variable bound to the liquid role; omission of that argument does not encode a WH amount. | `A1.9-Q-969-EN-PL:realize`, `A1.9-Q-970-PL-EN:realize` | 2 |
| The requested count requires a typed quantity variable bound to the food role; omission of that argument does not encode a WH count. | `A1.9-Q-963-EN-PL:realize`, `A1.9-Q-964-PL-EN:realize` | 2 |
| The requested count requires a typed quantity variable bound to the material role; omission of that argument does not encode a WH count. | `A1.9-Q-965-EN-PL:realize`, `A1.9-Q-966-PL-EN:realize` | 2 |
| The requested count requires a typed quantity variable bound to the stimulus role; omission of that argument does not encode a WH count. | `A1.9-Q-961-EN-PL:realize`, `A1.9-Q-962-PL-EN:realize`, `A1.9-Q-967-EN-PL:realize`, `A1.9-Q-968-PL-EN:realize` | 4 |
| The required located_in predicate is a CORE KindRelation and cannot be represented as a Meaning node; its CORE roles are occupant/location. The unknown location must remain unfilled, never guessed. | `A1-Q-003-EN-PL:realize`, `A1-Q-004-PL-EN:realize` | 2 |
| The sentence requires additive focus for not only in addition to negative sleep; Meaning has no focus operator, so plain negated sleeping would omit required meaning. | `A1.8-NEG-889-EN-PL:realize`, `A1.8-NEG-890-PL-EN:realize` | 2 |
| The sentence requires explicit negative scope over a frequency restriction (never / not always); no frequency operator is represented in Meaning. | `A1.8-NEG-881-EN-PL:realize`, `A1.8-NEG-882-PL-EN:realize`, `A1.8-NEG-883-EN-PL:realize`, `A1.8-NEG-884-PL-EN:realize` | 4 |
| The source asserts a future event or state, but Meaning.TemporalFrame represents only past or present; the future frame value/operator cannot be represented. | `A1.10-TIME-1097-EN-PL:realize`, `A1.10-TIME-1098-PL-EN:realize`, `A1.10-TIME-1099-EN-PL:realize`, `A1.10-TIME-1100-PL-EN:realize`, `A1.10-TIME-1101-EN-PL:realize`, `A1.10-TIME-1102-PL-EN:realize`, `A1.10-TIME-1103-EN-PL:realize`, `A1.10-TIME-1104-PL-EN:realize` | 8 |
| The source asserts an ongoing progressive event at now; Meaning has no progressive/aspect operator or event-in-progress relation, and temporal_frame cannot encode progressive aspect. | `A1.10-TIME-1105-EN-PL:realize`, `A1.10-TIME-1106-PL-EN:realize` | 2 |
| The source asserts habitual/frequency over repeated events (daily/often); Meaning has no frequency operator or temporal quantification over the event. | `A1.10-TIME-1107-EN-PL:realize`, `A1.10-TIME-1108-PL-EN:realize`, `A1.10-TIME-1111-EN-PL:realize`, `A1.10-TIME-1112-PL-EN:realize` | 4 |
| The source asserts two past events with temporal overlap in a when-clause; Meaning has no temporal relation/binding between event nodes, so event overlap and the linked reading cannot be represented. | `A1.10-TIME-1109-EN-PL:realize`, `A1.10-TIME-1110-PL-EN:realize` | 2 |
| The source explicitly refers to multiple windows; the full open-state predicate with a bearer role is absent from CORE. | `A1.9-Q-1003-EN-PL:realize` | 1 |
| The source refers to multiple windows; the full open-state predicate with a bearer role is absent from CORE. | `A1.9-Q-1004-PL-EN:realize` | 1 |
| The source relation is KindRelation and cannot be a Meaning predicate; the question also requires an explicit unknown source variable. | `A1.9-Q-957-EN-PL:realize`, `A1.9-Q-958-PL-EN:realize` | 2 |
| on_support means direct support/contact, not generic located_in. CORE has no on_support predicate; related spatial_contact/supports_physical concepts are KindRelation and cannot be Meaning nodes. The required supported/support roles therefore cannot be represented. | `A1-LQ-001-EN-PL:realize`, `A1-LQ-002-PL-EN:realize` | 2 |
| “Both dogs” requires a quantifier bound to the sleeper participant; a quantifier over the event does not encode that role binding. | `A1.9-Q-1015-EN-PL:realize` | 1 |
| “Oba psy” requires an exactly-two quantifier bound to the sleeper participant; a quantifier over the event does not encode that role binding. | `A1.9-Q-1016-PL-EN:realize` | 1 |

### Separate observed source probes

- context_source_analysis, adapter executed `true`: 152
- source_gap_probe, adapter executed `true`: 348

## Direct-result JSON observations

- Explicit `result=` objects decoded with `JSONDecoder.raw_decode`: 173. They remain separate from adapter-call counts.
- Parse issues: 0

## Observed result outcomes

Observed top-level outcomes by runtime stage; these do not replace check statuses.

| Stage | Result outcome | Count |
|---|---|---:|
| analysis | unknown | 4 |
| analysis | unsupported | 12 |
| context_source_analysis | ambiguous | 6 |
| context_source_analysis | incomplete | 11 |
| context_source_analysis | ok | 54 |
| context_source_analysis | unknown | 7 |
| context_source_analysis | unsupported | 74 |
| direct_result |  | 1 |
| direct_result | ambiguous | 29 |
| direct_result | incomplete | 15 |
| direct_result | ok | 66 |
| direct_result | unknown | 18 |
| direct_result | unsupported | 44 |
| independent_realization |  | 25 |
| independent_realization | ambiguous | 162 |
| independent_realization | incomplete | 26 |
| independent_realization | ok | 81 |
| independent_realization | unknown | 47 |
| independent_realization | unsupported | 421 |
| integration | ambiguous | 120 |
| integration | incomplete | 70 |
| integration | ok | 20 |
| integration | unknown | 28 |
| integration | unsupported | 762 |
| native_analysis | ambiguous | 38 |
| native_analysis | incomplete | 47 |
| native_analysis | ok | 143 |
| native_analysis | unknown | 23 |
| native_analysis | unsupported | 761 |
| source_gap_probe | ok | 1 |
| source_gap_probe | unknown | 1 |
| source_gap_probe | unsupported | 346 |
| translation | invalid | 2 |
| translation | unknown | 2 |
| translation | unsupported | 12 |

## Authored independent realization Meaning inputs

These are the authored graphs supplied to the realizer, distinct from source analyzer and runtime result graphs. Repeated identical inputs are deduplicated by case/stage/MeaningID.

- Unique authored Meaning inputs observed: 762

| Authored graph property | Value | Count |
|---|---|---:|
| node kind | discourse | 48 |
| node kind | event | 662 |
| node kind | quantifier | 8 |
| node kind | question | 38 |
| node kind | referent | 1390 |
| node kind | scope | 128 |
| node kind | state | 202 |
| participant role | agent | 52 |
| participant role | bearer | 136 |
| participant role | beneficiary | 12 |
| participant role | buyer | 4 |
| participant role | carrier | 4 |
| participant role | drinker | 46 |
| participant role | eater | 36 |
| participant role | experiencer | 304 |
| participant role | food | 36 |
| participant role | giver | 32 |
| participant role | goods | 8 |
| participant role | helper | 12 |
| participant role | holder | 4 |
| participant role | liquid | 46 |
| participant role | load | 4 |
| participant role | material | 82 |
| participant role | member | 66 |
| participant role | mover | 6 |
| participant role | object | 30 |
| participant role | organism | 10 |
| participant role | person | 2 |
| participant role | reader | 76 |
| participant role | recipient | 46 |
| participant role | runner | 18 |
| participant role | seller | 4 |
| participant role | sleeper | 28 |
| participant role | stimulus | 300 |
| participant role | target | 20 |
| participant role | thanker | 4 |
| participant role | theme | 46 |
| participant role | walker | 8 |
| participant role | worker | 2 |
| temporal_frame | past | 109 |
| temporal_frame | present | 677 |
| temporal_frame | unknown | 1690 |
| referent multiplicity | many | 232 |
| referent multiplicity | one | 1048 |
| referent multiplicity | unknown | 110 |

## Observed source/result Meaning graphs

Authored `meaning_input` supplied to independent realization is excluded from observed graphs below. Graph summaries deduplicate identical graphs within a case and stage. Missing temporal frames or multiplicities are counted as unknown.

- Unique observed graph snapshots: 679

| Graph property | Value | Count |
|---|---|---:|
| node kind | event | 534 |
| node kind | referent | 1205 |
| node kind | scope | 48 |
| node kind | state | 184 |
| predicate (event) | `concept:hearing` | 112 |
| predicate (event) | `concept:looking` | 24 |
| predicate (event) | `concept:reading` | 140 |
| predicate (event) | `concept:receiving` | 22 |
| predicate (event) | `concept:seeing` | 236 |
| predicate (state) | `concept:black` | 81 |
| predicate (state) | `concept:class_membership` | 103 |
| participant role | agent | 24 |
| participant role | bearer | 81 |
| participant role | experiencer | 344 |
| participant role | material | 140 |
| participant role | member | 103 |
| participant role | reader | 140 |
| participant role | recipient | 22 |
| participant role | stimulus | 344 |
| participant role | target | 24 |
| participant role | theme | 22 |
| temporal_frame | past | 83 |
| temporal_frame | present | 592 |
| temporal_frame | unknown | 1296 |
| referent multiplicity | many | 221 |
| referent multiplicity | one | 951 |
| referent multiplicity | unknown | 33 |

## Observed trace and feature associations

Trace rows are deduplicated by case/stage/family/variant/frame/outcome. Feature values are observed only when present in emitted result traces.

| Family ID | Variant ID | Frame ID | Outcome | Trace records |
|---|---|---|---|---:|
| `construction:en.attributed_seeing` | `construction:en.attributed_seeing.svo` | `en.see.visual.transitive` | ok | 41 |
| `construction:en.attributed_seeing` | `construction:en.attributed_seeing.svo.indefinite_object` | `en.see.visual.transitive` | ok | 18 |
| `construction:en.attributed_seeing` | `construction:en.attributed_seeing.svo.indefinite_object_an` | `en.see.visual.transitive` | ok | 2 |
| `construction:en.attributed_seeing` | `construction:en.attributed_seeing.svo.indefinite_subject` | `en.see.visual.transitive` | ok | 20 |
| `construction:en.attributed_seeing` | `construction:en.attributed_seeing.svo.indefinite_subject_object` | `en.see.visual.transitive` | ok | 19 |
| `construction:en.attribution` | `construction:en.attribution.that` | `en.black.color.frame` | ok | 12 |
| `construction:en.attribution` | `construction:en.attribution.this` | `en.black.color.frame` | ok | 37 |
| `construction:en.identification` | `construction:en.identification.it` | `en.be.class_membership.frame` | ok | 26 |
| `construction:en.identification` | `construction:en.identification.that` | `en.be.class_membership.frame` | ok | 26 |
| `construction:en.identification` | `construction:en.identification.this` | `en.be.class_membership.frame` | ok | 69 |
| `construction:en.identification_negative` | `construction:en.identification_negative.it` | `en.be.class_membership.frame` | ok | 2 |
| `construction:en.identification_negative` | `construction:en.identification_negative.that` | `en.be.class_membership.frame` | ok | 2 |
| `construction:en.identification_negative` | `construction:en.identification_negative.this` | `en.be.class_membership.frame` | ok | 9 |
| `construction:en.oblique` | `construction:en.oblique.svo` | `en.look.directed.at` | ok | 20 |
| `construction:en.oblique` | `construction:en.oblique.svo.indefinite_subject` | `en.look.directed.at` | ok | 10 |
| `construction:en.oblique` | `construction:en.oblique.svo.indefinite_subject_an` | `en.look.directed.at` | ok | 2 |
| `construction:en.oblique` | `construction:en.oblique.svo.indefinite_subject_an_target_a` | `en.look.directed.at` | ok | 2 |
| `construction:en.oblique` | `construction:en.oblique.svo.indefinite_subject_target` | `en.look.directed.at` | ok | 10 |
| `construction:en.oblique` | `construction:en.oblique.svo.indefinite_target` | `en.look.directed.at` | ok | 10 |
| `construction:en.question.wh` | `construction:en.question.wh.object.neutral.analyze` | `en.see.visual.transitive` | ok | 2 |
| `construction:en.question.wh` | `construction:en.question.wh.subject` | `en.see.visual.transitive` | ok | 2 |
| `construction:en.transitive` | `construction:en.transitive.svo` | `en.hear.auditory.transitive` | ok | 59 |
| `construction:en.transitive` | `construction:en.transitive.svo` | `en.see.visual.transitive` | ok | 116 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_object.subject_a` | `en.hear.auditory.transitive` | ok | 1 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_object.subject_a` | `en.see.visual.transitive` | ok | 2 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_object.subject_the` | `en.hear.auditory.transitive` | ok | 7 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_object.subject_the` | `en.see.visual.transitive` | ok | 15 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.analyze` | `en.read.text.transitive` | ok | 1 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.analyze` | `en.see.visual.transitive` | ok | 3 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_a` | `en.hear.auditory.transitive` | ok | 6 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_a` | `en.read.text.transitive` | ok | 15 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_a` | `en.see.visual.transitive` | ok | 4 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_bare` | `en.hear.auditory.transitive` | ok | 6 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_bare` | `en.see.visual.transitive` | ok | 14 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_the` | `en.hear.auditory.transitive` | ok | 12 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_the` | `en.read.text.transitive` | ok | 15 |
| `construction:en.transitive` | `construction:en.transitive.svo.bare_subject.object_the` | `en.see.visual.transitive` | ok | 17 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_object` | `en.hear.auditory.transitive` | ok | 21 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_object` | `en.read.text.transitive` | ok | 23 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_object` | `en.see.visual.transitive` | ok | 42 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_object_an` | `en.hear.auditory.transitive` | ok | 1 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_object_an` | `en.see.visual.transitive` | ok | 5 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject` | `en.hear.auditory.transitive` | ok | 16 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject` | `en.read.text.transitive` | ok | 8 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject` | `en.see.visual.transitive` | ok | 40 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_a_object_an` | `en.hear.auditory.transitive` | ok | 1 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_a_object_an` | `en.see.visual.transitive` | ok | 5 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_an` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_an` | `en.see.visual.transitive` | ok | 6 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_an_object_a` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_an_object_a` | `en.see.visual.transitive` | ok | 6 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_object` | `en.hear.auditory.transitive` | ok | 15 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_object` | `en.read.text.transitive` | ok | 9 |
| `construction:en.transitive` | `construction:en.transitive.svo.indefinite_subject_object` | `en.see.visual.transitive` | ok | 38 |
| `construction:en.transitive` | `construction:en.transitive.svo.reading` | `en.read.text.transitive` | ok | 55 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo` | `en.hear.auditory.transitive` | ok | 9 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo` | `en.see.visual.transitive` | ok | 17 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.bare_object.subject_a` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.bare_object.subject_the` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_object` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_object` | `en.see.visual.transitive` | ok | 8 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_object_an` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_subject` | `en.hear.auditory.transitive` | ok | 6 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_subject` | `en.see.visual.transitive` | ok | 8 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_subject_a_object_an` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_subject_object` | `en.hear.auditory.transitive` | ok | 3 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.indefinite_subject_object` | `en.see.visual.transitive` | ok | 8 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural` | `en.hear.auditory.transitive` | ok | 4 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural` | `en.see.visual.transitive` | ok | 6 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural.indefinite_object` | `en.hear.auditory.transitive` | ok | 2 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural.indefinite_object` | `en.see.visual.transitive` | ok | 4 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural.subject_bare.object_a` | `en.hear.auditory.transitive` | ok | 2 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural.subject_bare.object_a` | `en.see.visual.transitive` | ok | 4 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural.subject_bare.object_the` | `en.hear.auditory.transitive` | ok | 2 |
| `construction:en.transitive_negative` | `construction:en.transitive_negative.svo.plural.subject_bare.object_the` | `en.see.visual.transitive` | ok | 4 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo` | `en.get.receive.transitive` | ok | 17 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo` | `en.see.visual.transitive` | ok | 8 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.bare_subject.analyze` | `en.read.text.transitive` | ok | 1 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.bare_subject.object_a` | `en.read.text.transitive` | ok | 4 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.bare_subject.object_a` | `en.see.visual.transitive` | ok | 2 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.bare_subject.object_the` | `en.read.text.transitive` | ok | 4 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.bare_subject.object_the` | `en.see.visual.transitive` | ok | 2 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_object` | `en.get.receive.transitive` | ok | 8 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_object` | `en.read.text.transitive` | ok | 10 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_subject` | `en.get.receive.transitive` | ok | 8 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_subject` | `en.read.text.transitive` | ok | 6 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_subject_an` | `en.get.receive.transitive` | ok | 2 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_subject_an_object_a` | `en.get.receive.transitive` | ok | 2 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_subject_object` | `en.get.receive.transitive` | ok | 8 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.indefinite_subject_object` | `en.read.text.transitive` | ok | 6 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.reading` | `en.read.text.transitive` | ok | 24 |
| `construction:en.transitive_past` | `construction:en.transitive_past.svo.reading.realize` | `en.read.text.transitive` | ok | 10 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first` | `en.hear.auditory.transitive` | ok | 10 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first` | `en.read.text.transitive` | ok | 17 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first` | `en.see.visual.transitive` | ok | 12 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first.bare_object` | `en.see.visual.transitive` | ok | 1 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first.indefinite_object` | `en.hear.auditory.transitive` | ok | 4 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first.indefinite_object` | `en.read.text.transitive` | ok | 8 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.first.indefinite_object` | `en.see.visual.transitive` | ok | 4 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.reading.first` | `en.read.text.transitive` | ok | 9 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.reading.second` | `en.read.text.transitive` | ok | 6 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second` | `en.hear.auditory.transitive` | ok | 10 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second` | `en.read.text.transitive` | ok | 11 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second` | `en.see.visual.transitive` | ok | 10 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second.indefinite_object` | `en.hear.auditory.transitive` | ok | 4 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second.indefinite_object` | `en.read.text.transitive` | ok | 5 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second.indefinite_object` | `en.see.visual.transitive` | ok | 4 |
| `construction:en.transitive_personal` | `construction:en.transitive_personal.svo.second.indefinite_object_an` | `en.hear.auditory.transitive` | ok | 2 |
| `construction:pl.attributed_seeing` | `construction:pl.attributed_seeing.svo` | `pl.widziec.visual.transitive` | ok | 33 |
| `construction:pl.attribution` | `construction:pl.attribution.ten` | `pl.czarny.black.frame` | ok | 22 |
| `construction:pl.attribution` | `construction:pl.attribution.to` | `pl.czarny.black.frame` | ok | 10 |
| `construction:pl.identification` | `construction:pl.identification.to` | `pl.byc.class_membership.frame` | ok | 85 |
| `construction:pl.identification_negative` | `construction:pl.identification_negative.to` | `pl.byc.class_membership.frame` | ok | 9 |
| `construction:pl.oblique` | `construction:pl.oblique.svo` | `pl.patrzec.directed.na` | ok | 20 |
| `construction:pl.question.wh` | `construction:pl.question.wh.object.fronted` | `pl.widziec.visual.transitive` | ok | 2 |
| `construction:pl.question.wh` | `construction:pl.question.wh.subject` | `pl.widziec.visual.transitive` | ok | 2 |
| `construction:pl.transitive` | `construction:pl.transitive.svo` | `pl.czytac.interpret.frame` | ok | 69 |
| `construction:pl.transitive` | `construction:pl.transitive.svo` | `pl.dostac.receive.transitive` | ok | 17 |
| `construction:pl.transitive` | `construction:pl.transitive.svo` | `pl.slyszec.auditory.transitive` | ok | 54 |
| `construction:pl.transitive` | `construction:pl.transitive.svo` | `pl.widziec.visual.transitive` | ok | 118 |
| `construction:pl.transitive_negative` | `construction:pl.transitive_negative.svo` | `pl.slyszec.auditory.transitive_negative` | ok | 13 |
| `construction:pl.transitive_negative` | `construction:pl.transitive_negative.svo` | `pl.widziec.visual.transitive_negative` | ok | 24 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.first.plural.masculine_personal` | `pl.czytac.interpret.frame` | ok | 4 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.first.plural.non_masculine_personal` | `pl.czytac.interpret.frame` | ok | 4 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.first.singular.feminine` | `pl.czytac.interpret.frame` | ok | 2 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.first.singular.masculine_personal` | `pl.czytac.interpret.frame` | ok | 2 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.second.plural.masculine_personal` | `pl.czytac.interpret.frame` | ok | 4 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.second.plural.non_masculine_personal` | `pl.czytac.interpret.frame` | ok | 4 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.second.singular.feminine` | `pl.czytac.interpret.frame` | ok | 4 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.past.second.singular.masculine_personal` | `pl.czytac.interpret.frame` | ok | 4 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.first.plural` | `pl.czytac.interpret.frame` | ok | 10 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.first.plural` | `pl.slyszec.auditory.transitive` | ok | 5 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.first.plural` | `pl.widziec.visual.transitive` | ok | 5 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.first.singular` | `pl.czytac.interpret.frame` | ok | 5 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.first.singular` | `pl.slyszec.auditory.transitive` | ok | 5 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.first.singular` | `pl.widziec.visual.transitive` | ok | 6 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.second.plural` | `pl.czytac.interpret.frame` | ok | 7 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.second.plural` | `pl.slyszec.auditory.transitive` | ok | 7 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.second.plural` | `pl.widziec.visual.transitive` | ok | 7 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.second.singular` | `pl.czytac.interpret.frame` | ok | 7 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.second.singular` | `pl.slyszec.auditory.transitive` | ok | 7 |
| `construction:pl.transitive_personal` | `construction:pl.transitive_personal.prodrop.present.second.singular` | `pl.widziec.visual.transitive` | ok | 7 |

### Observed features

| Feature | State | Value | Deduplicated observations |
|---|---|---|---:|
| `adjective_degree` | known | positive | 48 |
| `aspect` | known | imperfective | 256 |
| `aspect` | known | perfective | 9 |
| `case` | known | accusative | 184 |
| `case` | known | genitive | 20 |
| `case` | known | nominative | 242 |
| `degree` | known | positive | 35 |
| `determiner` | known | bare | 8 |
| `determiner` | known | definite | 228 |
| `determiner` | known | indefinite | 63 |
| `gender` | known | feminine | 81 |
| `gender` | known | masculine_animal | 152 |
| `gender` | known | masculine_inanimate | 25 |
| `gender` | known | masculine_personal | 99 |
| `gender` | known | neuter | 44 |
| `gender` | known | non_masculine_personal | 2 |
| `mood` | known | indicative | 543 |
| `noun_number` | known | plural | 59 |
| `noun_number` | known | singular | 281 |
| `number` | known | plural | 162 |
| `number` | known | singular | 549 |
| `person` | known | first | 44 |
| `person` | known | second | 36 |
| `person` | known | third | 585 |
| `pronoun_case` | known | subject | 40 |
| `pronoun_number` | known | plural | 13 |
| `pronoun_number` | known | singular | 9 |
| `pronoun_person` | known | first | 22 |
| `pronoun_person` | known | second | 18 |
| `tense` | known | past | 79 |
| `tense` | known | present | 497 |
| `verb_form` | known | base | 18 |
| `verb_form` | known | finite | 525 |
| `verb_number` | known | plural | 83 |
| `verb_number` | known | singular | 175 |
| `voice` | known | active | 543 |

## Representation gap diagnosis

Gaps are grouped by required predicate and authored missing-representation reason. Diagnostic codes shown are observed codes from the raw log; absence of a code is reported explicitly.

| Layer | Required predicate | Missing representation / reason | Observed diagnostic codes | Records |
|---|---|---|---|---|
| Meaning representation | `accompaniment` | No accompaniment relation/predicate is licensed; another social or ownership assertion would change meaning. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1181-EN-PL`, `A1.11-CASE-1182-PL-EN` |
| Meaning representation | `belonging_association` | CORE has no generic BELONG TO association predicate with associate/item roles. It is not equivalent to part-whole, temporary possession, or legal ownership; candidate possession concepts are KindRelation and cannot express the complete sentence as Meaning. No partial graph is supplied. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-636-PL-EN` |
| Meaning representation | `belonging_association` | CORE has no generic BELONG TO association predicate with associate/item roles. It is not equivalent to part-whole, temporary possession, or legal ownership; candidate possession concepts are KindRelation and cannot express the complete sentence as Meaning. No partial graph is supplied. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-633-EN-PL`, `A1.6-PO-634-PL-EN`, `A1.6-PO-635-EN-PL` |
| Meaning representation | `between_position` | The full sentence requires the present-time spatial relation between_position with roles located, landmark_a, landmark_b; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve both distinct landmark anchors; their order is not asserted. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-729-EN-PL`, `A1.7-LQ-730-PL-EN`, `A1.7-LQ-731-EN-PL`, `A1.7-LQ-732-PL-EN` |
| Meaning representation | `black` | The question requires an unknown choice variable restricted to book referents and bound to the black bearer role; Question has no such variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1013-EN-PL`, `A1.9-Q-1014-PL-EN` |
| Meaning representation | `black` | The source asserts a future event or state, but Meaning.TemporalFrame represents only past or present; the future frame value/operator cannot be represented. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1103-EN-PL`, `A1.10-TIME-1104-PL-EN` |
| Meaning representation | `blue` | CORE has black but no blue property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.5-ID-574-PL-EN` |
| Meaning representation | `blue` | CORE has black but no blue property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.5-ID-573-EN-PL` |
| Meaning representation | `closing_object` | Meaning v1 request act does not preserve explicit please/politeness; request force alone does not meet the selected full meaning. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1343-EN-PL`, `A1.13-COMP-1344-PL-EN` |
| Meaning representation | `conversation` | No reciprocal conversation predicate is licensed; speaking would assert a different one-way event. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1179-EN-PL`, `A1.11-CASE-1180-PL-EN` |
| Meaning representation | `crying_emotionally` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-975-EN-PL`, `A1.9-Q-976-PL-EN` |
| Meaning representation | `cutting_object` | CORE cutting_object omits the instrument role; knife_utensil is only a referent concept. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1175-EN-PL`, `A1.11-CASE-1176-PL-EN` |
| Meaning representation | `drinking` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-935-EN-PL`, `A1.9-Q-936-PL-EN` |
| Meaning representation | `drinking` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-933-EN-PL`, `A1.9-Q-934-PL-EN` |
| Meaning representation | `drinking` | Negation over the existentially bound cat participant is required: the actor must be quantified, while negating one event does not quantify its drinker. Meaning has no typed bound-variable link from a quantifier to that event role. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.8-NEG-887-EN-PL` |
| Meaning representation | `drinking` | Negation over the existentially bound cat participant is required: the actor must be quantified, while negating one event does not quantify its drinker. Meaning has no typed bound-variable link from a quantifier to that event role. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-888-PL-EN` |
| Meaning representation | `drinking` | The requested amount requires a typed quantity variable bound to the liquid role; omission of that argument does not encode a WH amount. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-969-EN-PL`, `A1.9-Q-970-PL-EN` |
| Meaning representation | `drinking` | The source asserts a future event or state, but Meaning.TemporalFrame represents only past or present; the future frame value/operator cannot be represented. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1101-EN-PL`, `A1.10-TIME-1102-PL-EN` |
| Meaning representation | `drinking+reading` | The source asserts two past events with temporal overlap in a when-clause; Meaning has no temporal relation/binding between event nodes, so event overlap and the linked reading cannot be represented. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1109-EN-PL`, `A1.10-TIME-1110-PL-EN` |
| Meaning representation | `eating` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-931-EN-PL`, `A1.9-Q-932-PL-EN` |
| Meaning representation | `eating` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-929-EN-PL`, `A1.9-Q-930-PL-EN` |
| Meaning representation | `eating` | The question requires an explicit unknown time variable bound to the event; a past frame alone does not encode a WH-time query. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-971-EN-PL`, `A1.9-Q-972-PL-EN` |
| Meaning representation | `eating` | The requested count requires a typed quantity variable bound to the food role; omission of that argument does not encode a WH count. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-963-EN-PL`, `A1.9-Q-964-PL-EN` |
| Meaning representation | `eating` | The sentence requires explicit negative scope over a frequency restriction (never / not always); no frequency operator is represented in Meaning. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-883-EN-PL`, `A1.8-NEG-884-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | CORE existence has no participant role to bind the quantified cats, and TemporalFrame is forbidden on a quantifier or referent. The supported exactly-two-cats component alone does not assert finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1-LQ-003-EN-PL`, `A1-LQ-004-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-1 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→book(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-795-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-1 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→book(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-796-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-1 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→cat(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-749-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-1 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→cat(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-750-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-1 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→window(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-783-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-1 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→window(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-784-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-10 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(10)→cat(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-757-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-10 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(10)→cat(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-758-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→apple_fruit(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-787-EN-PL`, `A1.7-LQ-788-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→bird(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-771-EN-PL`, `A1.7-LQ-772-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→book(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-763-EN-PL`, `A1.7-LQ-764-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→car(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-767-EN-PL`, `A1.7-LQ-768-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→dog(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-759-EN-PL`, `A1.7-LQ-760-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted door set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→door(many) component alone asserts cardinality, not finite present existence. The collective Polish numeral dwoje takes singular finite agreement; the counted units are doors, not pairs of doors. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.7-LQ-791-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted door set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→door(many) component alone asserts cardinality, not finite present existence. The collective Polish numeral dwoje takes singular finite agreement; the counted units are doors, not pairs of doors. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-792-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→teacher(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-775-EN-PL`, `A1.7-LQ-776-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-2 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.7-LQ-779-EN-PL`, `A1.7-LQ-780-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-3 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(3)→cat(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-751-EN-PL`, `A1.7-LQ-752-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-4 counted bag set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(4)→bag(many) component alone asserts cardinality, not finite present existence. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.7-LQ-793-EN-PL`, `A1.7-LQ-794-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-4 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(4)→cat(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-753-EN-PL`, `A1.7-LQ-754-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→apple_fruit(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-789-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→apple_fruit(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-790-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→bird(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-773-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→bird(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-774-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→book(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-765-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→book(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-766-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→car(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-769-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→car(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-770-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→cat(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-755-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→cat(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-756-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→dog(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-761-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→dog(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-762-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→teacher(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-777-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→teacher(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-778-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→window(many) component alone asserts cardinality, not finite present existence. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-785-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→window(many) component alone asserts cardinality, not finite present existence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-786-PL-EN` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.7-LQ-781-EN-PL` |
| Meaning representation | `finite_present_existence_with_exact_quantity` | The full sentence needs a present-time existence assertion linked to an exactly-5 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-782-PL-EN` |
| Meaning representation | `friendship_relationship` | CORE friendship_relationship is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles friend_a, friend_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-667-EN-PL` |
| Meaning representation | `friendship_relationship` | CORE friendship_relationship is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles friend_a, friend_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. Polish friend noun morphology does not establish the friend’s personal gender. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-668-PL-EN` |
| Meaning representation | `front_back_order` | The full sentence requires the present-time spatial relation front_back_order with roles front, back; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Front/back viewpoint remains unresolved; do not invent an observer position. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-717-EN-PL`, `A1.7-LQ-718-PL-EN`, `A1.7-LQ-719-EN-PL`, `A1.7-LQ-720-PL-EN`, `A1.7-LQ-733-EN-PL`, `A1.7-LQ-734-PL-EN`, `A1.7-LQ-735-EN-PL`, `A1.7-LQ-736-PL-EN` |
| Meaning representation | `giving` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-943-EN-PL`, `A1.9-Q-944-PL-EN`, `A1.9-Q-945-EN-PL`, `A1.9-Q-946-PL-EN` |
| Meaning representation | `giving` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-941-EN-PL`, `A1.9-Q-942-PL-EN` |
| Meaning representation | `giving` | Meaning v1 request act does not preserve explicit please/politeness; neither a bare request nor one child sense represents the selected politeness and child-sense ambiguity. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1349-EN-PL`, `A1.13-COMP-1350-PL-EN` |
| Meaning representation | `green` | CORE has black but no green property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.5-ID-575-EN-PL`, `A1.5-ID-576-PL-EN` |
| Meaning representation | `have_unspecified` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_SEMANTIC_ALTERNATIVES, UNKNOWN_SURFACE | `A1.9-Q-985-EN-PL`, `A1.9-Q-986-PL-EN` |
| Meaning representation | `have_unspecified` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-983-EN-PL`, `A1.9-Q-984-PL-EN`, `A1.9-Q-987-EN-PL`, `A1.9-Q-988-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-616-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-615-EN-PL` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-618-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-617-EN-PL`, `A1.6-PO-619-EN-PL`, `A1.6-PO-620-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-602-PL-EN`, `A1.6-PO-604-PL-EN`, `A1.6-PO-605-EN-PL`, `A1.6-PO-606-PL-EN`, `A1.6-PO-607-EN-PL`, `A1.6-PO-608-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-601-EN-PL`, `A1.6-PO-603-EN-PL`, `A1.6-PO-621-EN-PL`, `A1.6-PO-622-PL-EN`, `A1.6-PO-625-EN-PL`, `A1.6-PO-626-PL-EN`, `A1.6-PO-627-EN-PL`, `A1.6-PO-628-PL-EN`, `A1.6-PO-629-EN-PL`, `A1.6-PO-630-PL-EN`, `A1.6-PO-631-EN-PL`, `A1.6-PO-632-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-610-PL-EN` |
| Meaning representation | `have_unspecified` | CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-609-EN-PL`, `A1.6-PO-611-EN-PL`, `A1.6-PO-612-PL-EN`, `A1.6-PO-613-EN-PL`, `A1.6-PO-614-PL-EN`, `A1.6-PO-623-EN-PL`, `A1.6-PO-624-PL-EN` |
| Meaning representation | `have_unspecified` | Generic HAVE does not distinguish temporary possession from legal ownership and implies neither legal title nor duration; CORE temporary_possession is KindRelation, not a Meaning State/Event predicate, and neutral Meaning has no holder/object possession roles. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1-PO-002-PL-EN`, `A1-PO-003-EN-PL`, `A1-PO-004-PL-EN`, `A1.2-PO-201-PL-EN`, `A1.2-PO-203-PL-EN`, `A1.2-PO-205-PL-EN`, `A1.2-PO-206-EN-PL`, `A1.2-PO-207-PL-EN` |
| Meaning representation | `have_unspecified` | Generic HAVE does not distinguish temporary possession from legal ownership and implies neither legal title nor duration; CORE temporary_possession is KindRelation, not a Meaning State/Event predicate, and neutral Meaning has no holder/object possession roles. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1-PO-001-EN-PL`, `A1.2-PO-202-EN-PL`, `A1.2-PO-204-EN-PL`, `A1.2-PO-208-EN-PL` |
| Meaning representation | `have_unspecified` | The full possession proposition is required; available CORE possession predicates encode a specific relation and cannot express unspecified control or possession without that inference. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-979-EN-PL`, `A1.9-Q-980-PL-EN`, `A1.9-Q-981-EN-PL`, `A1.9-Q-982-PL-EN` |
| Meaning representation | `hearing` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1011-EN-PL`, `A1.9-Q-1012-PL-EN`, `A1.9-Q-925-EN-PL`, `A1.9-Q-926-PL-EN`, `A1.9-Q-927-EN-PL`, `A1.9-Q-928-PL-EN` |
| Meaning representation | `hearing` | The source asserts a future event or state, but Meaning.TemporalFrame represents only past or present; the future frame value/operator cannot be represented. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1099-EN-PL`, `A1.10-TIME-1100-PL-EN` |
| Meaning representation | `interior_location` | The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-710-PL-EN` |
| Meaning representation | `interior_location` | The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-707-EN-PL`, `A1.7-LQ-708-PL-EN`, `A1.7-LQ-709-EN-PL`, `A1.7-LQ-743-EN-PL`, `A1.7-LQ-744-PL-EN`, `A1.7-LQ-745-EN-PL`, `A1.7-LQ-746-PL-EN` |
| Meaning representation | `interior_location` | The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. Preserve grammatical plurality but do not infer an exact count. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.7-LQ-740-PL-EN` |
| Meaning representation | `interior_location` | The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. Preserve grammatical plurality but do not infer an exact count. | NO_SEMANTIC_ALTERNATIVES, UNKNOWN_SURFACE | `A1.7-LQ-739-EN-PL` |
| Meaning representation | `legal_ownership` | CORE legal_ownership is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles owner, asset; no partial graph is supplied. The wording explicitly licenses legal ownership, but the relation kind remains unavailable as a Meaning predicate. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-637-EN-PL`, `A1.6-PO-638-PL-EN`, `A1.6-PO-639-EN-PL`, `A1.6-PO-640-PL-EN` |
| Meaning representation | `located_in` | The full location proposition requires a located-in relation plus an explicit unknown location variable; CORE located_in is KindRelation, not a state predicate. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-953-EN-PL`, `A1.9-Q-954-PL-EN`, `A1.9-Q-955-EN-PL`, `A1.9-Q-956-PL-EN` |
| Meaning representation | `located_in` | The required located_in predicate is a CORE KindRelation and cannot be represented as a Meaning node; its CORE roles are occupant/location. The unknown location must remain unfilled, never guessed. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1-Q-003-EN-PL`, `A1-Q-004-PL-EN` |
| Meaning representation | `negative_absence_or_have` | Meaning cannot represent finite negative absence together with the source-language HAVE/pro-drop alternative without conflation. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1274-PL-EN` |
| Meaning representation | `negative_existence` | Meaning lacks a full present finite absence predicate tying existence and its discourse domain to zero cat referents; the English source licenses absence only. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1273-EN-PL` |
| Meaning representation | `on_support` | CORE has supports_physical, but no on-support assertion with figure/landmark roles in Meaning. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1104-PL-EN` |
| Meaning representation | `on_support` | CORE has supports_physical, but no on-support assertion with figure/landmark roles in Meaning. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1103-EN-PL` |
| Meaning representation | `on_support` | The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-702-PL-EN` |
| Meaning representation | `on_support` | The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-722-PL-EN` |
| Meaning representation | `on_support` | The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-701-EN-PL`, `A1.7-LQ-705-EN-PL`, `A1.7-LQ-706-PL-EN`, `A1.7-LQ-721-EN-PL` |
| Meaning representation | `on_support` | The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. Preserve grammatical plurality but do not infer an exact count. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-737-EN-PL`, `A1.7-LQ-738-PL-EN` |
| Meaning representation | `on_support` | on_support means direct support/contact, not generic located_in. CORE has no on_support predicate; related spatial_contact/supports_physical concepts are KindRelation and cannot be Meaning nodes. The required supported/support roles therefore cannot be represented. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1-LQ-001-EN-PL`, `A1-LQ-002-PL-EN` |
| Meaning representation | `open_state` | CORE has opening_object EVENT; open-state is a distinct untimed state and is not encoded. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1128-PL-EN` |
| Meaning representation | `open_state` | CORE has opening_object EVENT; open-state is a distinct untimed state and is not encoded. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1127-EN-PL` |
| Meaning representation | `open_state` | Polish drzwi is plural-only but does not establish numeric multiplicity; preserve unknown count. The full open-state predicate with a bearer role is absent from CORE. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.9-Q-1002-PL-EN` |
| Meaning representation | `open_state` | The English source explicitly presents one door; the required open-state predicate with a bearer role is absent from CORE. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1001-EN-PL` |
| Meaning representation | `open_state` | The source explicitly refers to multiple windows; the full open-state predicate with a bearer role is absent from CORE. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1003-EN-PL` |
| Meaning representation | `open_state` | The source refers to multiple windows; the full open-state predicate with a bearer role is absent from CORE. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1004-PL-EN` |
| Meaning representation | `origin_from` | The source relation is KindRelation and cannot be a Meaning predicate; the question also requires an explicit unknown source variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-957-EN-PL`, `A1.9-Q-958-PL-EN` |
| Meaning representation | `possessive_association` | CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-654-PL-EN` |
| Meaning representation | `possessive_association` | CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-641-EN-PL`, `A1.6-PO-642-PL-EN`, `A1.6-PO-643-EN-PL`, `A1.6-PO-644-PL-EN`, `A1.6-PO-645-EN-PL`, `A1.6-PO-647-EN-PL`, `A1.6-PO-648-PL-EN`, `A1.6-PO-649-EN-PL`, `A1.6-PO-650-PL-EN`, `A1.6-PO-651-EN-PL`, `A1.6-PO-652-PL-EN`, `A1.6-PO-653-EN-PL`, `A1.6-PO-655-EN-PL`, `A1.6-PO-656-PL-EN` |
| Meaning representation | `possessive_association` | CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-657-EN-PL`, `A1.6-PO-658-PL-EN` |
| Meaning representation | `possessive_association` | CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. Polish jego permits masculine or neuter reference. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-646-PL-EN` |
| Meaning representation | `proximity_relation` | The full sentence requires the present-time spatial relation proximity_relation with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Nearness has no numeric distance threshold. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-713-EN-PL`, `A1.7-LQ-714-PL-EN`, `A1.7-LQ-747-EN-PL`, `A1.7-LQ-748-PL-EN` |
| Meaning representation | `reading` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-937-EN-PL`, `A1.9-Q-938-PL-EN` |
| Meaning representation | `reading` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-939-EN-PL`, `A1.9-Q-940-PL-EN` |
| Meaning representation | `reading` | The question requires an explicit unknown time variable bound to the event; a present frame alone does not encode a WH-time query. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-973-EN-PL`, `A1.9-Q-974-PL-EN` |
| Meaning representation | `reading` | The requested count requires a typed quantity variable bound to the material role; omission of that argument does not encode a WH count. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-965-EN-PL`, `A1.9-Q-966-PL-EN` |
| Meaning representation | `reading` | The source asserts a future event or state, but Meaning.TemporalFrame represents only past or present; the future frame value/operator cannot be represented. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1097-EN-PL`, `A1.10-TIME-1098-PL-EN` |
| Meaning representation | `reading` | The source asserts an ongoing progressive event at now; Meaning has no progressive/aspect operator or event-in-progress relation, and temporal_frame cannot encode progressive aspect. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1105-EN-PL`, `A1.10-TIME-1106-PL-EN` |
| Meaning representation | `reading` | The source asserts habitual/frequency over repeated events (daily/often); Meaning has no frequency operator or temporal quantification over the event. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1111-EN-PL`, `A1.10-TIME-1112-PL-EN` |
| Meaning representation | `red` | CORE has black but no red property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.5-ID-570-PL-EN` |
| Meaning representation | `red` | CORE has black but no red property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.5-ID-569-EN-PL` |
| Meaning representation | `reflexive_possessive_association` | CORE has no generic reflexive-possessive association predicate with holder/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for reflexive association. No partial graph is supplied. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-659-EN-PL`, `A1.6-PO-660-PL-EN`, `A1.6-PO-661-EN-PL`, `A1.6-PO-662-PL-EN` |
| Meaning representation | `running` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-977-EN-PL`, `A1.9-Q-978-PL-EN` |
| Meaning representation | `sadness` | The negative sadness state requires a bearer role, but CORE sadness has no compatible bearer role; no partial graph substitutes for the complete proposition. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-855-EN-PL`, `A1.8-NEG-856-PL-EN`, `A1.8-NEG-857-EN-PL`, `A1.8-NEG-858-PL-EN`, `A1.8-NEG-859-EN-PL`, `A1.8-NEG-860-PL-EN` |
| Meaning representation | `seeing` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | AMBIGUOUS_PARSE, MISSING_REQUIRED_ROLE | `A1.9-Q-921-EN-PL`, `A1.9-Q-922-PL-EN`, `A1.9-Q-923-EN-PL`, `A1.9-Q-924-PL-EN` |
| Meaning representation | `seeing` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1009-EN-PL`, `A1.9-Q-1010-PL-EN` |
| Meaning representation | `seeing` | Meaning cannot bind the child specifically to the stated mother; child_of_person alone does not encode which parent. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1286-PL-EN` |
| Meaning representation | `seeing` | Meaning cannot bind the child specifically to the stated mother; child_of_person alone does not encode which parent. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1285-EN-PL` |
| Meaning representation | `seeing` | Meaning cannot bind “on the table” specifically to cat versus seeing event/observer without a location/attachment relation. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1278-PL-EN` |
| Meaning representation | `seeing` | Meaning cannot bind “on the table” specifically to cat versus seeing event/observer without a location/attachment relation. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1277-EN-PL` |
| Meaning representation | `seeing` | Meaning cannot bind “with a book” specifically to teacher versus observer without asserting ownership. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1279-EN-PL`, `A1.12-AMB-1280-PL-EN` |
| Meaning representation | `seeing` | Meaning cannot bind “with the child” specifically to mother versus observer without an accompaniment/attachment relation. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1275-EN-PL`, `A1.12-AMB-1276-PL-EN` |
| Meaning representation | `seeing` | Meaning has no explicit quantifier-variable binding/scope for a shared teacher with every dog. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1267-EN-PL`, `A1.12-AMB-1268-PL-EN` |
| Meaning representation | `seeing` | Meaning has no explicit quantifier-variable binding/scope for every teacher to a dog that may vary per teacher. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.12-AMB-1265-EN-PL`, `A1.12-AMB-1266-PL-EN` |
| Meaning representation | `seeing` | Meaning v1 has no heterogeneous coordinated-group representation that preserves the selected collective/distributive ambiguity; separate atomic predicates would not satisfy the group reading. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1323-EN-PL`, `A1.13-COMP-1324-PL-EN` |
| Meaning representation | `seeing` | Meaning v1 lacks a heterogeneous group and unresolved collective/distributive event representation; separate atomic predicates would not satisfy the selected subject-group reading. | NO_PARSE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1325-EN-PL`, `A1.13-COMP-1326-PL-EN` |
| Meaning representation | `seeing` | The requested count requires a typed quantity variable bound to the stimulus role; omission of that argument does not encode a WH count. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-961-EN-PL`, `A1.9-Q-962-PL-EN`, `A1.9-Q-967-EN-PL`, `A1.9-Q-968-PL-EN` |
| Meaning representation | `seeing+hearing` | Meaning v1 has no causal relation binding the two propositions. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1331-EN-PL`, `A1.13-COMP-1332-PL-EN` |
| Meaning representation | `seeing+hearing` | Meaning v1 has no contrast relation; a plain conjunction erases the selected contrast. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1363-EN-PL`, `A1.13-COMP-1364-PL-EN` |
| Meaning representation | `seeing+hearing` | Meaning v1 has no disjunction operator for alternative propositions; a root conjunction would assert the wrong relation. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1327-EN-PL`, `A1.13-COMP-1328-PL-EN` |
| Meaning representation | `seeing+hearing` | Meaning v1 has no ordered temporal relation between events; discourse order is not event sequence. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1333-EN-PL`, `A1.13-COMP-1334-PL-EN` |
| Meaning representation | `siblinghood` | CORE siblinghood is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles sibling_a, sibling_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.6-PO-663-EN-PL`, `A1.6-PO-664-PL-EN`, `A1.6-PO-665-EN-PL`, `A1.6-PO-666-PL-EN` |
| Meaning representation | `sitting_change` | CORE has no posture-change event for moving into sitting; the existing sitting posture state is not equivalent. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.13-COMP-1353-EN-PL`, `A1.13-COMP-1354-PL-EN` |
| Meaning representation | `sleeping` | Meaning v1 cannot preserve both epistemic possibility and permission readings (and the Polish ability reading) without asserting actual sleep. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1357-EN-PL`, `A1.13-COMP-1358-PL-EN` |
| Meaning representation | `sleeping` | The full sentence requires a home-location relation scoped under not-every; CORE location relation is KindRelation and cannot be a Meaning predicate. | NO_SEMANTIC_ALTERNATIVES, NO_START_PLAN, UNKNOWN_SURFACE | `A1.8-NEG-885-EN-PL` |
| Meaning representation | `sleeping` | The full sentence requires a home-location relation scoped under not-every; CORE location relation is KindRelation and cannot be a Meaning predicate. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-886-PL-EN` |
| Meaning representation | `sleeping` | The full sleeping-in-the-house proposition requires located_in(occupant, location); CORE located_in is KindRelation and cannot be a Meaning predicate. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1005-EN-PL`, `A1.9-Q-1006-PL-EN` |
| Meaning representation | `sleeping` | The question requires both a locative participant on the sleeping event and an explicit unknown location variable; the event has no location role. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-951-EN-PL`, `A1.9-Q-952-PL-EN` |
| Meaning representation | `sleeping` | The sentence requires additive focus for not only in addition to negative sleep; Meaning has no focus operator, so plain negated sleeping would omit required meaning. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-889-EN-PL`, `A1.8-NEG-890-PL-EN` |
| Meaning representation | `sleeping` | The sentence requires explicit negative scope over a frequency restriction (never / not always); no frequency operator is represented in Meaning. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-881-EN-PL`, `A1.8-NEG-882-PL-EN` |
| Meaning representation | `sleeping` | The source asserts habitual/frequency over repeated events (daily/often); Meaning has no frequency operator or temporal quantification over the event. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.10-TIME-1107-EN-PL`, `A1.10-TIME-1108-PL-EN` |
| Meaning representation | `sleeping` | “Both dogs” requires a quantifier bound to the sleeper participant; a quantifier over the event does not encode that role binding. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1015-EN-PL` |
| Meaning representation | `sleeping` | “Oba psy” requires an exactly-two quantifier bound to the sleeper participant; a quantifier over the event does not encode that role binding. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-1016-PL-EN` |
| Meaning representation | `sleeping+reading` | Meaning v1 temporal_frame records reference time per event but cannot encode overlap between events. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1335-EN-PL`, `A1.13-COMP-1336-PL-EN` |
| Meaning representation | `sleeping+running` | Meaning v1 hypothetical scope cannot encode the selected conditional implication between two propositions. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1329-EN-PL`, `A1.13-COMP-1330-PL-EN` |
| Meaning representation | `spatial_adjacency` | The full sentence requires the present-time spatial relation spatial_adjacency with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Adjacency does not mean an arbitrary proximity threshold. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-711-EN-PL`, `A1.7-LQ-712-PL-EN`, `A1.7-LQ-723-EN-PL`, `A1.7-LQ-724-PL-EN` |
| Meaning representation | `speaking+sleeping` | CORE speaking has no content role to bind the embedded sleeping proposition; adding a bare sleep root would assert it. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.13-COMP-1359-EN-PL`, `A1.13-COMP-1360-PL-EN` |
| Meaning representation | `under` | The full sentence requires the present-time spatial relation under with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-703-EN-PL`, `A1.7-LQ-704-PL-EN` |
| Meaning representation | `under` | The full sentence requires the present-time spatial relation under with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve grammatical plurality but do not infer an exact count. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-741-EN-PL`, `A1.7-LQ-742-PL-EN` |
| Meaning representation | `using_object` | CORE has no using_object predicate with user/tool roles; the tool concept alone cannot encode the action. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1173-EN-PL`, `A1.11-CASE-1174-PL-EN` |
| Meaning representation | `vertical_order` | The full sentence requires the present-time spatial relation vertical_order with roles higher, lower; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve relative higher/lower order without inferring contact or posture. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-728-PL-EN` |
| Meaning representation | `vertical_order` | The full sentence requires the present-time spatial relation vertical_order with roles higher, lower; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve relative higher/lower order without inferring contact or posture. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.7-LQ-715-EN-PL`, `A1.7-LQ-716-PL-EN`, `A1.7-LQ-725-EN-PL`, `A1.7-LQ-726-PL-EN`, `A1.7-LQ-727-EN-PL` |
| Meaning representation | `walking` | The input explicitly licenses walking; the WH destination remains unrepresented because Meaning has no typed unknown-variable binding. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-959-EN-PL`, `A1.9-Q-960-PL-EN` |
| Meaning representation | `white` | CORE has black but no white property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.5-ID-571-EN-PL`, `A1.5-ID-572-PL-EN` |
| Meaning representation | `writing_content` | Full negative writing-content proposition requires content binding on writing_message; writing_document would assert a durable document instead. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.8-NEG-809-EN-PL`, `A1.8-NEG-810-PL-EN` |
| Meaning representation | `writing_content` | The full writing-content event needs a content participant, and the WH writer needs an explicit unknown-variable binding; neither is supplied by a partial writing-message event. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-947-EN-PL`, `A1.9-Q-948-PL-EN` |
| Meaning representation | `writing_instrument` | No writing-with-instrument predicate/role is licensed. writing_document would add durable-document meaning. | UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.11-CASE-1177-EN-PL`, `A1.11-CASE-1178-PL-EN` |
| Meaning representation | `writing_message` | A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable. | NO_START_PLAN, UNKNOWN_SURFACE, UNMATCHED_PARSE_CANDIDATE | `A1.9-Q-949-EN-PL`, `A1.9-Q-950-PL-EN` |

### Diagnostic code counts

| Stage | Diagnostic code | Occurrences |
|---|---|---:|
| analysis | `EMPTY_LATTICE` | 2 |
| analysis | `NO_PARSE` | 12 |
| analysis | `NO_SEMANTIC_ALTERNATIVES` | 4 |
| analysis | `NO_START_PLAN` | 2 |
| analysis | `UNKNOWN_SURFACE` | 4 |
| analysis | `UNMATCHED_PARSE_CANDIDATE` | 51 |
| context_source_analysis | `AMBIGUOUS_PARSE` | 5 |
| context_source_analysis | `DISAMBIGUATION_INCOMPLETE` | 3 |
| context_source_analysis | `NO_PARSE` | 5 |
| context_source_analysis | `NO_SEMANTIC_ALTERNATIVES` | 7 |
| context_source_analysis | `NO_START_PLAN` | 60 |
| context_source_analysis | `PARTIAL_PARSE_EVIDENCE` | 24 |
| context_source_analysis | `UNKNOWN_SURFACE` | 111 |
| context_source_analysis | `UNMATCHED_PARSE_CANDIDATE` | 250 |
| context_source_analysis | `UNRESOLVED_ROLE_SENSE` | 24 |
| direct_result | `AMBIGUOUS_PARSE` | 2 |
| direct_result | `EXPLICIT_REALIZATION_ASSUMPTION` | 212 |
| direct_result | `MODIFIER_TRANSFER_AMBIGUOUS` | 2 |
| direct_result | `MODIFIER_TRANSFER_UNRESOLVED` | 2 |
| direct_result | `NO_PARSE` | 8 |
| direct_result | `NO_START_PLAN` | 1 |
| direct_result | `TARGET_ALTERNATIVES_AMBIGUOUS` | 35 |
| direct_result | `TARGET_FRAME_VARIANT_MISMATCH` | 53 |
| direct_result | `TARGET_LEXICAL_COVERAGE_UNAVAILABLE` | 18 |
| direct_result | `TARGET_NO_SURFACE_CANDIDATE` | 18 |
| direct_result | `UNMATCHED_PARSE_CANDIDATE` | 28 |
| independent_realization | `DISCOURSE_ACT_UNSUPPORTED` | 40 |
| independent_realization | `EXPLICIT_REALIZATION_ASSUMPTION` | 1150 |
| independent_realization | `NO_LICENSED_TARGET` | 246 |
| independent_realization | `OPERATOR_COVERAGE_UNKNOWN` | 116 |
| independent_realization | `OPERATOR_MAPPING_UNAVAILABLE` | 120 |
| independent_realization | `QUESTION_ANSWER_UNAVAILABLE` | 38 |
| independent_realization | `REALIZATION_FIELD_LOSS` | 70 |
| independent_realization | `TARGET_ALTERNATIVES_AMBIGUOUS` | 227 |
| independent_realization | `TARGET_FRAME_VARIANT_MISMATCH` | 170 |
| independent_realization | `TARGET_LEXICAL_COVERAGE_UNAVAILABLE` | 72 |
| independent_realization | `TARGET_NO_SURFACE_CANDIDATE` | 47 |
| independent_realization | `TARGET_OPERATOR_COVERAGE_UNAVAILABLE` | 116 |
| integration | `AMBIGUOUS_PARSE` | 17 |
| integration | `DISAMBIGUATION_INCOMPLETE` | 22 |
| integration | `EXPLICIT_REALIZATION_ASSUMPTION` | 800 |
| integration | `MISSING_REQUIRED_ROLE` | 20 |
| integration | `NO_PARSE` | 106 |
| integration | `NO_SEMANTIC_ALTERNATIVES` | 23 |
| integration | `NO_START_PLAN` | 224 |
| integration | `OPERATOR_COVERAGE_UNKNOWN` | 3 |
| integration | `PARTIAL_PARSE_EVIDENCE` | 231 |
| integration | `PRINCIPAL_PARTS_UNAVAILABLE` | 2 |
| integration | `REALIZATION_FIELD_LOSS` | 64 |
| integration | `TARGET_ALTERNATIVES_AMBIGUOUS` | 118 |
| integration | `TARGET_FRAME_VARIANT_MISMATCH` | 74 |
| integration | `TARGET_NO_SURFACE_CANDIDATE` | 16 |
| integration | `TRANSLATION_NOT_OK` | 16 |
| integration | `UNKNOWN_SURFACE` | 1068 |
| integration | `UNMATCHED_PARSE_CANDIDATE` | 4829 |
| integration | `UNRESOLVED_ROLE_SENSE` | 231 |
| native_analysis | `AMBIGUOUS_PARSE` | 17 |
| native_analysis | `DISAMBIGUATION_INCOMPLETE` | 30 |
| native_analysis | `MISSING_REQUIRED_ROLE` | 12 |
| native_analysis | `NO_PARSE` | 106 |
| native_analysis | `NO_SEMANTIC_ALTERNATIVES` | 23 |
| native_analysis | `NO_START_PLAN` | 224 |
| native_analysis | `OPERATOR_COVERAGE_UNKNOWN` | 2 |
| native_analysis | `PARTIAL_PARSE_EVIDENCE` | 91 |
| native_analysis | `UNKNOWN_SURFACE` | 1068 |
| native_analysis | `UNMATCHED_PARSE_CANDIDATE` | 2795 |
| native_analysis | `UNRESOLVED_ROLE_SENSE` | 91 |
| source_gap_probe | `NO_PARSE` | 81 |
| source_gap_probe | `NO_SEMANTIC_ALTERNATIVES` | 1 |
| source_gap_probe | `NO_START_PLAN` | 22 |
| source_gap_probe | `UNKNOWN_SURFACE` | 358 |
| source_gap_probe | `UNMATCHED_PARSE_CANDIDATE` | 1363 |
| translation | `INVALID_TRANSLATION_INPUT` | 2 |
| translation | `NO_PARSE` | 12 |
| translation | `NO_SEMANTIC_ALTERNATIVES` | 2 |
| translation | `NO_START_PLAN` | 2 |
| translation | `UNKNOWN_SURFACE` | 4 |
| translation | `UNMATCHED_PARSE_CANDIDATE` | 90 |

## Counts

PASS: 475 | FAIL: 140 | GAP_ANALYZE: 856 | GAP_REALIZE: 449 | GAP_I: 1080 | SPEC_PENDING: 0 | INFRA: 0 | NOT RUN: 0

`SPEC_PENDING` means the check ran but lacks a closed independent contract. `NOT RUN` means no matching record check result was observed. Runtime adapter NOT RUN markers are counted separately above. Additional checks retain their own statuses. Consult the full raw log artifact for every individual observation.

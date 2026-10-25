# Szczegółowe wymagania bez pełnego I

115 różnych opisów wymagań z istniejącego inwentarza, 360 rekordów. Status raportu GAP_I pozostaje bez zmian. Kategoria przyczyny w audycie rozdziela brak danych CORE od ograniczenia schematu I. Nie wszystkie dodatkowe blokery są rozstrzygnięte; oryginalne opisy zachowano poniżej. Rodziny i granice diagnozy: [katalog](causes.md). Dowód wykonania każdego etapu: [checks.tsv](checks.tsv), numery linii w rozpakowanym matrix.log.

## G001 D-CORE

on_support means direct support/contact, not generic located_in. CORE has no on_support predicate; related spatial_contact/supports_physical concepts are KindRelation and cannot be Meaning nodes. The required supported/support roles therefore cannot be represented.

Rekordy: `A1-LQ-001-EN-PL`, `A1-LQ-002-PL-EN`.

## G002 D-EXIST

CORE existence has no participant role to bind the quantified cats, and TemporalFrame is forbidden on a quantifier or referent. The supported exactly-two-cats component alone does not assert finite present existence.

Rekordy: `A1-LQ-003-EN-PL`, `A1-LQ-004-PL-EN`.

## G003 D-CORE

Generic HAVE does not distinguish temporary possession from legal ownership and implies neither legal title nor duration; CORE temporary_possession is KindRelation, not a Meaning State/Event predicate, and neutral Meaning has no holder/object possession roles.

Rekordy: `A1-PO-001-EN-PL`, `A1-PO-002-PL-EN`, `A1-PO-003-EN-PL`, `A1-PO-004-PL-EN`, `A1.2-PO-201-PL-EN`, `A1.2-PO-202-EN-PL`, `A1.2-PO-203-PL-EN`, `A1.2-PO-204-EN-PL`, `A1.2-PO-205-PL-EN`, `A1.2-PO-206-EN-PL`, `A1.2-PO-207-PL-EN`, `A1.2-PO-208-EN-PL`.

## G004 I-REL

The required located_in predicate is a CORE KindRelation and cannot be represented as a Meaning node; its CORE roles are occupant/location. The unknown location must remain unfilled, never guessed.

Rekordy: `A1-Q-003-EN-PL`, `A1-Q-004-PL-EN`.

## G005 D-CORE

CORE has black but no red property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture

Rekordy: `A1.5-ID-569-EN-PL`, `A1.5-ID-570-PL-EN`.

## G006 D-CORE

CORE has black but no white property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture

Rekordy: `A1.5-ID-571-EN-PL`, `A1.5-ID-572-PL-EN`.

## G007 D-CORE

CORE has black but no blue property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture

Rekordy: `A1.5-ID-573-EN-PL`, `A1.5-ID-574-PL-EN`.

## G008 D-CORE

CORE has black but no green property predicate; generic color has no bearer role/value, so this sentence has no honest MeaningGraph fixture

Rekordy: `A1.5-ID-575-EN-PL`, `A1.5-ID-576-PL-EN`.

## G009 D-CORE

CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied.

Rekordy: `A1.6-PO-601-EN-PL`, `A1.6-PO-602-PL-EN`, `A1.6-PO-603-EN-PL`, `A1.6-PO-604-PL-EN`, `A1.6-PO-605-EN-PL`, `A1.6-PO-606-PL-EN`, `A1.6-PO-607-EN-PL`, `A1.6-PO-608-PL-EN`, `A1.6-PO-621-EN-PL`, `A1.6-PO-622-PL-EN`, `A1.6-PO-625-EN-PL`, `A1.6-PO-626-PL-EN`, `A1.6-PO-627-EN-PL`, `A1.6-PO-628-PL-EN`, `A1.6-PO-629-EN-PL`, `A1.6-PO-630-PL-EN`, `A1.6-PO-631-EN-PL`, `A1.6-PO-632-PL-EN`.

## G010 D-CORE

CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved.

Rekordy: `A1.6-PO-609-EN-PL`, `A1.6-PO-610-PL-EN`, `A1.6-PO-611-EN-PL`, `A1.6-PO-612-PL-EN`, `A1.6-PO-613-EN-PL`, `A1.6-PO-614-PL-EN`, `A1.6-PO-623-EN-PL`, `A1.6-PO-624-PL-EN`.

## G011 D-CORE

CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied.

Rekordy: `A1.6-PO-615-EN-PL`, `A1.6-PO-616-PL-EN`.

## G012 D-CORE

CORE has no generic HAVE predicate with holder/object roles. Candidate concepts temporary_possession and legal_ownership are KindRelation, not Meaning State/Event predicates; neither represents generic availability/control or its negation without title or duration. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved.

Rekordy: `A1.6-PO-617-EN-PL`, `A1.6-PO-618-PL-EN`, `A1.6-PO-619-EN-PL`, `A1.6-PO-620-PL-EN`.

## G013 D-CORE

CORE has no generic BELONG TO association predicate with associate/item roles. It is not equivalent to part-whole, temporary possession, or legal ownership; candidate possession concepts are KindRelation and cannot express the complete sentence as Meaning. No partial graph is supplied.

Rekordy: `A1.6-PO-633-EN-PL`, `A1.6-PO-634-PL-EN`, `A1.6-PO-635-EN-PL`, `A1.6-PO-636-PL-EN`.

## G014 I-REL

CORE legal_ownership is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles owner, asset; no partial graph is supplied. The wording explicitly licenses legal ownership, but the relation kind remains unavailable as a Meaning predicate.

Rekordy: `A1.6-PO-637-EN-PL`, `A1.6-PO-638-PL-EN`, `A1.6-PO-639-EN-PL`, `A1.6-PO-640-PL-EN`.

## G015 D-CORE

CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied.

Rekordy: `A1.6-PO-641-EN-PL`, `A1.6-PO-642-PL-EN`, `A1.6-PO-643-EN-PL`, `A1.6-PO-644-PL-EN`, `A1.6-PO-645-EN-PL`, `A1.6-PO-647-EN-PL`, `A1.6-PO-648-PL-EN`, `A1.6-PO-649-EN-PL`, `A1.6-PO-650-PL-EN`, `A1.6-PO-651-EN-PL`, `A1.6-PO-652-PL-EN`, `A1.6-PO-653-EN-PL`, `A1.6-PO-654-PL-EN`, `A1.6-PO-655-EN-PL`, `A1.6-PO-656-PL-EN`.

## G016 D-CORE

CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. Polish jego permits masculine or neuter reference.

Rekordy: `A1.6-PO-646-PL-EN`.

## G017 D-CORE

CORE has no generic possessive-association predicate with possessor/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for association marked by a possessive determiner. No partial graph is supplied. Collective versus distributive scope and physical copy count remain unresolved.

Rekordy: `A1.6-PO-657-EN-PL`, `A1.6-PO-658-PL-EN`.

## G018 D-CORE

CORE has no generic reflexive-possessive association predicate with holder/associated roles. temporary_possession and legal_ownership are KindRelation and cannot substitute for reflexive association. No partial graph is supplied.

Rekordy: `A1.6-PO-659-EN-PL`, `A1.6-PO-660-PL-EN`, `A1.6-PO-661-EN-PL`, `A1.6-PO-662-PL-EN`.

## G019 I-REL

CORE siblinghood is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles sibling_a, sibling_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object.

Rekordy: `A1.6-PO-663-EN-PL`, `A1.6-PO-664-PL-EN`, `A1.6-PO-665-EN-PL`, `A1.6-PO-666-PL-EN`.

## G020 I-REL

CORE friendship_relationship is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles friend_a, friend_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object.

Rekordy: `A1.6-PO-667-EN-PL`.

## G021 I-REL

CORE friendship_relationship is KindRelation, not a Meaning State/Event predicate. The neutral Meaning model cannot encode the full relation with roles friend_a, friend_b; no partial graph is supplied. Association or kinship is not legal ownership of the associated person or object. Polish friend noun morphology does not establish the friend’s personal gender.

Rekordy: `A1.6-PO-668-PL-EN`.

## G022 D-CORE

The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state.

Rekordy: `A1.7-LQ-701-EN-PL`, `A1.7-LQ-702-PL-EN`, `A1.7-LQ-705-EN-PL`, `A1.7-LQ-706-PL-EN`, `A1.7-LQ-721-EN-PL`, `A1.7-LQ-722-PL-EN`.

## G023 D-CORE

The full sentence requires the present-time spatial relation under with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation.

Rekordy: `A1.7-LQ-703-EN-PL`, `A1.7-LQ-704-PL-EN`.

## G024 I-REL

The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container.

Rekordy: `A1.7-LQ-707-EN-PL`, `A1.7-LQ-708-PL-EN`, `A1.7-LQ-709-EN-PL`, `A1.7-LQ-710-PL-EN`, `A1.7-LQ-743-EN-PL`, `A1.7-LQ-744-PL-EN`, `A1.7-LQ-745-EN-PL`, `A1.7-LQ-746-PL-EN`.

## G025 I-REL

The full sentence requires the present-time spatial relation spatial_adjacency with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Adjacency does not mean an arbitrary proximity threshold.

Rekordy: `A1.7-LQ-711-EN-PL`, `A1.7-LQ-712-PL-EN`, `A1.7-LQ-723-EN-PL`, `A1.7-LQ-724-PL-EN`.

## G026 I-REL

The full sentence requires the present-time spatial relation proximity_relation with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Nearness has no numeric distance threshold.

Rekordy: `A1.7-LQ-713-EN-PL`, `A1.7-LQ-714-PL-EN`, `A1.7-LQ-747-EN-PL`, `A1.7-LQ-748-PL-EN`.

## G027 I-REL

The full sentence requires the present-time spatial relation vertical_order with roles higher, lower; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve relative higher/lower order without inferring contact or posture.

Rekordy: `A1.7-LQ-715-EN-PL`, `A1.7-LQ-716-PL-EN`, `A1.7-LQ-725-EN-PL`, `A1.7-LQ-726-PL-EN`, `A1.7-LQ-727-EN-PL`, `A1.7-LQ-728-PL-EN`.

## G028 I-REL

The full sentence requires the present-time spatial relation front_back_order with roles front, back; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Front/back viewpoint remains unresolved; do not invent an observer position.

Rekordy: `A1.7-LQ-717-EN-PL`, `A1.7-LQ-718-PL-EN`, `A1.7-LQ-719-EN-PL`, `A1.7-LQ-720-PL-EN`, `A1.7-LQ-733-EN-PL`, `A1.7-LQ-734-PL-EN`, `A1.7-LQ-735-EN-PL`, `A1.7-LQ-736-PL-EN`.

## G029 I-REL

The full sentence requires the present-time spatial relation between_position with roles located, landmark_a, landmark_b; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve both distinct landmark anchors; their order is not asserted.

Rekordy: `A1.7-LQ-729-EN-PL`, `A1.7-LQ-730-PL-EN`, `A1.7-LQ-731-EN-PL`, `A1.7-LQ-732-PL-EN`.

## G030 D-CORE

The full sentence requires the present-time spatial relation on_support with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. CORE has no on_support predicate; support/contact cannot be represented as a Meaning state. Preserve grammatical plurality but do not infer an exact count.

Rekordy: `A1.7-LQ-737-EN-PL`, `A1.7-LQ-738-PL-EN`.

## G031 I-REL

The full sentence requires the present-time spatial relation interior_location with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Interior location does not entail full enclosure or a closed container. Preserve grammatical plurality but do not infer an exact count.

Rekordy: `A1.7-LQ-739-EN-PL`, `A1.7-LQ-740-PL-EN`.

## G032 D-CORE

The full sentence requires the present-time spatial relation under with roles located, landmark; CORE spatial relation concepts are KindRelation, not Meaning State predicates. A generic located_in relation cannot substitute for this specific relation. Preserve grammatical plurality but do not infer an exact count.

Rekordy: `A1.7-LQ-741-EN-PL`, `A1.7-LQ-742-PL-EN`.

## G033 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-1 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→cat(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-749-EN-PL`, `A1.7-LQ-750-PL-EN`.

## G034 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-3 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(3)→cat(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-751-EN-PL`, `A1.7-LQ-752-PL-EN`.

## G035 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-4 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(4)→cat(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-753-EN-PL`, `A1.7-LQ-754-PL-EN`.

## G036 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→cat(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-755-EN-PL`, `A1.7-LQ-756-PL-EN`.

## G037 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-10 counted cat set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(10)→cat(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-757-EN-PL`, `A1.7-LQ-758-PL-EN`.

## G038 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→dog(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-759-EN-PL`, `A1.7-LQ-760-PL-EN`.

## G039 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted dog set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→dog(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-761-EN-PL`, `A1.7-LQ-762-PL-EN`.

## G040 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→book(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-763-EN-PL`, `A1.7-LQ-764-PL-EN`.

## G041 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→book(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-765-EN-PL`, `A1.7-LQ-766-PL-EN`.

## G042 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→car(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-767-EN-PL`, `A1.7-LQ-768-PL-EN`.

## G043 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted car set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→car(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-769-EN-PL`, `A1.7-LQ-770-PL-EN`.

## G044 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→bird(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-771-EN-PL`, `A1.7-LQ-772-PL-EN`.

## G045 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted bird set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→bird(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-773-EN-PL`, `A1.7-LQ-774-PL-EN`.

## G046 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→teacher(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-775-EN-PL`, `A1.7-LQ-776-PL-EN`.

## G047 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted teacher set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→teacher(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-777-EN-PL`, `A1.7-LQ-778-PL-EN`.

## G048 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology.

Rekordy: `A1.7-LQ-779-EN-PL`, `A1.7-LQ-780-PL-EN`.

## G049 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted woman set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→woman(many) component alone asserts cardinality, not finite present existence. desired woman class is not a current CORE ID; do not substitute generic person or infer biology from morphology. Do not substitute person or infer biology from morphology.

Rekordy: `A1.7-LQ-781-EN-PL`, `A1.7-LQ-782-PL-EN`.

## G050 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-1 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→window(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-783-EN-PL`, `A1.7-LQ-784-PL-EN`.

## G051 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted window set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→window(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-785-EN-PL`, `A1.7-LQ-786-PL-EN`.

## G052 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→apple_fruit(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-787-EN-PL`, `A1.7-LQ-788-PL-EN`.

## G053 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-5 counted apple_fruit set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(5)→apple_fruit(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-789-EN-PL`, `A1.7-LQ-790-PL-EN`.

## G054 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-2 counted door set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(2)→door(many) component alone asserts cardinality, not finite present existence. The collective Polish numeral dwoje takes singular finite agreement; the counted units are doors, not pairs of doors.

Rekordy: `A1.7-LQ-791-EN-PL`, `A1.7-LQ-792-PL-EN`.

## G055 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-4 counted bag set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(4)→bag(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-793-EN-PL`, `A1.7-LQ-794-PL-EN`.

## G056 D-EXIST

The full sentence needs a present-time existence assertion linked to an exactly-1 counted book set in an unspecified discourse domain. CORE existence is abstract but has no participant roles to link that assertion to the counted referent; TemporalFrame is forbidden on a quantifier/referent. The valid QuantifierExactly(1)→book(many) component alone asserts cardinality, not finite present existence.

Rekordy: `A1.7-LQ-795-EN-PL`, `A1.7-LQ-796-PL-EN`.

## G057 D-CORE

Full negative writing-content proposition requires content binding on writing_message; writing_document would assert a durable document instead.

Rekordy: `A1.8-NEG-809-EN-PL`, `A1.8-NEG-810-PL-EN`.

## G058 D-CORE

The negative sadness state requires a bearer role, but CORE sadness has no compatible bearer role; no partial graph substitutes for the complete proposition.

Rekordy: `A1.8-NEG-855-EN-PL`, `A1.8-NEG-856-PL-EN`, `A1.8-NEG-857-EN-PL`, `A1.8-NEG-858-PL-EN`, `A1.8-NEG-859-EN-PL`, `A1.8-NEG-860-PL-EN`.

## G059 I-TIME

The sentence requires explicit negative scope over a frequency restriction (never / not always); no frequency operator is represented in Meaning.

Rekordy: `A1.8-NEG-881-EN-PL`, `A1.8-NEG-882-PL-EN`, `A1.8-NEG-883-EN-PL`, `A1.8-NEG-884-PL-EN`.

## G060 I-REL

The full sentence requires a home-location relation scoped under not-every; CORE location relation is KindRelation and cannot be a Meaning predicate.

Rekordy: `A1.8-NEG-885-EN-PL`, `A1.8-NEG-886-PL-EN`.

## G061 I-SCOPE

Negation over the existentially bound cat participant is required: the actor must be quantified, while negating one event does not quantify its drinker. Meaning has no typed bound-variable link from a quantifier to that event role.

Rekordy: `A1.8-NEG-887-EN-PL`, `A1.8-NEG-888-PL-EN`.

## G062 I-COMPOSE

The sentence requires additive focus for not only in addition to negative sleep; Meaning has no focus operator, so plain negated sleeping would omit required meaning.

Rekordy: `A1.8-NEG-889-EN-PL`, `A1.8-NEG-890-PL-EN`.

## G063 I-WH

A WH question requires an explicit unknown-variable binding and any stated restriction; the current Meaning Question only targets a proposition and has no WH variable.

Rekordy: `A1.9-Q-921-EN-PL`, `A1.9-Q-922-PL-EN`, `A1.9-Q-923-EN-PL`, `A1.9-Q-924-PL-EN`, `A1.9-Q-925-EN-PL`, `A1.9-Q-926-PL-EN`, `A1.9-Q-927-EN-PL`, `A1.9-Q-928-PL-EN`, `A1.9-Q-929-EN-PL`, `A1.9-Q-930-PL-EN`, `A1.9-Q-931-EN-PL`, `A1.9-Q-932-PL-EN`, `A1.9-Q-933-EN-PL`, `A1.9-Q-934-PL-EN`, `A1.9-Q-935-EN-PL`, `A1.9-Q-936-PL-EN`, `A1.9-Q-937-EN-PL`, `A1.9-Q-938-PL-EN`, `A1.9-Q-939-EN-PL`, `A1.9-Q-940-PL-EN`, `A1.9-Q-941-EN-PL`, `A1.9-Q-942-PL-EN`, `A1.9-Q-943-EN-PL`, `A1.9-Q-944-PL-EN`, `A1.9-Q-945-EN-PL`, `A1.9-Q-946-PL-EN`, `A1.9-Q-949-EN-PL`, `A1.9-Q-950-PL-EN`, `A1.9-Q-975-EN-PL`, `A1.9-Q-976-PL-EN`, `A1.9-Q-977-EN-PL`, `A1.9-Q-978-PL-EN`, `A1.9-Q-983-EN-PL`, `A1.9-Q-984-PL-EN`, `A1.9-Q-985-EN-PL`, `A1.9-Q-986-PL-EN`, `A1.9-Q-987-EN-PL`, `A1.9-Q-988-PL-EN`, `A1.9-Q-1009-EN-PL`, `A1.9-Q-1010-PL-EN`, `A1.9-Q-1011-EN-PL`, `A1.9-Q-1012-PL-EN`.

## G064 I-WH

The full writing-content event needs a content participant, and the WH writer needs an explicit unknown-variable binding; neither is supplied by a partial writing-message event.

Rekordy: `A1.9-Q-947-EN-PL`, `A1.9-Q-948-PL-EN`.

## G065 I-WH

The question requires both a locative participant on the sleeping event and an explicit unknown location variable; the event has no location role.

Rekordy: `A1.9-Q-951-EN-PL`, `A1.9-Q-952-PL-EN`.

## G066 I-WH

The full location proposition requires a located-in relation plus an explicit unknown location variable; CORE located_in is KindRelation, not a state predicate.

Rekordy: `A1.9-Q-953-EN-PL`, `A1.9-Q-954-PL-EN`, `A1.9-Q-955-EN-PL`, `A1.9-Q-956-PL-EN`.

## G067 I-REL

The source relation is KindRelation and cannot be a Meaning predicate; the question also requires an explicit unknown source variable.

Rekordy: `A1.9-Q-957-EN-PL`, `A1.9-Q-958-PL-EN`.

## G068 I-WH

The input explicitly licenses walking; the WH destination remains unrepresented because Meaning has no typed unknown-variable binding.

Rekordy: `A1.9-Q-959-EN-PL`, `A1.9-Q-960-PL-EN`.

## G069 I-WH

The requested count requires a typed quantity variable bound to the stimulus role; omission of that argument does not encode a WH count.

Rekordy: `A1.9-Q-961-EN-PL`, `A1.9-Q-962-PL-EN`, `A1.9-Q-967-EN-PL`, `A1.9-Q-968-PL-EN`.

## G070 I-WH

The requested count requires a typed quantity variable bound to the food role; omission of that argument does not encode a WH count.

Rekordy: `A1.9-Q-963-EN-PL`, `A1.9-Q-964-PL-EN`.

## G071 I-WH

The requested count requires a typed quantity variable bound to the material role; omission of that argument does not encode a WH count.

Rekordy: `A1.9-Q-965-EN-PL`, `A1.9-Q-966-PL-EN`.

## G072 I-WH

The requested amount requires a typed quantity variable bound to the liquid role; omission of that argument does not encode a WH amount.

Rekordy: `A1.9-Q-969-EN-PL`, `A1.9-Q-970-PL-EN`.

## G073 I-WH

The question requires an explicit unknown time variable bound to the event; a past frame alone does not encode a WH-time query.

Rekordy: `A1.9-Q-971-EN-PL`, `A1.9-Q-972-PL-EN`.

## G074 I-WH

The question requires an explicit unknown time variable bound to the event; a present frame alone does not encode a WH-time query.

Rekordy: `A1.9-Q-973-EN-PL`, `A1.9-Q-974-PL-EN`.

## G075 D-CORE

The full possession proposition is required; available CORE possession predicates encode a specific relation and cannot express unspecified control or possession without that inference.

Rekordy: `A1.9-Q-979-EN-PL`, `A1.9-Q-980-PL-EN`, `A1.9-Q-981-EN-PL`, `A1.9-Q-982-PL-EN`.

## G076 D-CORE

The English source explicitly presents one door; the required open-state predicate with a bearer role is absent from CORE.

Rekordy: `A1.9-Q-1001-EN-PL`.

## G077 D-CORE

Polish drzwi is plural-only but does not establish numeric multiplicity; preserve unknown count. The full open-state predicate with a bearer role is absent from CORE.

Rekordy: `A1.9-Q-1002-PL-EN`.

## G078 D-CORE

The source explicitly refers to multiple windows; the full open-state predicate with a bearer role is absent from CORE.

Rekordy: `A1.9-Q-1003-EN-PL`.

## G079 D-CORE

The source refers to multiple windows; the full open-state predicate with a bearer role is absent from CORE.

Rekordy: `A1.9-Q-1004-PL-EN`.

## G080 I-REL

The full sleeping-in-the-house proposition requires located_in(occupant, location); CORE located_in is KindRelation and cannot be a Meaning predicate.

Rekordy: `A1.9-Q-1005-EN-PL`, `A1.9-Q-1006-PL-EN`.

## G081 I-WH

The question requires an unknown choice variable restricted to book referents and bound to the black bearer role; Question has no such variable.

Rekordy: `A1.9-Q-1013-EN-PL`, `A1.9-Q-1014-PL-EN`.

## G082 I-SCOPE

“Both dogs” requires a quantifier bound to the sleeper participant; a quantifier over the event does not encode that role binding.

Rekordy: `A1.9-Q-1015-EN-PL`.

## G083 I-SCOPE

“Oba psy” requires an exactly-two quantifier bound to the sleeper participant; a quantifier over the event does not encode that role binding.

Rekordy: `A1.9-Q-1016-PL-EN`.

## G084 I-TIME

The source asserts a future event or state, but Meaning.TemporalFrame represents only past or present; the future frame value/operator cannot be represented.

Rekordy: `A1.10-TIME-1097-EN-PL`, `A1.10-TIME-1098-PL-EN`, `A1.10-TIME-1099-EN-PL`, `A1.10-TIME-1100-PL-EN`, `A1.10-TIME-1101-EN-PL`, `A1.10-TIME-1102-PL-EN`, `A1.10-TIME-1103-EN-PL`, `A1.10-TIME-1104-PL-EN`.

## G085 I-TIME

The source asserts an ongoing progressive event at now; Meaning has no progressive/aspect operator or event-in-progress relation, and temporal_frame cannot encode progressive aspect.

Rekordy: `A1.10-TIME-1105-EN-PL`, `A1.10-TIME-1106-PL-EN`.

## G086 I-TIME

The source asserts habitual/frequency over repeated events (daily/often); Meaning has no frequency operator or temporal quantification over the event.

Rekordy: `A1.10-TIME-1107-EN-PL`, `A1.10-TIME-1108-PL-EN`, `A1.10-TIME-1111-EN-PL`, `A1.10-TIME-1112-PL-EN`.

## G087 I-TIME

The source asserts two past events with temporal overlap in a when-clause; Meaning has no temporal relation/binding between event nodes, so event overlap and the linked reading cannot be represented.

Rekordy: `A1.10-TIME-1109-EN-PL`, `A1.10-TIME-1110-PL-EN`.

## G088 D-CORE

CORE has supports_physical, but no on-support assertion with figure/landmark roles in Meaning.

Rekordy: `A1.11-CASE-1103-EN-PL`, `A1.11-CASE-1104-PL-EN`.

## G089 D-CORE

CORE has opening_object EVENT; open-state is a distinct untimed state and is not encoded.

Rekordy: `A1.11-CASE-1127-EN-PL`, `A1.11-CASE-1128-PL-EN`.

## G090 D-CORE

CORE has no using_object predicate with user/tool roles; the tool concept alone cannot encode the action.

Rekordy: `A1.11-CASE-1173-EN-PL`, `A1.11-CASE-1174-PL-EN`.

## G091 D-CORE

CORE cutting_object omits the instrument role; knife_utensil is only a referent concept.

Rekordy: `A1.11-CASE-1175-EN-PL`, `A1.11-CASE-1176-PL-EN`.

## G092 D-CORE

No writing-with-instrument predicate/role is licensed. writing_document would add durable-document meaning.

Rekordy: `A1.11-CASE-1177-EN-PL`, `A1.11-CASE-1178-PL-EN`.

## G093 D-CORE

No reciprocal conversation predicate is licensed; speaking would assert a different one-way event.

Rekordy: `A1.11-CASE-1179-EN-PL`, `A1.11-CASE-1180-PL-EN`.

## G094 D-CORE

No accompaniment relation/predicate is licensed; another social or ownership assertion would change meaning.

Rekordy: `A1.11-CASE-1181-EN-PL`, `A1.11-CASE-1182-PL-EN`.

## G095 I-SCOPE

Meaning has no explicit quantifier-variable binding/scope for every teacher to a dog that may vary per teacher.

Rekordy: `A1.12-AMB-1265-EN-PL`, `A1.12-AMB-1266-PL-EN`.

## G096 I-SCOPE

Meaning has no explicit quantifier-variable binding/scope for a shared teacher with every dog.

Rekordy: `A1.12-AMB-1267-EN-PL`, `A1.12-AMB-1268-PL-EN`.

## G097 D-EXIST

Meaning lacks a full present finite absence predicate tying existence and its discourse domain to zero cat referents; the English source licenses absence only.

Rekordy: `A1.12-AMB-1273-EN-PL`.

## G098 D-EXIST

Meaning cannot represent finite negative absence together with the source-language HAVE/pro-drop alternative without conflation.

Rekordy: `A1.12-AMB-1274-PL-EN`.

## G099 D-CORE

Meaning cannot bind “with the child” specifically to mother versus observer without an accompaniment/attachment relation.

Rekordy: `A1.12-AMB-1275-EN-PL`, `A1.12-AMB-1276-PL-EN`.

## G100 D-CORE

Meaning cannot bind “on the table” specifically to cat versus seeing event/observer without a location/attachment relation.

Rekordy: `A1.12-AMB-1277-EN-PL`, `A1.12-AMB-1278-PL-EN`.

## G101 D-CORE

Meaning cannot bind “with a book” specifically to teacher versus observer without asserting ownership.

Rekordy: `A1.12-AMB-1279-EN-PL`, `A1.12-AMB-1280-PL-EN`.

## G102 D-CORE

Meaning cannot bind the child specifically to the stated mother; child_of_person alone does not encode which parent.

Rekordy: `A1.12-AMB-1285-EN-PL`, `A1.12-AMB-1286-PL-EN`.

## G103 I-COMPOSE

Meaning v1 has no heterogeneous coordinated-group representation that preserves the selected collective/distributive ambiguity; separate atomic predicates would not satisfy the group reading.

Rekordy: `A1.13-COMP-1323-EN-PL`, `A1.13-COMP-1324-PL-EN`.

## G104 I-COMPOSE

Meaning v1 lacks a heterogeneous group and unresolved collective/distributive event representation; separate atomic predicates would not satisfy the selected subject-group reading.

Rekordy: `A1.13-COMP-1325-EN-PL`, `A1.13-COMP-1326-PL-EN`.

## G105 I-COMPOSE

Meaning v1 has no disjunction operator for alternative propositions; a root conjunction would assert the wrong relation.

Rekordy: `A1.13-COMP-1327-EN-PL`, `A1.13-COMP-1328-PL-EN`.

## G106 I-COMPOSE

Meaning v1 hypothetical scope cannot encode the selected conditional implication between two propositions.

Rekordy: `A1.13-COMP-1329-EN-PL`, `A1.13-COMP-1330-PL-EN`.

## G107 I-COMPOSE

Meaning v1 has no causal relation binding the two propositions.

Rekordy: `A1.13-COMP-1331-EN-PL`, `A1.13-COMP-1332-PL-EN`.

## G108 I-TIME

Meaning v1 has no ordered temporal relation between events; discourse order is not event sequence.

Rekordy: `A1.13-COMP-1333-EN-PL`, `A1.13-COMP-1334-PL-EN`.

## G109 I-TIME

Meaning v1 temporal_frame records reference time per event but cannot encode overlap between events.

Rekordy: `A1.13-COMP-1335-EN-PL`, `A1.13-COMP-1336-PL-EN`.

## G110 I-COMPOSE

Meaning v1 request act does not preserve explicit please/politeness; request force alone does not meet the selected full meaning.

Rekordy: `A1.13-COMP-1343-EN-PL`, `A1.13-COMP-1344-PL-EN`.

## G111 I-COMPOSE

Meaning v1 request act does not preserve explicit please/politeness; neither a bare request nor one child sense represents the selected politeness and child-sense ambiguity.

Rekordy: `A1.13-COMP-1349-EN-PL`, `A1.13-COMP-1350-PL-EN`.

## G112 D-CORE

CORE has no posture-change event for moving into sitting; the existing sitting posture state is not equivalent.

Rekordy: `A1.13-COMP-1353-EN-PL`, `A1.13-COMP-1354-PL-EN`.

## G113 I-COMPOSE

Meaning v1 cannot preserve both epistemic possibility and permission readings (and the Polish ability reading) without asserting actual sleep.

Rekordy: `A1.13-COMP-1357-EN-PL`, `A1.13-COMP-1358-PL-EN`.

## G114 D-CORE

CORE speaking has no content role to bind the embedded sleeping proposition; adding a bare sleep root would assert it.

Rekordy: `A1.13-COMP-1359-EN-PL`, `A1.13-COMP-1360-PL-EN`.

## G115 I-COMPOSE

Meaning v1 has no contrast relation; a plain conjunction erases the selected contrast.

Rekordy: `A1.13-COMP-1363-EN-PL`, `A1.13-COMP-1364-PL-EN`.

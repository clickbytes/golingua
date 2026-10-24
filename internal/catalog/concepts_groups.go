package catalog

import "github.com/clickbytes/golingua/internal/provenance"

type catalogGroup struct {
	source provenance.SourceRef
	rows   []declaration
}

func draftSource() provenance.SourceRef {
	return provenance.SourceRef{
		ID:      "golingua-core-proposed-2026-10-21",
		Path:    "data/core/1000-concept-proposals.tsv",
		Version: "2026-10-21-draft",
		License: "unknown",
	}
}

func conceptGroups() []catalogGroup {
	return []catalogGroup{
		{source: authoredSource, rows: declarations},
		{source: draftSource(), rows: peopleDeclarations()},
		{source: draftSource(), rows: biologyDeclarations()},
		{source: draftSource(), rows: natureDeclarations()},
		{source: draftSource(), rows: placesDeclarations()},
		{source: draftSource(), rows: artifactsDeclarations()},
		{source: draftSource(), rows: foodGoodsDeclarations()},
		{source: draftSource(), rows: transportDeclarations()},
		{source: draftSource(), rows: everydayDeclarations()},
		{source: draftSource(), rows: eventsMotionDeclarations()},
		{source: draftSource(), rows: eventsPhysicalDeclarations()},
		{source: draftSource(), rows: eventsExchangeDeclarations()},
		{source: draftSource(), rows: eventsCognitionDeclarations()},
		{source: draftSource(), rows: eventsSocialDeclarations()},
		{source: draftSource(), rows: eventsLifeDeclarations()},
		{source: draftSource(), rows: propertiesDeclarations()},
		{source: draftSource(), rows: relationsDeclarations()},
		{source: draftSource(), rows: abstractDeclarations()},
		{source: provenance.SourceRef{
			ID:  "golingua-identification-agent-contract",
			URI: "https://github.com/clickbytes/golingua/issues/314", Version: "2026-10-24",
			License: "internal", Author: "Codex by owner delegation; owner language acceptance NOT_EVALUATED",
		}, rows: identificationDeclarations()},
	}
}

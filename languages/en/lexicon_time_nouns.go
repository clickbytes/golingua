package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishTimeNounSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_time_nouns.go",
	Version: "2026-10-22-batch-32", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name states and common calendar intervals.
var englishTimeNounEntries = []lexicalEntry{
	{"en.existence.n", "existence", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.existence.direct", "a condition in which a referent is present in a stated discourse or reality domain", "existence"}}},
	{"en.absence.n", "absence", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.absence.direct", "a specified expected referent not being present in a relevant context", "absence"}}},
	{"en.state_description.n", "state", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.state_description.direct", "a representation of conditions holding for a referent at a given time", "state_description"}}},
	{"en.change_description.n", "change", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.change_description.direct", "a representation of a difference in state across two ordered observations", "change_description"}}},
	{"en.day_interval.n", "day", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.day_interval.direct", "a calendar or approximately 24-hour time span conventionally counted as a day", "day_interval"}}},
	{"en.week_interval.n", "week", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.week_interval.direct", "a conventional sequence of seven consecutive calendar days", "week_interval"}}},
	{"en.month_interval.n", "month", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.month_interval.direct", "a named or counted subdivision of a calendar year with calendar-dependent length", "month_interval"}}},
	{"en.year_interval.n", "year", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.year_interval.direct", "a conventional calendar cycle used to count annual dates and intervals", "year_interval"}}},
	{"en.morning_interval.n", "morning", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.morning_interval.direct", "the portion of a day from its early hours until before midday", "morning_interval"}}},
	{"en.evening_interval.n", "evening", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.evening_interval.direct", "the later waking portion of a day following afternoon and preceding night", "evening_interval"}}},
	{"en.night_interval.n", "night", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.night_interval.direct", "the daily interval conventionally associated with darkness and sleep", "night_interval"}}},
	{"en.date_value.n", "date", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.date_value.direct", "a calendar designation locating an occasion at a day or greater precision", "date_value"}}},
	{"en.time_of_day.n", "time of day", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.time_of_day.direct", "a clock or conventional designation locating an instant within a day", "time_of_day"}}},
	{"en.season_interval.n", "season", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.season_interval.direct", "a recurring part of a year characterized by calendar or climatic conventions", "season_interval"}}},
	{"en.currency_unit.n", "currency unit", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.currency_unit.direct", "a standardized unit of monetary value within a currency system", "currency_unit"}}},
	{"en.temporal_relation.n", "temporal relation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.temporal_relation.direct", "a semantic relationship ordering or locating events and intervals in time", "temporal_relation"}}},
	{"en.precedes.prep", "before", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.precedes.relation", "one event or interval ends or is ordered before another under a temporal convention", "precedes"}}},
	{"en.simultaneous_with.prep", "at the same time as", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.simultaneous_with.relation", "two events or instants occur at the same temporal point or defined reference time", "simultaneous_with"}}},
}

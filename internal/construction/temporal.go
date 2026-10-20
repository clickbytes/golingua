package construction

import (
	"time"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
)

type TemporalKind string

const (
	TemporalDateKind        TemporalKind = "calendar_date"
	TemporalClockKind       TemporalKind = "clock_time"
	TemporalDurationKind    TemporalKind = "duration"
	TemporalRelativeDayKind TemporalKind = "relative_day"
	MaxTemporalOffsetDays                = 365_000
)

type CalendarDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}
type ClockTime struct {
	Hour             int  `json:"hour"`
	Minute           int  `json:"minute"`
	Second           int  `json:"second"`
	Nanosecond       int  `json:"nanosecond,omitempty"`
	UTCOffsetMinutes *int `json:"utc_offset_minutes,omitempty"`
}
type TemporalDuration struct {
	Days        int `json:"days"`
	Seconds     int `json:"seconds"`
	Nanoseconds int `json:"nanoseconds,omitempty"`
}
type RelativeDay string

const (
	RelativePreviousDay RelativeDay = "previous_day"
	RelativeCurrentDay  RelativeDay = "current_day"
	RelativeNextDay     RelativeDay = "next_day"
)

// TemporalPlan represents calendar/deictic meaning independently of tense.
// An omitted timezone or anchor remains unknown and is never supplied here.
type TemporalPlan struct {
	ID          string            `json:"id"`
	Kind        TemporalKind      `json:"kind"`
	Date        *CalendarDate     `json:"date,omitempty"`
	Clock       *ClockTime        `json:"clock_time,omitempty"`
	Duration    *TemporalDuration `json:"duration,omitempty"`
	RelativeDay RelativeDay       `json:"relative_day,omitempty"`
	Anchor      meaning.NodeID    `json:"anchor,omitempty"`
}
type CompiledTemporalPlan struct{ Plan TemporalPlan }

func CompileTemporal(plan TemporalPlan, profile QuantityTimeMWEProfile, featureProfile *features.LanguageFeatureProfile, graph meaning.Meaning) (CompiledTemporalPlan, error) {
	if err := ValidateQuantityTimeMWEProfile(profile, featureProfile); err != nil {
		return CompiledTemporalPlan{}, err
	}
	if !validToken(plan.ID) || !containsTemporalKind(profile.TemporalKinds, plan.Kind) {
		return CompiledTemporalPlan{}, unsupported("TEMPORAL_KIND_UNSUPPORTED", "kind", "temporal plan ID or kind is not supported by the language profile")
	}
	count := 0
	if plan.Date != nil {
		count++
	}
	if plan.Clock != nil {
		count++
	}
	if plan.Duration != nil {
		count++
	}
	if plan.RelativeDay != "" {
		count++
	}
	if count != 1 {
		return CompiledTemporalPlan{}, invalid("TEMPORAL_UNION_INVALID", "plan", "temporal plan must contain exactly one value matching its kind")
	}
	switch plan.Kind {
	case TemporalDateKind:
		if plan.Date == nil || plan.Clock != nil || plan.Duration != nil || plan.RelativeDay != "" || plan.Anchor != "" {
			return CompiledTemporalPlan{}, invalid("TEMPORAL_DATE_INVALID", "date", "calendar date contains incompatible fields")
		}
		d := plan.Date
		if d.Year < 1 || d.Year > 9999 || d.Month < 1 || d.Month > 12 || d.Day < 1 || d.Day > 31 {
			return CompiledTemporalPlan{}, invalid("INVALID_CALENDAR_DATE", "date", "calendar date is outside the Gregorian range")
		}
		parsed := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
		if parsed.Day() != d.Day || int(parsed.Month()) != d.Month || parsed.Year() != d.Year {
			return CompiledTemporalPlan{}, invalid("INVALID_CALENDAR_DATE", "date", "calendar date is outside the Gregorian range")
		}
	case TemporalClockKind:
		if plan.Clock == nil || plan.Date != nil || plan.Duration != nil || plan.RelativeDay != "" || plan.Anchor != "" {
			return CompiledTemporalPlan{}, invalid("TEMPORAL_CLOCK_INVALID", "clock_time", "clock time contains incompatible fields")
		}
		c := plan.Clock
		if c.Hour < 0 || c.Hour > 23 || c.Minute < 0 || c.Minute > 59 || c.Second < 0 || c.Second > 59 || c.Nanosecond < 0 || c.Nanosecond >= 1_000_000_000 {
			return CompiledTemporalPlan{}, invalid("INVALID_CLOCK_TIME", "clock_time", "clock component is outside its range")
		}
		if c.UTCOffsetMinutes != nil && (*c.UTCOffsetMinutes < -14*60 || *c.UTCOffsetMinutes > 14*60) {
			return CompiledTemporalPlan{}, invalid("INVALID_UTC_OFFSET", "clock_time.utc_offset_minutes", "UTC offset must be between -14:00 and +14:00")
		}
	case TemporalDurationKind:
		if plan.Duration == nil || plan.Date != nil || plan.Clock != nil || plan.RelativeDay != "" || plan.Anchor != "" {
			return CompiledTemporalPlan{}, invalid("TEMPORAL_DURATION_INVALID", "duration", "duration contains incompatible fields")
		}
		d := plan.Duration
		if d.Days < 0 || d.Days > MaxTemporalOffsetDays || d.Seconds < 0 || d.Seconds >= 86400 || d.Nanoseconds < 0 || d.Nanoseconds >= 1_000_000_000 {
			return CompiledTemporalPlan{}, invalid("INVALID_TEMPORAL_DURATION", "duration", "duration must be non-negative and bounded")
		}
	case TemporalRelativeDayKind:
		if plan.Date != nil || plan.Clock != nil || plan.Duration != nil || (plan.RelativeDay != RelativePreviousDay && plan.RelativeDay != RelativeCurrentDay && plan.RelativeDay != RelativeNextDay) {
			return CompiledTemporalPlan{}, invalid("INVALID_RELATIVE_DAY", "relative_day", "relative day must be explicit previous/current/next day")
		}
		if !validMeaningNodeID(plan.Anchor) {
			return CompiledTemporalPlan{}, incomplete("MISSING_DEICTIC_ANCHOR", "anchor", "deictic day requires an explicit Meaning anchor")
		}
		if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
			return CompiledTemporalPlan{}, err
		}
		found := false
		for _, node := range graph.Nodes {
			if node.ID == plan.Anchor {
				found = true
				break
			}
		}
		if !found {
			return CompiledTemporalPlan{}, unknown("UNKNOWN_DEICTIC_ANCHOR", "anchor", "deictic anchor node is absent from Meaning")
		}
	default:
		return CompiledTemporalPlan{}, invalid("INVALID_TEMPORAL_KIND", "kind", "temporal kind is unsupported")
	}
	return CompiledTemporalPlan{Plan: cloneTemporalPlan(plan)}, nil
}

func cloneTemporalPlan(plan TemporalPlan) TemporalPlan {
	result := plan
	if plan.Date != nil {
		date := *plan.Date
		result.Date = &date
	}
	if plan.Clock != nil {
		clock := *plan.Clock
		if plan.Clock.UTCOffsetMinutes != nil {
			offset := *plan.Clock.UTCOffsetMinutes
			clock.UTCOffsetMinutes = &offset
		}
		result.Clock = &clock
	}
	if plan.Duration != nil {
		duration := *plan.Duration
		result.Duration = &duration
	}
	return result
}

func containsTemporalKind(values []TemporalKind, value TemporalKind) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
func validMeaningNodeID(id meaning.NodeID) bool {
	value := string(id)
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

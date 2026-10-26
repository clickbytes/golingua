package synthesis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	DefaultSearchMaxPathLength     = 3
	MaxSearchMaxPathLength         = 16
	DefaultSearchMaxBranchDepth    = 64
	MaxSearchMaxBranchDepth        = 64
	DefaultSearchMaxExpansions     = 128
	MaxSearchMaxExpansions         = 4096
	DefaultSearchMaxCandidates     = 16
	MaxSearchMaxCandidates         = 128
	DefaultSearchMaxTraceEntries   = 64
	MaxSearchMaxTraceEntries       = 256
	DefaultSearchMaxWorkUnits      = 1_048_576
	MaxSearchMaxWorkUnits          = 16_777_216
	DefaultSearchMaxCandidateBytes = 8_388_608
	MaxSearchMaxCandidateBytes     = 16_777_216
	DefaultSearchTimeoutMillis     = 2_000
	MaxSearchTimeoutMillis         = 30_000
)

// SearchLimits bound the finite candidate exploration profile. Zero selects
// the documented finite default; it never means unlimited work. MaxCandidateBytes
// measures retained serialized candidate payloads, excluding the surrounding
// response envelope and trace; the CLI applies a separate whole-response cap.
type SearchLimits struct {
	MaxPathLength     int `json:"max_path_length,omitempty"`
	MaxBranchDepth    int `json:"max_branch_depth,omitempty"`
	MaxExpansions     int `json:"max_expansions,omitempty"`
	MaxCandidates     int `json:"max_candidates,omitempty"`
	MaxTraceEntries   int `json:"max_trace_entries,omitempty"`
	MaxWorkUnits      int `json:"max_work_units,omitempty"`
	MaxCandidateBytes int `json:"max_candidate_bytes,omitempty"`
	TimeoutMillis     int `json:"timeout_millis,omitempty"`
}

// NormalizeSearchLimits applies finite defaults and rejects negative or
// over-cap values before a search budget is allocated.
func NormalizeSearchLimits(requested SearchLimits) (SearchLimits, error) {
	fields := []struct {
		name     string
		value    *int
		fallback int
		maximum  int
	}{
		{"max_path_length", &requested.MaxPathLength, DefaultSearchMaxPathLength, MaxSearchMaxPathLength},
		{"max_branch_depth", &requested.MaxBranchDepth, DefaultSearchMaxBranchDepth, MaxSearchMaxBranchDepth},
		{"max_expansions", &requested.MaxExpansions, DefaultSearchMaxExpansions, MaxSearchMaxExpansions},
		{"max_candidates", &requested.MaxCandidates, DefaultSearchMaxCandidates, MaxSearchMaxCandidates},
		{"max_trace_entries", &requested.MaxTraceEntries, DefaultSearchMaxTraceEntries, MaxSearchMaxTraceEntries},
		{"max_work_units", &requested.MaxWorkUnits, DefaultSearchMaxWorkUnits, MaxSearchMaxWorkUnits},
		{"max_candidate_bytes", &requested.MaxCandidateBytes, DefaultSearchMaxCandidateBytes, MaxSearchMaxCandidateBytes},
		{"timeout_millis", &requested.TimeoutMillis, DefaultSearchTimeoutMillis, MaxSearchTimeoutMillis},
	}
	for _, field := range fields {
		if *field.value == 0 {
			*field.value = field.fallback
			continue
		}
		if *field.value < 0 || *field.value > field.maximum {
			return SearchLimits{}, &algebra.Error{
				Status: status.OutcomeInvalid, Code: "INVALID_SEARCH_LIMIT",
				Path:    "limits." + field.name,
				Message: fmt.Sprintf("must be zero (default) or between 1 and %d", field.maximum),
			}
		}
	}
	return requested, nil
}

// SearchStopReason records an operational bound that prevented a complete
// exploration. It is not a no-solution conclusion.
type SearchStopReason string

const (
	SearchStopTimeout           SearchStopReason = "timeout"
	SearchStopCancelled         SearchStopReason = "cancelled"
	SearchStopMaxExpansions     SearchStopReason = "max_expansions"
	SearchStopMaxCandidates     SearchStopReason = "max_candidates"
	SearchStopMaxWorkUnits      SearchStopReason = "max_work_units"
	SearchStopMaxCandidateBytes SearchStopReason = "max_candidate_bytes"
)

// SearchUsage contains request-local bounded counters. CandidateBytes counts
// serialized candidate payload bytes only, not the goal, space descriptor,
// trace, or whole response. Trace omission is reported separately and does
// not itself stop candidate exploration.
type SearchUsage struct {
	Expansions     int `json:"expansions"`
	WorkUnits      int `json:"work_units"`
	Candidates     int `json:"candidates"`
	CandidateBytes int `json:"candidate_bytes"`
	TraceEntries   int `json:"trace_entries"`
	TraceOmitted   int `json:"trace_omitted"`
}

// searchBudget is request-local state shared by one bounded search. Its
// counters are logical units, not elapsed CPU time or energy measurements.
type searchBudget struct {
	ctx        context.Context
	limits     SearchLimits
	usage      SearchUsage
	stopReason SearchStopReason
}

// newSearchBudget normalizes limits before inspecting or deriving the request
// context. The caller owns cancel and must call it when the search completes.
func newSearchBudget(parent context.Context, requested SearchLimits) (*searchBudget, context.CancelFunc, error) {
	limits, err := NormalizeSearchLimits(requested)
	if err != nil {
		return nil, nil, err
	}
	if parent == nil {
		return nil, nil, &algebra.Error{
			Status: status.OutcomeInvalid, Code: "INVALID_SEARCH_CONTEXT",
			Path: "context", Message: "search context is required",
		}
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(limits.TimeoutMillis)*time.Millisecond)
	return &searchBudget{ctx: ctx, limits: limits}, cancel, nil
}

// checkDeadlineCancel returns false after recording the first context stop.
func (budget *searchBudget) checkDeadlineCancel() bool {
	if budget == nil || budget.stopReason != "" {
		return false
	}
	if budget.ctx == nil {
		budget.stopReason = SearchStopCancelled
		return false
	}
	select {
	case <-budget.ctx.Done():
		if errors.Is(budget.ctx.Err(), context.DeadlineExceeded) {
			budget.stopReason = SearchStopTimeout
		} else {
			budget.stopReason = SearchStopCancelled
		}
		return false
	default:
		return true
	}
}

// consumeExpansion reserves one completed choice attempt before it runs.
func (budget *searchBudget) consumeExpansion() bool {
	if !budget.checkDeadlineCancel() {
		return false
	}
	if budget.usage.Expansions >= budget.limits.MaxExpansions {
		budget.stopReason = SearchStopMaxExpansions
		return false
	}
	budget.usage.Expansions++
	return true
}

// consumeWork charges nonnegative deterministic logical work without integer
// overflow. A false result means the search must stop with StopReason().
func (budget *searchBudget) consumeWork(units int) (bool, error) {
	if units < 0 {
		return false, &algebra.Error{
			Status: status.OutcomeInvalid, Code: "INVALID_SEARCH_WORK_CHARGE",
			Path: "work_units", Message: "work charge must be nonnegative",
		}
	}
	if !budget.checkDeadlineCancel() {
		return false, nil
	}
	if units > budget.limits.MaxWorkUnits-budget.usage.WorkUnits {
		budget.stopReason = SearchStopMaxWorkUnits
		return false, nil
	}
	budget.usage.WorkUnits += units
	return true, nil
}

// remainingWork returns the unused logical work allowance, clamped at zero.
func (budget *searchBudget) remainingWork() int {
	if budget == nil || budget.usage.WorkUnits >= budget.limits.MaxWorkUnits {
		return 0
	}
	return budget.limits.MaxWorkUnits - budget.usage.WorkUnits
}

// acceptCandidate reserves a candidate slot before the candidate is appended.
func (budget *searchBudget) acceptCandidate() bool {
	if !budget.checkDeadlineCancel() {
		return false
	}
	if budget.usage.Candidates >= budget.limits.MaxCandidates {
		budget.stopReason = SearchStopMaxCandidates
		return false
	}
	budget.usage.Candidates++
	return true
}

// consumeCandidateBytes accounts serialized candidate payload bytes before a
// candidate is retained. It excludes the response envelope and trace.
func (budget *searchBudget) consumeCandidateBytes(size int) (bool, error) {
	if size < 0 {
		return false, &algebra.Error{
			Status: status.OutcomeInvalid, Code: "INVALID_SEARCH_RESULT_SIZE",
			Path: "candidate_bytes", Message: "candidate byte charge must be nonnegative",
		}
	}
	if !budget.checkDeadlineCancel() {
		return false, nil
	}
	if size > budget.limits.MaxCandidateBytes-budget.usage.CandidateBytes {
		budget.stopReason = SearchStopMaxCandidateBytes
		return false, nil
	}
	budget.usage.CandidateBytes += size
	return true, nil
}

// retainTrace keeps a trace entry while capacity remains. Once full, it
// increments TraceOmitted and returns false without stopping the search.
func (budget *searchBudget) retainTrace() bool {
	if !budget.checkDeadlineCancel() {
		return false
	}
	if budget.usage.TraceEntries >= budget.limits.MaxTraceEntries {
		if budget.usage.TraceOmitted < int(^uint(0)>>1) {
			budget.usage.TraceOmitted++
		}
		return false
	}
	budget.usage.TraceEntries++
	return true
}

func (budget *searchBudget) stop() SearchStopReason {
	if budget != nil {
		return budget.stopReason
	}
	return SearchStopCancelled
}

func (budget *searchBudget) usageSnapshot() SearchUsage {
	if budget == nil {
		return SearchUsage{}
	}
	return budget.usage
}

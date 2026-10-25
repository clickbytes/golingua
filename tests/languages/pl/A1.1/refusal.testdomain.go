// Package a1pl11 contains owner-audit boundary refusal inputs.
package a1pl11

import "github.com/clickbytes/golingua/tests/matrix"

var RefusalCases = []matrix.RefusalCase{
	{ID: "REF-PL-001", Language: "pl", Input: "Nauczyciele widzi psa.", Requirement: "subject-verb agreement plural", InputOrigin: "owner_audit_2026-10-25", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-002", Language: "pl", Input: "Nauczyciel widzą psa.", Requirement: "subject-verb agreement singular", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-003", Language: "pl", Input: "Nauczyciel nie widzą psa.", Requirement: "negation does not repair agreement", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-004", Language: "pl", Input: "Nauczyciel widzi psa i", Requirement: "dangling conjunction with no second conjunct", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-005", Language: "pl", Input: "Nauczyciel widzi na psa.", Requirement: "wrong valency preposition", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-006", Language: "pl", Input: "Nauczyciel nie nie widzi psa.", Requirement: "duplicated negation without licensed metalinguistic context", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-007", Language: "pl", Input: "", Requirement: "empty input", InputOrigin: "coordinator_authored_boundary", AllowUnknownAnalysis: true, UnknownAllowanceReason: "Source lexical analysis reports an unknown empty lattice with explicit no-token diagnostics; no candidate Meaning is returned.", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-PL-008", Language: "pl", Input: "qzxv blorp.", Requirement: "unrecognized lexemes", InputOrigin: "coordinator_authored_boundary", AllowUnknownAnalysis: true, AllowUnknownTranslation: true, UnknownAllowanceReason: "Unknown lexical surfaces produce explicit UNKNOWN_SURFACE and NO_START_PLAN diagnostics; no candidate Meaning is returned.", OwnerAcceptance: "NOT_EVALUATED"},
}

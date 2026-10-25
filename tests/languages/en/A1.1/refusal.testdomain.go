// Package a1en11 contains owner-audit boundary refusal inputs.
package a1en11

import "github.com/clickbytes/golingua/tests/matrix"

var RefusalCases = []matrix.RefusalCase{
	{ID: "REF-EN-001", Language: "en", Input: "The teachers sees the dog.", Requirement: "subject-verb agreement plural", InputOrigin: "owner_audit_2026-10-25", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-002", Language: "en", Input: "The teacher see the dog.", Requirement: "subject-verb agreement singular", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-003", Language: "en", Input: "The teacher does not sees the dog.", Requirement: "auxiliary takes base verb", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-004", Language: "en", Input: "The teacher sees the dog and", Requirement: "dangling conjunction with no second conjunct", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-005", Language: "en", Input: "The teacher sees the.", Requirement: "incomplete noun phrase", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-006", Language: "en", Input: "The teacher not sees the dog.", Requirement: "missing do support", InputOrigin: "coordinator_authored_boundary", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-007", Language: "en", Input: "", Requirement: "empty input", InputOrigin: "coordinator_authored_boundary", AllowUnknownAnalysis: true, UnknownAllowanceReason: "Source lexical analysis reports an unknown empty lattice with explicit no-token diagnostics; no candidate Meaning is returned.", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "REF-EN-008", Language: "en", Input: "qzxv blorp.", Requirement: "unrecognized lexemes", InputOrigin: "coordinator_authored_boundary", AllowUnknownAnalysis: true, AllowUnknownTranslation: true, UnknownAllowanceReason: "Unknown lexical surfaces produce explicit UNKNOWN_SURFACE and NO_START_PLAN diagnostics; no candidate Meaning is returned.", OwnerAcceptance: "NOT_EVALUATED"},
}

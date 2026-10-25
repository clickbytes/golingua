package a1en111_test
import (
 "testing"
 "github.com/clickbytes/golingua/internal/features"
 a1 "github.com/clickbytes/golingua/tests/languages/en/A1.11"
 "github.com/clickbytes/golingua/tests/matrix"
)
func TestCaseOrderAnalysis(t *testing.T) { matrix.RunAnalysis(t,features.LanguageID("en"),a1.CaseOrderAnalysisCases) }
func TestCaseOrderRealization(t *testing.T) { matrix.RunRealization(t,features.LanguageID("en"),a1.CaseOrderRealizationCases) }

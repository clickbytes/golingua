package a1pl111_test
import (
 "testing"
 "github.com/clickbytes/golingua/internal/features"
 a1 "github.com/clickbytes/golingua/tests/languages/pl/A1.11"
 "github.com/clickbytes/golingua/tests/matrix"
)
func TestCaseOrderAnalysis(t *testing.T) { matrix.RunAnalysis(t,features.LanguageID("pl"),a1.CaseOrderAnalysisCases) }
func TestCaseOrderRealization(t *testing.T) { matrix.RunRealization(t,features.LanguageID("pl"),a1.CaseOrderRealizationCases) }

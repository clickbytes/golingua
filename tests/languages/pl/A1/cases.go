package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var AnalysisCases = append(append(append(append(append(append([]matrix.AnalysisCase{}, ActionAnalysisCases...), IdentificationAnalysisCases...), LocationQuantityAnalysisCases...), NegationAnalysisCases...), PossessionAnalysisCases...), QuestionAnalysisCases...)

var RealizationCases = append(append(append(append(append(append([]matrix.RealizationCase{}, ActionRealizationCases...), IdentificationRealizationCases...), LocationQuantityRealizationCases...), NegationRealizationCases...), PossessionRealizationCases...), QuestionRealizationCases...)

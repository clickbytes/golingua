package importer

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
)

// ParadigmProposal groups only observed forms and mapped feature signatures.
// Its fingerprint is a review aid, not an executable rule or approved class.
type ParadigmProposal struct {
	ID                  string                  `json:"id"`
	Language            features.LanguageID     `json:"language"`
	Lemma               string                  `json:"lemma"`
	POS                 lexicon.PartOfSpeech    `json:"pos,omitempty"`
	ClassFingerprint    string                  `json:"class_fingerprint"`
	ObservedForms       []ObservedUniMorphForm  `json:"observed_forms"`
	StemCandidates      []ObservedStemCandidate `json:"stem_candidates"`
	PotentialExceptions []ObservedUniMorphForm  `json:"potential_exceptions"`
	UnmappedTags        []string                `json:"unmapped_tags,omitempty"`
	ReviewStatus        string                  `json:"review_status"`
	Activation          string                  `json:"activation"`
}

type ObservedUniMorphForm struct {
	Line     int                    `json:"line"`
	Surface  string                 `json:"surface"`
	Features features.FeatureBundle `json:"features"`
	Tags     []string               `json:"tags"`
}

type ObservedStemCandidate struct {
	Form   string `json:"form"`
	Reason string `json:"reason"`
}

// BuildParadigmProposals creates deterministic, non-executable review records.
// It does not infer stems, transformations, exceptions, or missing forms.
func BuildParadigmProposals(records []UniMorphRecord) []ParadigmProposal {
	type group struct {
		language features.LanguageID
		lemma    string
		pos      lexicon.PartOfSpeech
		records  []UniMorphRecord
	}
	groups := make(map[string]*group)
	for _, record := range records {
		key := string(record.Features.Language) + "\x00" + record.Lemma + "\x00" + string(record.POS)
		if groups[key] == nil {
			groups[key] = &group{language: record.Features.Language, lemma: record.Lemma, pos: record.POS}
		}
		groups[key].records = append(groups[key].records, record)
	}
	result := make([]ParadigmProposal, 0, len(groups))
	for _, item := range groups {
		forms := make([]ObservedUniMorphForm, 0, len(item.records))
		stems := make([]ObservedStemCandidate, 0, 1)
		unmappedSet, patternSet := map[string]struct{}{}, map[string]struct{}{}
		for _, record := range item.records {
			forms = append(forms, ObservedUniMorphForm{Line: record.Line, Surface: record.Surface, Features: record.Features, Tags: append([]string(nil), record.Tags...)})
			if record.Surface == item.lemma && len(stems) == 0 {
				stems = append(stems, ObservedStemCandidate{Form: record.Surface, Reason: "observed_exact_lemma_form"})
			}
			for _, tag := range record.UnmappedTags {
				unmappedSet[tag] = struct{}{}
			}
			patternSet[featureKey(record.Features)] = struct{}{}
		}
		sort.Slice(forms, func(i, j int) bool {
			if forms[i].Surface != forms[j].Surface {
				return forms[i].Surface < forms[j].Surface
			}
			if featureKey(forms[i].Features) != featureKey(forms[j].Features) {
				return featureKey(forms[i].Features) < featureKey(forms[j].Features)
			}
			return strings.Join(forms[i].Tags, ";") < strings.Join(forms[j].Tags, ";")
		})
		patterns := make([]string, 0, len(patternSet))
		for pattern := range patternSet {
			patterns = append(patterns, pattern)
		}
		sort.Strings(patterns)
		fingerprint := sha256.Sum256([]byte(joinStrings(patterns, "\n")))
		unmapped := make([]string, 0, len(unmappedSet))
		for tag := range unmappedSet {
			unmapped = append(unmapped, tag)
		}
		sort.Strings(unmapped)
		identity := string(item.language) + "\x00" + string(item.pos) + "\x00" + item.lemma + "\x00" + hex.EncodeToString(fingerprint[:])
		idDigest := sha256.Sum256([]byte(identity))
		result = append(result, ParadigmProposal{
			ID:       "unimorph-candidate-" + hex.EncodeToString(idDigest[:]),
			Language: item.language, Lemma: item.lemma, POS: item.pos,
			ClassFingerprint: hex.EncodeToString(fingerprint[:]), ObservedForms: forms,
			StemCandidates: stems, PotentialExceptions: []ObservedUniMorphForm{},
			UnmappedTags: unmapped, ReviewStatus: "needs_review", Activation: "not_activated",
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func joinStrings(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += separator + value
	}
	return result
}

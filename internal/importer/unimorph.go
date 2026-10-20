package importer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	UniMorphProfileSchemaVersion = 1
	MaxUniMorphProfileBytes      = 1 << 20
	MaxUniMorphInputBytes        = 64 << 20
	MaxUniMorphLineBytes         = 4 << 10
	MaxUniMorphRecords           = 100_000
	MaxUniMorphLemmaBytes        = 256
	MaxUniMorphSurfaceBytes      = 256
	MaxUniMorphTagsPerRecord     = 64
	MaxUniMorphOutputBytes       = 128 << 20
)

// UniMorphProfile maps only explicitly reviewed source tags to one language's
// existing POS and feature inventory.
type UniMorphProfile struct {
	SchemaVersion  int                     `json:"schema_version"`
	Language       features.LanguageID     `json:"language"`
	ProfileVersion features.ProfileVersion `json:"profile_version"`
	Mappings       []UniMorphTagMapping    `json:"mappings"`
}

type UniMorphTagMapping struct {
	Tag     string                  `json:"tag"`
	WhenPOS lexicon.PartOfSpeech    `json:"when_pos,omitempty"`
	POS     lexicon.PartOfSpeech    `json:"pos,omitempty"`
	Feature features.FeatureID      `json:"feature,omitempty"`
	Value   features.FeatureValueID `json:"value,omitempty"`
}

type UniMorphRecord struct {
	Line         int                    `json:"line"`
	Lemma        string                 `json:"lemma"`
	Surface      string                 `json:"surface"`
	POS          lexicon.PartOfSpeech   `json:"pos,omitempty"`
	Features     features.FeatureBundle `json:"features"`
	Tags         []string               `json:"tags"`
	UnmappedTags []string               `json:"unmapped_tags,omitempty"`
	Provenance   provenance.SourceRef   `json:"provenance"`
}

type UniMorphImport struct {
	Language   features.LanguageID `json:"language"`
	Records    []UniMorphRecord    `json:"records"`
	Proposals  []ParadigmProposal  `json:"proposals"`
	Duplicates int                 `json:"duplicates"`
}

// DecodeUniMorphProfile validates a strict profile against the language's
// existing feature inventory. It never normalizes or guesses tag spellings.
func DecodeUniMorphProfile(data []byte, featureProfile *features.LanguageFeatureProfile) (UniMorphProfile, error) {
	if len(data) == 0 {
		return UniMorphProfile{}, invalid("EMPTY_UNIMORPH_PROFILE", "profile", "UniMorph profile must not be empty")
	}
	if len(data) > MaxUniMorphProfileBytes {
		return UniMorphProfile{}, resourceLimit("unimorph_profile_bytes", len(data), MaxUniMorphProfileBytes)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return UniMorphProfile{}, invalid("INVALID_UNIMORPH_PROFILE", "profile", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var profile UniMorphProfile
	if err := decoder.Decode(&profile); err != nil {
		return UniMorphProfile{}, invalid("INVALID_UNIMORPH_PROFILE", "profile", err.Error())
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return UniMorphProfile{}, invalid("TRAILING_UNIMORPH_PROFILE_DATA", "profile", "profile must contain exactly one JSON value")
	}
	if err := profile.Validate(featureProfile); err != nil {
		return UniMorphProfile{}, err
	}
	profile.Mappings = append([]UniMorphTagMapping(nil), profile.Mappings...)
	return profile, nil
}

func (profile UniMorphProfile) Validate(featureProfile *features.LanguageFeatureProfile) error {
	if profile.SchemaVersion != UniMorphProfileSchemaVersion || featureProfile == nil || profile.Language != featureProfile.Language() || profile.ProfileVersion != featureProfile.Version() {
		return invalid("UNIMORPH_PROFILE_MISMATCH", "profile", "UniMorph and language feature profile version/language must match")
	}
	if len(profile.Mappings) == 0 || len(profile.Mappings) > features.MaxDimensions*features.MaxValuesPerDim {
		return invalid("INVALID_UNIMORPH_MAPPING_COUNT", "profile.mappings", "profile must contain a bounded non-empty tag mapping list")
	}
	seen := make(map[string]struct{}, len(profile.Mappings))
	for index, mapping := range profile.Mappings {
		path := fmt.Sprintf("profile.mappings[%d]", index)
		if !validUniMorphTag(mapping.Tag) {
			return invalid("INVALID_UNIMORPH_TAG", path+".tag", "tag must be a bounded uppercase UniMorph token")
		}
		key := mapping.Tag + "\x00" + string(mapping.WhenPOS) + "\x00" + string(mapping.POS) + "\x00" + string(mapping.Feature)
		if _, exists := seen[key]; exists {
			return invalid("DUPLICATE_UNIMORPH_TAG", path+".tag", "source tag mapping is duplicated")
		}
		seen[key] = struct{}{}
		if mapping.WhenPOS != "" && !validPOS(string(mapping.WhenPOS)) {
			return invalid("INVALID_UNIMORPH_MAPPING", path+".when_pos", "mapping context POS is malformed")
		}
		if mapping.POS != "" && mapping.WhenPOS != "" {
			return invalid("INVALID_UNIMORPH_MAPPING", path+".when_pos", "POS declarations cannot themselves be conditioned on another POS")
		}
		if mapping.POS == "" && (mapping.Feature == "" || mapping.Value == "") || mapping.POS != "" && !validPOS(string(mapping.POS)) {
			return invalid("INVALID_UNIMORPH_MAPPING", path, "mapping must declare a valid POS and/or a complete feature/value pair")
		}
		if (mapping.Feature == "") != (mapping.Value == "") {
			return invalid("INVALID_UNIMORPH_MAPPING", path, "feature and value must be declared together")
		}
		if mapping.Feature != "" {
			dimension, exists := featureProfile.Lookup(mapping.Feature)
			if !exists || !containsFeatureValue(dimension.Values, mapping.Value) {
				return invalid("UNSUPPORTED_UNIMORPH_FEATURE", path, "mapped feature/value is not present in the language feature profile")
			}
		}
	}
	return nil
}

// ReadUniMorph parses the standard three-column lemma/form/tags TSV as a
// stream. Identical rows are deduplicated; unmapped tags stay visible in output.
func ReadUniMorph(input io.Reader, profile UniMorphProfile, featureProfile *features.LanguageFeatureProfile, source provenance.SourceRef) (UniMorphImport, error) {
	if err := profile.Validate(featureProfile); err != nil {
		return UniMorphImport{}, err
	}
	if err := source.Validate(); err != nil {
		return UniMorphImport{}, invalid("INVALID_UNIMORPH_SOURCE", "source", err.Error())
	}
	if input == nil {
		return UniMorphImport{}, invalid("MISSING_UNIMORPH_INPUT", "input", "TSV input reader is required")
	}
	limited := &countingReader{reader: io.LimitReader(input, MaxUniMorphInputBytes+1)}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 1024), MaxUniMorphLineBytes+1)
	tagMappings := make(map[string][]UniMorphTagMapping, len(profile.Mappings))
	for _, mapping := range profile.Mappings {
		tagMappings[mapping.Tag] = append(tagMappings[mapping.Tag], mapping)
	}
	records := make([]UniMorphRecord, 0)
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSuffix(scanner.Text(), "\r")
		if len(raw) > MaxUniMorphLineBytes {
			return UniMorphImport{}, resourceLimit("unimorph_line_bytes", len(raw), MaxUniMorphLineBytes)
		}
		if raw == "" {
			continue
		}
		if line > MaxUniMorphRecords {
			return UniMorphImport{}, resourceLimit("unimorph_records", line, MaxUniMorphRecords)
		}
		if !utf8.ValidString(raw) {
			return UniMorphImport{}, invalid("INVALID_UNIMORPH_UTF8", fmt.Sprintf("lines[%d]", line), "TSV row must be valid UTF-8")
		}
		columns := strings.Split(raw, "\t")
		if len(columns) != 3 {
			return UniMorphImport{}, invalid("INVALID_UNIMORPH_ROW", fmt.Sprintf("lines[%d]", line), "TSV row must have exactly lemma, surface, and semicolon-separated tags columns")
		}
		if columns[0] == "lemma" && columns[1] == "form" && columns[2] == "tags" {
			return UniMorphImport{}, invalid("UNIMORPH_HEADER_UNSUPPORTED", fmt.Sprintf("lines[%d]", line), "TSV header rows are not part of the three-column data format")
		}
		lemma, surface := columns[0], columns[1]
		if !validUniMorphText(lemma, MaxUniMorphLemmaBytes) || !validUniMorphText(surface, MaxUniMorphSurfaceBytes) {
			return UniMorphImport{}, invalid("INVALID_UNIMORPH_FORM", fmt.Sprintf("lines[%d]", line), "lemma and surface must be non-empty, trimmed, control-free, UTF-8 fields within limits")
		}
		rawTags := strings.Split(columns[2], ";")
		if len(rawTags) == 0 || len(rawTags) > MaxUniMorphTagsPerRecord {
			return UniMorphImport{}, resourceLimit("unimorph_tags_per_record", len(rawTags), MaxUniMorphTagsPerRecord)
		}
		featuresOut := make([]features.FeatureValue, 0, len(rawTags))
		tags := make([]string, len(rawTags))
		unmapped := make([]string, 0)
		var pos lexicon.PartOfSpeech
		seenTags, seenFeatures := map[string]struct{}{}, map[features.FeatureID]struct{}{}
		posMappedTags := make(map[string]struct{})
		for _, tag := range rawTags {
			for _, mapping := range tagMappings[tag] {
				if mapping.POS == "" || mapping.WhenPOS != "" {
					continue
				}
				posMappedTags[tag] = struct{}{}
				if pos != "" && pos != mapping.POS {
					return UniMorphImport{}, invalid("CONFLICTING_UNIMORPH_POS", fmt.Sprintf("lines[%d].tags", line), "row maps to conflicting parts of speech")
				}
				pos = mapping.POS
			}
		}
		for index, tag := range rawTags {
			if !validUniMorphTag(tag) {
				return UniMorphImport{}, invalid("INVALID_UNIMORPH_TAG", fmt.Sprintf("lines[%d].tags[%d]", line, index), "tag is empty or malformed")
			}
			if _, exists := seenTags[tag]; exists {
				return UniMorphImport{}, invalid("DUPLICATE_UNIMORPH_TAG", fmt.Sprintf("lines[%d].tags[%d]", line, index), "tag is repeated in the same row")
			}
			seenTags[tag] = struct{}{}
			tags[index] = tag
			mappings, exists := tagMappings[tag]
			if !exists {
				unmapped = append(unmapped, tag)
				continue
			}
			_, applied := posMappedTags[tag]
			for _, mapping := range mappings {
				if mapping.WhenPOS != "" && mapping.WhenPOS != pos {
					continue
				}
				if mapping.Feature == "" {
					continue
				}
				applied = true
				if _, exists := seenFeatures[mapping.Feature]; exists {
					return UniMorphImport{}, invalid("CONFLICTING_UNIMORPH_FEATURE", fmt.Sprintf("lines[%d].tags", line), "row maps multiple tags to the same feature dimension")
				}
				seenFeatures[mapping.Feature] = struct{}{}
				featuresOut = append(featuresOut, features.FeatureValue{Feature: mapping.Feature, State: features.StateKnown, Value: mapping.Value})
			}
			if !applied {
				unmapped = append(unmapped, tag)
			}
		}
		bundle, err := features.NewBundle(featureProfile, featuresOut)
		if err != nil {
			return UniMorphImport{}, invalid("INVALID_UNIMORPH_FEATURE_BUNDLE", fmt.Sprintf("lines[%d].tags", line), err.Error())
		}
		sort.Strings(tags)
		sort.Strings(unmapped)
		record := UniMorphRecord{Line: line, Lemma: lemma, Surface: surface, POS: pos, Features: bundle, Tags: tags, UnmappedTags: unmapped, Provenance: source}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		if limited.count >= MaxUniMorphInputBytes {
			return UniMorphImport{}, resourceLimit("unimorph_input_bytes", int(limited.count), MaxUniMorphInputBytes)
		}
		if strings.HasPrefix(err.Error(), "bufio.Scanner: token too long") {
			return UniMorphImport{}, resourceLimit("unimorph_line_bytes", MaxUniMorphLineBytes+1, MaxUniMorphLineBytes)
		}
		return UniMorphImport{}, invalid("INVALID_UNIMORPH_TSV", "input", err.Error())
	}
	if limited.count > MaxUniMorphInputBytes {
		return UniMorphImport{}, resourceLimit("unimorph_input_bytes", int(limited.count), MaxUniMorphInputBytes)
	}
	if len(records) == 0 {
		return UniMorphImport{}, invalid("EMPTY_UNIMORPH_TSV", "input", "TSV has no data records")
	}
	records, duplicates, err := DeduplicateUniMorph(records)
	if err != nil {
		return UniMorphImport{}, invalid("INVALID_UNIMORPH_RECORD", "input", err.Error())
	}
	result := UniMorphImport{Language: profile.Language, Records: records, Proposals: BuildParadigmProposals(records), Duplicates: duplicates}
	return result, nil
}

type countingReader struct {
	reader io.Reader
	count  int64
}

func (reader *countingReader) Read(data []byte) (int, error) {
	n, err := reader.reader.Read(data)
	reader.count += int64(n)
	return n, err
}

func validPOS(value string) bool {
	if len(value) == 0 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, char := range value {
		if !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '_' && char != '-' {
			return false
		}
	}
	return true
}

func validUniMorphTag(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if !(char >= 'A' && char <= 'Z') && !(char >= '0' && char <= '9') && char != '.' && char != '_' && char != '-' {
			return false
		}
	}
	return true
}

func validUniMorphText(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
}

func containsFeatureValue(values []features.FeatureValueID, target features.FeatureValueID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func uniMorphRecordKey(record UniMorphRecord) (string, error) {
	keyRecord := struct {
		Lemma, Surface string
		POS            lexicon.PartOfSpeech
		Tags           []string
	}{record.Lemma, record.Surface, record.POS, record.Tags}
	encoded, err := json.Marshal(keyRecord)
	return string(encoded), err
}

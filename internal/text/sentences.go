package text

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/status"
)

// Sentence is a source-preserving sentence candidate. Span is a half-open
// UTF-8 byte range into the original input; Text is exactly input[Span.Start:Span.End].
// Terminated is true only when a recognized sentence-final mark closed it.
// Recoverable and UnsupportedPunctuation report punctuation that was retained
// as source but was not interpreted as a sentence boundary.
type Sentence struct {
	Span                   Span
	Text                   string
	Terminated             bool
	Recoverable            bool
	UnsupportedPunctuation []Span
}

// SentenceRules adds language-profile sentence marks and exact abbreviations
// to the common Latin/CJK defaults. The fields are data, not regular
// expressions; callers should derive them from a trusted language profile.
type SentenceRules struct {
	AdditionalTerminators           string
	Abbreviations                   []string
	AdditionalRecognizedPunctuation string
}

// SegmentSentences finds conservative sentence boundaries in bounded UTF-8
// input. It recognizes common Latin and CJK sentence-final marks. Other
// punctuation remains in the exact source segment; unfamiliar punctuation
// marks are listed as recoverable evidence instead of becoming an error.
// A final nonempty segment without a recognized terminator is returned with
// Terminated=false. Whitespace between and around segments is retained in the
// original input, but excluded from their spans.
func SegmentSentences(input string) ([]Sentence, error) {
	return SegmentSentencesWithRules(input, SentenceRules{})
}

// SegmentSentencesWithRules segments text using bounded language-profile
// additions to the default terminal punctuation set.
func SegmentSentencesWithRules(input string, rules SentenceRules) ([]Sentence, error) {
	if err := validateSentenceRules(rules); err != nil {
		return nil, err
	}
	if len(input) > status.DefaultBounds().MaxInputBytes {
		return nil, &TextError{
			Kind:   TextErrorResourceLimit,
			Limit:  status.DefaultBounds().MaxInputBytes,
			Actual: len(input),
			Detail: "sentence segmentation input bytes",
		}
	}
	if _, err := Normalize(input); err != nil {
		return nil, err
	}

	segments := make([]Sentence, 0, 1)
	segmentStart := 0
	for offset := 0; offset < len(input); {
		r, size := utf8.DecodeRuneInString(input[offset:])
		if !isSentenceTerminator(r, rules) {
			offset += size
			continue
		}

		end := consumeSentenceEnding(input, offset+size, rules)
		// A punctuation mark inside a token (for example a decimal point) does
		// not end a sentence. At end of input it is unambiguous; otherwise a
		// boundary requires whitespace or a closing quote followed by whitespace.
		if end < len(input) && !isSpaceAt(input, end) {
			offset += size
			continue
		}
		if abbreviationAt(input, segmentStart, offset+size, end, rules.Abbreviations) {
			offset += size
			continue
		}
		if sentence, ok := makeSentence(input, segmentStart, end, true, rules); ok {
			segments = append(segments, sentence)
		}
		segmentStart = end
		offset = end
	}
	if sentence, ok := makeSentence(input, segmentStart, len(input), false, rules); ok {
		segments = append(segments, sentence)
	}
	return segments, nil
}

func makeSentence(input string, start, end int, terminated bool, rules SentenceRules) (Sentence, bool) {
	for start < end {
		r, size := utf8.DecodeRuneInString(input[start:end])
		if !unicode.IsSpace(r) {
			break
		}
		start += size
	}
	for start < end {
		r, size := utf8.DecodeLastRuneInString(input[start:end])
		if !unicode.IsSpace(r) {
			break
		}
		end -= size
	}
	if start == end {
		return Sentence{}, false
	}

	result := Sentence{
		Span:       Span{Start: start, End: end},
		Text:       input[start:end],
		Terminated: terminated,
	}
	for offset := start; offset < end; {
		r, size := utf8.DecodeRuneInString(input[offset:end])
		if isUnrecognizedPunctuation(r, rules) {
			result.UnsupportedPunctuation = append(result.UnsupportedPunctuation, Span{Start: offset, End: offset + size})
		}
		offset += size
	}
	result.Recoverable = len(result.UnsupportedPunctuation) > 0
	return result, true
}

func isSentenceTerminator(r rune, rules SentenceRules) bool {
	switch r {
	case '.', '?', '!', '…', '。', '？', '！':
		return true
	default:
		return strings.ContainsRune(rules.AdditionalTerminators, r)
	}
}

// consumeSentenceEnding includes adjacent terminal marks and common closing
// quotation/bracket marks so the complete ending remains within source evidence.
func consumeSentenceEnding(input string, offset int, rules SentenceRules) int {
	for offset < len(input) {
		r, size := utf8.DecodeRuneInString(input[offset:])
		if isSentenceTerminator(r, rules) || isClosingMark(r) {
			offset += size
			continue
		}
		break
	}
	return offset
}

func isClosingMark(r rune) bool {
	switch r {
	case '"', '\'', '”', '’', '»', '›', ')', ']', '}', '」', '』', '】', '〉', '》':
		return true
	default:
		return false
	}
}

func isSpaceAt(input string, offset int) bool {
	if offset >= len(input) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(input[offset:])
	return unicode.IsSpace(r)
}

func isUnrecognizedPunctuation(r rune, rules SentenceRules) bool {
	if !unicode.IsPunct(r) || isSentenceTerminator(r, rules) || strings.ContainsRune(rules.AdditionalRecognizedPunctuation, r) {
		return false
	}
	// These marks are ordinary in-sentence separators or delimiters. They are
	// retained as evidence and do not need a recoverable boundary diagnostic.
	if strings.ContainsRune(",;:—–-‐‑()[]{}«»‹›\"'“”‘’/\\", r) {
		return false
	}
	return true
}

func validateSentenceRules(rules SentenceRules) error {
	const (
		maxRuleBytes            = 4096
		maxPunctuationRuleBytes = 256
		maxAbbreviations        = 128
		maxAbbreviationLen      = 64
	)
	if len(rules.AdditionalTerminators) > maxPunctuationRuleBytes {
		return &TextError{Kind: TextErrorResourceLimit, Limit: maxPunctuationRuleBytes, Actual: len(rules.AdditionalTerminators), Detail: "sentence terminator rule bytes"}
	}
	if len(rules.AdditionalRecognizedPunctuation) > maxPunctuationRuleBytes {
		return &TextError{Kind: TextErrorResourceLimit, Limit: maxPunctuationRuleBytes, Actual: len(rules.AdditionalRecognizedPunctuation), Detail: "recognized punctuation rule bytes"}
	}
	actual := len(rules.AdditionalTerminators) + len(rules.AdditionalRecognizedPunctuation)
	if len(rules.Abbreviations) > maxAbbreviations {
		return &TextError{Kind: TextErrorResourceLimit, Limit: maxAbbreviations, Actual: len(rules.Abbreviations), Detail: "sentence abbreviation rule count"}
	}
	for _, abbreviation := range rules.Abbreviations {
		if len(abbreviation) > maxAbbreviationLen {
			return &TextError{Kind: TextErrorResourceLimit, Limit: maxAbbreviationLen, Actual: len(abbreviation), Detail: "sentence abbreviation rule bytes"}
		}
		actual += len(abbreviation)
	}
	if actual > maxRuleBytes {
		return &TextError{Kind: TextErrorResourceLimit, Limit: maxRuleBytes, Actual: actual, Detail: "sentence rule bytes"}
	}
	if !utf8.ValidString(rules.AdditionalTerminators) || !utf8.ValidString(rules.AdditionalRecognizedPunctuation) {
		return &TextError{Kind: TextErrorInvalidProfile, Detail: "sentence punctuation rules must be valid UTF-8"}
	}
	for _, abbreviation := range rules.Abbreviations {
		if !utf8.ValidString(abbreviation) {
			return &TextError{Kind: TextErrorInvalidProfile, Detail: "abbreviation rules must be valid UTF-8"}
		}
	}
	for _, mark := range rules.AdditionalTerminators + rules.AdditionalRecognizedPunctuation {
		if !unicode.IsPunct(mark) {
			return &TextError{Kind: TextErrorInvalidProfile, Detail: "additional sentence rules must contain punctuation characters"}
		}
	}
	return nil
}

func abbreviationAt(input string, sentenceStart, markEnd, segmentEnd int, abbreviations []string) bool {
	if len(abbreviations) == 0 || segmentEnd >= len(input) || !hasFollowingText(input, segmentEnd) {
		return false
	}
	_, markSize := utf8.DecodeLastRuneInString(input[:markEnd])
	markStart := markEnd - markSize
	earliest := max(sentenceStart, markStart-64)
	for start := markStart; start >= earliest; {
		if isTokenBoundaryBefore(input, sentenceStart, start) {
			prefix := input[start:markEnd]
			for _, abbreviation := range abbreviations {
				if abbreviation == "" || len(abbreviation) < len(prefix) || !strings.HasPrefix(abbreviation, prefix) {
					continue
				}
				end := start + len(abbreviation)
				if end > len(input) || input[start:end] != abbreviation || !isBoundaryAfterAbbreviation(input, end) {
					continue
				}
				if hasFollowingText(input, end) {
					return true
				}
			}
		}
		if start == earliest {
			break
		}
		_, previousSize := utf8.DecodeLastRuneInString(input[:start])
		start -= previousSize
	}
	return false
}

func isBoundaryAfterAbbreviation(input string, offset int) bool {
	if offset >= len(input) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(input[offset:])
	return unicode.IsSpace(r) || unicode.IsPunct(r)
}

func isTokenBoundaryBefore(input string, lowerBound, offset int) bool {
	if offset == lowerBound {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(input[lowerBound:offset])
	return unicode.IsSpace(r) || unicode.IsPunct(r)
}

func hasFollowingText(input string, offset int) bool {
	for offset < len(input) {
		r, size := utf8.DecodeRuneInString(input[offset:])
		if !unicode.IsSpace(r) {
			return true
		}
		offset += size
	}
	return false
}

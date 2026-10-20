// Package text provides source-preserving text operations.
package text

import (
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/status"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// TokenKind distinguishes lexical-looking spans from punctuation. The
// tokenizer does not assign a lexical or semantic interpretation to either.
type TokenKind string

const (
	TokenWord        TokenKind = "word"
	TokenPunctuation TokenKind = "punctuation"
	TokenOther       TokenKind = "other"
)

// Token records a contiguous slice of the original input. Span is half-open
// and contains UTF-8 byte offsets, not rune indexes.
type Token struct {
	Surface    string    `json:"surface"`
	Normalized string    `json:"normalized"`
	Folded     string    `json:"folded,omitempty"`
	Span       Span      `json:"span"`
	Kind       TokenKind `json:"kind"`
}

// TokenizerProfile supplies language-specific lexical boundaries without
// placing language rules in shared semantic code. WordRune may extend the
// default Unicode letter/number/mark set. Apostrophes are accepted inside a
// word only when surrounded by letters.
type TokenizerProfile struct {
	WordRune        func(rune) bool
	PunctuationRune func(rune) bool
	Apostrophes     string
	CaseFold        bool
}

// TokenizeOptions bounds input and output. Zero values select the package
// defaults. A profile may add language-specific word characters or apostrophe
// conventions.
type TokenizeOptions struct {
	MaxInputBytes int
	MaxTokens     int
	Apostrophes   string
	Profile       *TokenizerProfile
}

const (
	defaultMaxTokens   = 1 << 16
	defaultApostrophes = "'’ʼ"
)

// Tokenize produces word and punctuation spans over the exact original input.
// Whitespace is retained implicitly in the gaps between spans. Combining
// marks remain attached to the preceding word when possible, and apostrophes
// inside letter sequences stay within that word. Multiword lexical matching
// can be performed over adjacent word tokens without changing their spans.
func Tokenize(input string, options TokenizeOptions) ([]Token, error) {
	maxInputBytes := options.MaxInputBytes
	if maxInputBytes == 0 {
		maxInputBytes = status.DefaultBounds().MaxInputBytes
	} else if maxInputBytes < 0 {
		return nil, &TextError{Kind: TextErrorInvalidLimit, Limit: maxInputBytes, Detail: "maximum input bytes must be positive"}
	}
	if _, err := NormalizeWithLimit(input, maxInputBytes); err != nil {
		return nil, err
	}

	maxTokens := options.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	} else if maxTokens < 0 {
		return nil, &TextError{Kind: TextErrorInvalidLimit, Limit: maxTokens, Detail: "maximum token count must be positive"}
	}
	apostrophes := options.Apostrophes
	if options.Profile != nil {
		if apostrophes == "" {
			apostrophes = options.Profile.Apostrophes
		}
	} else if apostrophes == "" {
		apostrophes = defaultApostrophes
	}
	if len(apostrophes) > 64 || !utf8.ValidString(apostrophes) {
		if len(apostrophes) > 64 {
			return nil, &TextError{Kind: TextErrorResourceLimit, Limit: 64, Actual: len(apostrophes), Detail: "apostrophe profile bytes"}
		}
		return nil, &TextError{Kind: TextErrorInvalidProfile, Detail: "apostrophe set must be valid UTF-8"}
	}
	var wordRune func(rune) bool
	var punctuationRune func(rune) bool
	if options.Profile != nil {
		wordRune = options.Profile.WordRune
		punctuationRune = options.Profile.PunctuationRune
	}
	if wordRune == nil {
		wordRune = isWordRune
	}
	if punctuationRune == nil {
		punctuationRune = unicode.IsPunct
	}
	caseFold := options.Profile != nil && options.Profile.CaseFold
	var caseFolder cases.Caser
	if caseFold {
		caseFolder = cases.Fold()
	}

	tokens := make([]Token, 0, min(len(input), maxTokens))
	for offset := 0; offset < len(input); {
		r, size := utf8.DecodeRuneInString(input[offset:])
		if unicode.IsSpace(r) {
			offset += size
			continue
		}

		if isWordStart(r, wordRune) {
			start := offset
			offset += size
			for offset < len(input) {
				next, nextSize := utf8.DecodeRuneInString(input[offset:])
				if isWordContinuation(next, wordRune) {
					offset += nextSize
					continue
				}
				if stringsContainsRune(apostrophes, next) && offset > start && offset+nextSize < len(input) {
					following, _ := utf8.DecodeRuneInString(input[offset+nextSize:])
					if endsInLetter(input[start:offset]) && unicode.IsLetter(following) {
						offset += nextSize
						continue
					}
				}
				break
			}
			surface := input[start:offset]
			normalized := norm.NFC.String(surface)
			token := Token{Surface: surface, Normalized: normalized, Span: Span{Start: start, End: offset}, Kind: TokenWord}
			if caseFold {
				token.Folded = caseFolder.String(normalized)
			}
			tokens = append(tokens, token)
		} else {
			start := offset
			offset += size
			surface := input[start:offset]
			kind := TokenOther
			if punctuationRune(r) {
				kind = TokenPunctuation
			}
			normalized := norm.NFC.String(surface)
			token := Token{Surface: surface, Normalized: normalized, Span: Span{Start: start, End: offset}, Kind: kind}
			if caseFold {
				token.Folded = caseFolder.String(normalized)
			}
			tokens = append(tokens, token)
		}
		if len(tokens) > maxTokens {
			return nil, &TextError{Kind: TextErrorResourceLimit, Limit: maxTokens, Actual: len(tokens), Detail: "tokenizer token count"}
		}
	}
	return tokens, nil
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r)
}

func isWordStart(r rune, wordRune func(rune) bool) bool {
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return true
	}
	return wordRune != nil && wordRune(r) && !unicode.IsMark(r)
}

func isWordContinuation(r rune, wordRune func(rune) bool) bool {
	return unicode.IsMark(r) || wordRune != nil && wordRune(r)
}

func stringsContainsRune(s string, target rune) bool {
	for _, r := range s {
		if r == target {
			return true
		}
	}
	return false
}

func endsInLetter(s string) bool {
	for i := len(s); i > 0; {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		i -= size
		if unicode.IsMark(r) {
			continue
		}
		return unicode.IsLetter(r)
	}
	return false
}

// Package text provides source-preserving text normalization helpers.
package text

import (
	"fmt"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/status"
	"golang.org/x/text/unicode/norm"
)

// Span is a half-open range of UTF-8 byte offsets.
type Span struct {
	Start int
	End   int
}

// Cluster pairs one NFC normalization segment with the bytes that produced it.
// Both spans are half-open UTF-8 byte offsets, in NFC and Raw respectively.
type Cluster struct {
	NFC    Span
	Source Span
}

// Normalized retains the source bytes and their NFC matching representation.
// Clusters provide exact source mapping at normalization-segment boundaries.
type Normalized struct {
	Raw      string
	NFC      string
	Clusters []Cluster
}

// TextErrorKind identifies the class of a text input failure.
type TextErrorKind string

const (
	TextErrorInvalidUTF8    TextErrorKind = "invalid_utf8"
	TextErrorResourceLimit  TextErrorKind = "resource_limit"
	TextErrorInvalidLimit   TextErrorKind = "invalid_limit"
	TextErrorInvalidProfile TextErrorKind = "invalid_profile"
)

// TextError describes malformed input, invalid profile/limit configuration, or
// a bounded input/work refusal.
type TextError struct {
	Kind          TextErrorKind
	Offset        int
	Limit, Actual int
	Detail        string
}

func (e *TextError) Error() string {
	if e == nil {
		return "text input error"
	}
	switch e.Kind {
	case TextErrorInvalidUTF8:
		return fmt.Sprintf("invalid UTF-8 at byte offset %d", e.Offset)
	case TextErrorResourceLimit:
		if e.Detail != "" {
			return fmt.Sprintf("%s exceeds limit: %d > %d", e.Detail, e.Actual, e.Limit)
		}
		return fmt.Sprintf("text input exceeds byte limit: %d > %d", e.Actual, e.Limit)
	case TextErrorInvalidLimit:
		return e.Detail
	case TextErrorInvalidProfile:
		return "invalid text profile: " + e.Detail
	default:
		return "text input error: " + e.Detail
	}
}

// Normalize returns input unchanged as Raw and its NFC representation. Each
// NFC segment is mapped to the exact source byte range consumed to produce it.
// The default request size bound is applied before normalization.
func Normalize(input string) (Normalized, error) {
	return NormalizeWithLimit(input, status.DefaultBounds().MaxInputBytes)
}

// NormalizeWithLimit is Normalize with an explicit maximum source size in
// UTF-8 bytes. For input within that limit, malformed UTF-8 is reported as
// invalid input rather than as an infrastructure error.
func NormalizeWithLimit(input string, maxBytes int) (Normalized, error) {
	if maxBytes <= 0 {
		return Normalized{}, &TextError{
			Kind: TextErrorInvalidLimit, Limit: maxBytes, Detail: "maximum input bytes must be positive",
		}
	}
	// Enforce the byte budget before decoding so malformed oversized input
	// cannot force an unbounded validation scan.
	if len(input) > maxBytes {
		return Normalized{}, &TextError{
			Kind: TextErrorResourceLimit, Limit: maxBytes, Actual: len(input),
		}
	}
	if !utf8.ValidString(input) {
		for i := 0; i < len(input); {
			_, size := utf8.DecodeRuneInString(input[i:])
			if size == 1 && input[i] >= utf8.RuneSelf {
				return Normalized{}, &TextError{Kind: TextErrorInvalidUTF8, Offset: i}
			}
			i += size
		}
	}
	result := Normalized{Raw: input}
	var iter norm.Iter
	iter.InitString(norm.NFC, input)
	for !iter.Done() {
		originalStart := iter.Pos()
		chunk := iter.Next()
		originalEnd := iter.Pos()
		normalizedStart := len(result.NFC)
		result.NFC += string(chunk)
		result.Clusters = append(result.Clusters, Cluster{
			NFC:    Span{Start: normalizedStart, End: len(result.NFC)},
			Source: Span{Start: originalStart, End: originalEnd},
		})
	}
	return result, nil
}

// OriginalSpan maps a normalized span to source bytes when both endpoints lie
// on normalization-segment boundaries. It returns false for spans that split
// a segment, since such a split has no exact source-byte counterpart.
func (n Normalized) OriginalSpan(span Span) (Span, bool) {
	if span.Start < 0 || span.End < span.Start || span.End > len(n.NFC) {
		return Span{}, false
	}
	start, startOK := n.sourceBoundary(span.Start)
	end, endOK := n.sourceBoundary(span.End)
	if !startOK || !endOK {
		return Span{}, false
	}
	return Span{Start: start, End: end}, true
}

func (n Normalized) sourceBoundary(offset int) (int, bool) {
	if offset == 0 {
		return 0, true
	}
	if offset == len(n.NFC) {
		return len(n.Raw), true
	}
	for _, cluster := range n.Clusters {
		if offset == cluster.NFC.Start {
			return cluster.Source.Start, true
		}
		if offset == cluster.NFC.End {
			return cluster.Source.End, true
		}
	}
	return 0, false
}

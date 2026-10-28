package engine

import (
	"context"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
)

// MaxScopedTargetAlternatives bounds the number of complete scope-preserving
// target alternatives returned by a scope realizer.
const MaxScopedTargetAlternatives = 16

// ScopedTargetLexicalSelection is the exact validated lexical selection for
// one binder or event predicate. These records are evidence supplied by the
// target-side selector; a target realizer must not substitute another sense.
type ScopedTargetLexicalSelection struct {
	Binding    lexicon.SemanticBinding
	Sense      lexicon.LexicalSense
	Lexeme     lexicon.MorphLexeme
	Validation language.TargetValidation
	Role       string
	Port       string
}

// ScopedTargetRoles explicitly maps the two quantified entity binders to
// semantic roles and event ports. EventPort names the event-valued port.
type ScopedTargetRoles struct {
	XRole     string
	XPort     string
	YRole     string
	YPort     string
	EventPort string
}

// ScopedTargetRequest carries the checked scope pattern and its exact
// language-specific lexical evidence. CheckedExpression must be the original
// term accepted by algebra.Check; binders and negation placement are retained
// separately so target construction cannot silently collapse scope.
type ScopedTargetRequest struct {
	CheckedExpression  *algebra.Term
	Paths              ScopedTargetPaths
	BinderX            string
	BinderY            string
	BinderEvent        string
	NegationLocation   ScopedNegationLocation
	EventTemporalFrame string
	Roles              ScopedTargetRoles
	XSelection         ScopedTargetLexicalSelection
	YSelection         ScopedTargetLexicalSelection
	EventSelection     ScopedTargetLexicalSelection
}

// ScopedTargetPaths names the exact nodes in CheckedExpression that establish
// the recognized quantified scope. Negation identifies the enclosing Not
// node at the requested placement.
type ScopedTargetPaths struct {
	OuterQuantifier  string `json:"outer_quantifier"`
	Restrictor       string `json:"restrictor"`
	InnerExistential string `json:"inner_existential"`
	InnerPredicate   string `json:"inner_predicate"`
	EventBinder      string `json:"event_binder"`
	EventAtom        string `json:"event_atom"`
	Negation         string `json:"negation"`
}

// ScopedNegationLocation identifies which side of the quantified scope
// contains negation. Its values correspond to the two recognized patterns.
type ScopedNegationLocation string

const (
	ScopedNegationOuter ScopedNegationLocation = "outer"
	ScopedNegationInner ScopedNegationLocation = "inner"
)

// ScopedSurfacePart is one ordered lexical/morphological piece. BinderID and
// Port identify the semantic argument licensed by the generated form; either
// may be empty for a constructional operator or connective.
type ScopedSurfacePart struct {
	Text      string               `json:"text"`
	BinderID  string               `json:"binder_id,omitempty"`
	Port      string               `json:"port,omitempty"`
	Generated *morph.GeneratedForm `json:"generated,omitempty"`
}

// ScopedTargetAlternative is a complete construction choice with its
// ordered surface pieces and the morphology evidence behind generated pieces.
type ScopedTargetAlternative struct {
	ID             string
	ConstructionID ids.ConstructionID
	Parts          []ScopedSurfacePart
}

// ScopedTargetResult contains at most MaxScopedTargetAlternatives alternatives.
// Status and diagnostics describe selection/coverage independently from text.
type ScopedTargetResult struct {
	Status       status.OutcomeKind
	Alternatives []ScopedTargetAlternative
	Diagnostics  []status.Diagnostic
}

// ScopedTargetRealizer realizes a checked quantified expression while
// preserving its binder identities, negation placement, and selected senses.
// Implementations must bound returned alternatives by
// MaxScopedTargetAlternatives and report truncation through status/diagnostics.
type ScopedTargetRealizer interface {
	RealizeScope(context.Context, ScopedTargetRequest) (ScopedTargetResult, error)
}

// ScopedTargetVerifier exposes the validated, language-local construction plan
// and independently checks that one exact alternative is licensed by a fresh
// realization of the supplied request.
type ScopedTargetVerifier interface {
	ScopedPlan() (construction.ScopedClausePlan, error)
	VerifyScope(context.Context, ScopedTargetRequest, ScopedTargetAlternative) error
}

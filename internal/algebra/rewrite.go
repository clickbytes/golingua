package algebra

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/status"
)

// RewriteRule is a closed, versioned rewrite license. A profile containing no
// rules is valid and licenses no transformations.
type RewriteRule string

const (
	RewriteAlphaV1          RewriteRule = "alpha_v1"
	RewriteBetaV1           RewriteRule = "beta_v1"
	RewriteAliasV1          RewriteRule = "alias_v1"
	RewriteAndFlattenV1     RewriteRule = "and_flatten_v1"
	RewriteAndOrderV1       RewriteRule = "and_order_v1"
	RewriteAndDeduplicateV1 RewriteRule = "and_deduplicate_v1"
)

// RewriteProfile explicitly licenses a bounded set of rewrite rules.
type RewriteProfile struct {
	ID    string        `json:"id"`
	Rules []RewriteRule `json:"rules"`
}

// RewriteBudget bounds normalization work and retained certificates.
type RewriteBudget struct {
	MaxSteps      int `json:"max_steps"`
	MaxExpansions int `json:"max_expansions"`
	MaxTraceNodes int `json:"max_trace_nodes"`
	MaxWorkNodes  int `json:"max_work_nodes"`
}

const (
	DefaultRewriteMaxSteps      = 128
	DefaultRewriteMaxExpansions = 64
	DefaultRewriteMaxTraceNodes = 32768
	DefaultRewriteMaxWorkNodes  = 262144
	MaxRewriteTermBytes         = MaxDocumentBytes
	MaxRewriteCertificateBytes  = 8 << 20
	MaxRewriteEnvelopeBytes     = 16 << 20
)

// NormalizeRewriteBudget applies finite defaults and rejects invalid limits.
func NormalizeRewriteBudget(requested RewriteBudget) (RewriteBudget, error) {
	fields := []struct {
		name  string
		value *int
		cap   int
	}{
		{"max_steps", &requested.MaxSteps, DefaultRewriteMaxSteps},
		{"max_expansions", &requested.MaxExpansions, DefaultRewriteMaxExpansions},
		{"max_trace_nodes", &requested.MaxTraceNodes, DefaultRewriteMaxTraceNodes},
		{"max_work_nodes", &requested.MaxWorkNodes, DefaultRewriteMaxWorkNodes},
	}
	for _, field := range fields {
		if *field.value == 0 {
			*field.value = field.cap
			continue
		}
		if *field.value < 0 || *field.value > field.cap {
			return RewriteBudget{}, invalid("INVALID_REWRITE_BUDGET", "budget."+field.name,
				fmt.Sprintf("must be zero (default) or between 1 and %d", field.cap))
		}
	}
	return requested, nil
}

// AliasDefinition supplies an approved alias expression to the rewrite
// environment. Its signature is present in Document.Signatures for checking;
// it does not make the alias an operational primitive.
type AliasDefinition struct {
	ID         string        `json:"id"`
	Ports      []Port        `json:"ports"`
	Expression *Term         `json:"expression"`
	Interface  []PortBinding `json:"interface"`
}

// RewriteEnvironment contains the checked input context and the immutable
// metadata snapshot against which alias expansion may be licensed.
type RewriteEnvironment struct {
	Document   Document          `json:"document"`
	Aliases    []AliasDefinition `json:"aliases"`
	SnapshotID string            `json:"snapshot_id"`
}

// BinderChange records a binder's identity correspondence across a step.
type BinderChange struct {
	Before string `json:"before"`
	After  string `json:"after"`
	Kind   string `json:"kind"`
}

// RewriteStep is one certified local transformation over a full root term.
type RewriteStep struct {
	RuleID           RewriteRule    `json:"rule_id"`
	Path             []int          `json:"path"`
	Before           *Term          `json:"before"`
	After            *Term          `json:"after"`
	BinderMapping    []BinderChange `json:"binder_mapping,omitempty"`
	ParameterMapping Substitution   `json:"parameter_mapping,omitempty"`
	SideConditions   []string       `json:"side_conditions,omitempty"`
	SnapshotID       string         `json:"snapshot_id"`
	ProfileID        string         `json:"profile_id"`
	BeforeNodes      int            `json:"before_nodes"`
	AfterNodes       int            `json:"after_nodes"`
}

// RewriteCertificate records a sequence of licensed transformations. It
// certifies only the listed rewrites, never truth or unrestricted equivalence.
type RewriteCertificate struct {
	SchemaVersion      int            `json:"schema_version"`
	SnapshotID         string         `json:"snapshot_id"`
	ContextFingerprint string         `json:"context_fingerprint,omitempty"`
	Profile            RewriteProfile `json:"profile"`
	Input              *Term          `json:"input"`
	Output             *Term          `json:"output"`
	Steps              []RewriteStep  `json:"steps"`
	Normalized         bool           `json:"normalized"`
}

// ValidateRewriteProfile checks the closed rule vocabulary and bounded ID.
func ValidateRewriteProfile(profile RewriteProfile) error {
	if !validText(profile.ID, 256) {
		return invalid("INVALID_REWRITE_PROFILE_ID", "profile.id", "profile ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	if len(profile.Rules) > 6 {
		return resourceLimit("profile.rules", "rewrite rule count exceeds the closed vocabulary size of 6")
	}
	known := map[RewriteRule]bool{
		RewriteAlphaV1: true, RewriteBetaV1: true, RewriteAliasV1: true,
		RewriteAndFlattenV1: true, RewriteAndOrderV1: true, RewriteAndDeduplicateV1: true,
	}
	seen := make(map[RewriteRule]bool, len(profile.Rules))
	for i, rule := range profile.Rules {
		path := fmt.Sprintf("profile.rules[%d]", i)
		if !known[rule] {
			return &Error{Status: status.OutcomeUnsupported, Code: "UNKNOWN_REWRITE_RULE", Path: path, Message: "rewrite rule is not supported"}
		}
		if seen[rule] {
			return invalid("DUPLICATE_REWRITE_RULE", path, "rewrite rule is listed more than once")
		}
		seen[rule] = true
	}
	return nil
}

// ValidateRewriteEnvironment checks the structure and typing of supplied
// declarations and alias definitions. It does not validate provenance or
// prove approval. A trusted caller must supply all and only approved aliases
// from validated registries and bind the real composite formal snapshot ID.
// This function neither evaluates propositions nor activates aliases as
// primitive relations.
func ValidateRewriteEnvironment(environment RewriteEnvironment) error {
	if !validText(environment.SnapshotID, 256) {
		return invalid("INVALID_SNAPSHOT_ID", "snapshot_id", "snapshot ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	// Check caller-owned structures before making any copies; the checker
	// detects cyclic and over-depth type graphs without unbounded recursion.
	if err := CheckDeclarations(environment.Document); err != nil {
		return err
	}
	if len(environment.Aliases)+len(environment.Document.Signatures)+len(environment.Document.Sorts)+len(environment.Document.Parameters)+len(environment.Document.References) > DefaultMaxSymbols {
		return resourceLimit("aliases", "aggregate alias and declaration count exceeds 512")
	}

	aliases := make(map[string]AliasDefinition, len(environment.Aliases))
	for i, alias := range environment.Aliases {
		path := fmt.Sprintf("aliases[%d]", i)
		if !validText(alias.ID, 256) {
			return invalid("INVALID_ALIAS_ID", path+".id", "alias ID must be non-empty valid UTF-8 and at most 256 bytes")
		}
		if _, exists := aliases[alias.ID]; exists {
			return invalid("DUPLICATE_ALIAS_ID", path+".id", "alias ID is duplicated")
		}
		if alias.Expression == nil {
			return invalid("MISSING_ALIAS_EXPRESSION", path+".expression", "alias expression is required")
		}
		aliases[alias.ID] = alias
	}

	combined := environment.Document
	combined.Signatures = append([]Signature(nil), environment.Document.Signatures...)
	for i, alias := range environment.Aliases {
		if _, exists := findSignature(environment.Document.Signatures, alias.ID); !exists {
			return invalid("ALIAS_SIGNATURE_MISSING", fmt.Sprintf("aliases[%d].id", i), "alias requires a matching declared signature")
		}
	}
	if err := CheckDeclarations(combined); err != nil {
		return err
	}

	signatures := make(map[string]Signature, len(combined.Signatures))
	for _, signature := range combined.Signatures {
		signatures[signature.ID] = signature
	}
	for i, alias := range environment.Aliases {
		path := fmt.Sprintf("aliases[%d]", i)
		signature, exists := signatures[alias.ID]
		if !exists {
			return invalid("ALIAS_SIGNATURE_MISSING", path+".id", "alias requires a matching declared signature")
		}
		if !samePorts(signature.Ports, alias.Ports) {
			return invalid("ALIAS_SIGNATURE_MISMATCH", path+".ports", "alias ports must exactly match the declared signature in order and type")
		}
		if len(alias.Interface) != len(alias.Ports) {
			return invalid("ALIAS_INTERFACE_MISMATCH", path+".interface", "alias interface must map every signature port exactly once")
		}
		for j, port := range alias.Ports {
			if alias.Interface[j].Port != port.Role {
				return invalid("ALIAS_INTERFACE_ORDER", fmt.Sprintf("%s.interface[%d]", path, j), "alias interface must follow signature port order")
			}
		}
		aliasDocument := combined
		aliasDocument.Parameters = nil
		aliasDocument.References = nil
		aliasDocument.Closed = true
		aliasDocument.Expression = alias.Expression
		checked, err := Check(aliasDocument)
		if err != nil {
			return err
		}
		if checked.Status != status.OutcomeOK || len(checked.Holes) != 0 || len(checked.FreeParameters) != 0 {
			return invalid("ALIAS_NOT_CLOSED", path+".expression", "alias must be closed and contain no holes")
		}
		function := Function{Expression: alias.Expression, Interface: alias.Interface}
		if err := ValidateFunction(aliasDocument, function); err != nil {
			return err
		}
	}

	return validateAliasAcyclic(environment.Aliases, DefaultMaxNodes)
}

func findSignature(signatures []Signature, id string) (Signature, bool) {
	for _, signature := range signatures {
		if signature.ID == id {
			return signature, true
		}
	}
	return Signature{}, false
}

func samePorts(left, right []Port) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].Role != right[i].Role || !TypeEqual(left[i].Type, right[i].Type) {
			return false
		}
	}
	return true
}

func validateAliasAcyclic(aliases []AliasDefinition, maxNodes int) error {
	aliasIDs := make(map[string]int, len(aliases))
	for i, alias := range aliases {
		aliasIDs[alias.ID] = i
	}
	graph := make([][]int, len(aliases))
	nodes := 0
	for i, alias := range aliases {
		stack := []*Term{alias.Expression}
		seen := make(map[*Term]bool)
		for len(stack) > 0 {
			term := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if term == nil || seen[term] {
				continue
			}
			seen[term] = true
			nodes++
			if nodes > maxNodes {
				return resourceLimit("aliases", "alias dependency scan exceeds node budget")
			}
			if term.Kind == TermAtom {
				if target, ok := aliasIDs[term.Symbol]; ok {
					graph[i] = append(graph[i], target)
				}
			}
			for _, argument := range term.Args {
				stack = append(stack, argument.Value)
			}
			stack = append(stack, term.Operands...)
			stack = append(stack, term.Body, term.Function, term.Argument)
		}
	}

	state := make([]uint8, len(aliases))
	var visit func(int) error
	visit = func(index int) error {
		if state[index] == 1 {
			return invalid("ALIAS_CYCLE", "aliases", "alias definitions contain a dependency cycle")
		}
		if state[index] == 2 {
			return nil
		}
		state[index] = 1
		for _, target := range graph[index] {
			if err := visit(target); err != nil {
				return err
			}
		}
		state[index] = 2
		return nil
	}
	for i := range aliases {
		if err := visit(i); err != nil {
			return err
		}
	}
	return nil
}

package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/catalogexport"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/metamodel"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/reason"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/synthesis"
)

type algebraOptions struct {
	command            string
	input              string
	languagePackage    string
	left               string
	right              string
	proof              string
	claim              string
	goal               string
	candidate          string
	certificate        string
	mode               string
	metadata           []string
	expression         string
	factsInput         string
	selection          string
	selectionSpecified bool
	contextID          string
	profile            string
	at                 string
	evaluationLimits   reason.AlgebraEvaluationLimits
	evaluationLimitSet map[string]bool
	comparisonLimits   algebra.EquivalenceLimits
	comparisonLimitSet map[string]bool
	json               bool
	noCache            bool
	searchLimits       synthesis.SearchLimits
	searchLimitSet     map[string]bool
	verifyLimits       synthesis.GoalVerificationLimits
	verifyLimitSet     map[string]bool
}

type algebraContextCheckOutput struct {
	Status            status.OutcomeKind        `json:"status"`
	Checked           algebra.CheckedExpression `json:"checked"`
	CanonicalDocument json.RawMessage           `json:"canonical_document"`
	Assessment        algebra.ContextAssessment `json:"request_context_assessment"`
	Metadata          []metadataReference       `json:"metadata,omitempty"`
	ValidationScope   string                    `json:"validation_scope"`
}

type algebraScopeCapability struct {
	CanRepresent bool   `json:"can_represent"`
	CanEvaluate  bool   `json:"can_evaluate"`
	Condition    string `json:"condition,omitempty"`
}

type algebraCapabilitiesOutput struct {
	Status       status.OutcomeKind                `json:"status"`
	Scopes       map[string]algebraScopeCapability `json:"scopes"`
	QuestionType string                            `json:"question_type"`
	Evaluation   algebraScopeCapability            `json:"question_evaluation"`
	Scope        string                            `json:"scope"`
}

type algebraNormalizeOutput struct {
	Status   status.OutcomeKind          `json:"status"`
	Result   algebra.NormalizationResult `json:"result"`
	Metadata []metadataReference         `json:"metadata"`
}

type algebraRewriteVerificationOutput struct {
	Status   status.OutcomeKind         `json:"status"`
	Result   algebra.VerificationResult `json:"result"`
	Metadata []metadataReference        `json:"metadata"`
}

type algebraOutput struct {
	Status            status.OutcomeKind        `json:"status"`
	Checked           algebra.CheckedExpression `json:"checked"`
	CanonicalDocument json.RawMessage           `json:"canonical_document"`
	Metadata          []metadataReference       `json:"metadata,omitempty"`
	Diagnostics       []status.Diagnostic       `json:"diagnostics,omitempty"`
}

type metadataReference struct {
	ID                  string `json:"id"`
	Version             string `json:"version"`
	ResourceFingerprint string `json:"resource_fingerprint"`
}

type algebraBuildOutput struct {
	algebra.BuildResult
	ResultKind   string                `json:"result_kind"`
	Substitution *algebra.Substitution `json:"substitution,omitempty"`
	Metadata     []metadataReference   `json:"metadata,omitempty"`
	Diagnostics  []status.Diagnostic   `json:"diagnostics,omitempty"`
}

type metadataPackageOutput struct {
	ID                  string                 `json:"id"`
	Version             string                 `json:"version"`
	Profile             string                 `json:"profile"`
	ResourceFingerprint string                 `json:"resource_fingerprint"`
	FormalDenotationID  string                 `json:"formal_denotation_id"`
	SortCount           int                    `json:"sort_count"`
	SymbolCount         int                    `json:"symbol_count"`
	PrimitiveCount      int                    `json:"primitive_count"`
	AliasCount          int                    `json:"alias_count"`
	OpaqueCount         int                    `json:"opaque_unformalized_count"`
	PendingCount        int                    `json:"pending_count"`
	ApprovedCount       int                    `json:"approved_count"`
	Provenance          []provenance.SourceRef `json:"provenance,omitempty"`
	Symbols             []metadataSymbolOutput `json:"symbols,omitempty"`
}

type semanticBindingInspectOutput struct {
	Status           status.OutcomeKind           `json:"status"`
	Package          semanticBindingPackageOutput `json:"language_package"`
	MetadataVerified bool                         `json:"metadata_verified"`
	Metadata         []metadataReference          `json:"metadata"`
	ValidationScope  string                       `json:"validation_scope"`
}

type semanticBindingPackageOutput struct {
	ID          string                        `json:"id"`
	Version     string                        `json:"version"`
	Language    string                        `json:"language"`
	Fingerprint string                        `json:"fingerprint"`
	Provenance  provenance.SourceRef          `json:"provenance"`
	Frames      []semanticBindingFrameOutput  `json:"frames"`
	Targets     []semanticBindingTargetOutput `json:"targets"`
}

type semanticBindingFrameOutput struct {
	ID               string                          `json:"id"`
	SenseID          string                          `json:"sense_id"`
	SemanticRoles    []string                        `json:"semantic_roles"`
	Arguments        []semanticBindingArgumentOutput `json:"arguments"`
	AllowedFamilyIDs []string                        `json:"allowed_family_ids"`
	Provenance       provenance.SourceRef            `json:"provenance"`
}

type semanticBindingArgumentOutput struct {
	ID                  string `json:"id"`
	Role                string `json:"role"`
	Required            bool   `json:"required"`
	GrammaticalFunction string `json:"grammatical_function,omitempty"`
	Case                string `json:"case,omitempty"`
	Adposition          string `json:"adposition,omitempty"`
}

type semanticBindingTargetOutput struct {
	SenseID    string                     `json:"sense_id"`
	Mapping    string                     `json:"mapping_kind"`
	Status     string                     `json:"binding_status"`
	Authorized bool                       `json:"authorized"`
	Provenance provenance.SourceRef       `json:"provenance"`
	Target     *semanticBindingTargetData `json:"target,omitempty"`
}

type semanticBindingTargetData struct {
	Kind      string                    `json:"kind"`
	Profile   string                    `json:"profile"`
	SymbolRef string                    `json:"symbol_ref,omitempty"`
	Goal      any                       `json:"goal,omitempty"`
	RolePorts []semanticBindingRolePort `json:"role_ports"`
}

type semanticBindingRolePort struct {
	FrameID string `json:"frame_id"`
	Role    string `json:"role"`
	PortID  string `json:"port_id"`
	Type    any    `json:"type"`
}

type semanticBindingValidationOutput struct {
	Status           status.OutcomeKind           `json:"status"`
	Package          semanticBindingPackageOutput `json:"language_package"`
	Metadata         []metadataReference          `json:"metadata"`
	MetadataVerified bool                         `json:"metadata_verified"`
	Targets          []language.TargetValidation  `json:"targets"`
	ValidationScope  string                       `json:"validation_scope"`
}

type metadataSymbolOutput struct {
	ID           string                 `json:"id"`
	Kind         metamodel.SymbolKind   `json:"kind"`
	Label        string                 `json:"label,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ReviewStatus metamodel.ReviewStatus `json:"review_status"`
	Provenance   []provenance.SourceRef `json:"provenance"`
}

type metadataCommandOutput struct {
	Status          status.OutcomeKind      `json:"status"`
	Operation       string                  `json:"operation"`
	ValidationScope string                  `json:"validation_scope,omitempty"`
	Packages        []metadataPackageOutput `json:"packages"`
}

type artifactExportOutput struct {
	Status        status.OutcomeKind        `json:"status"`
	Artifact      algebra.ArtifactEnvelope  `json:"artifact"`
	Checked       algebra.CheckedExpression `json:"checked"`
	UniqueNodes   int                       `json:"unique_nodes"`
	ExpandedNodes int                       `json:"expanded_nodes"`
	Interning     bool                      `json:"interning_enabled"`
	CacheHit      bool                      `json:"cache_hit"`
	Metadata      []metadataReference       `json:"metadata"`
}

type artifactReplayOutput struct {
	Status        status.OutcomeKind        `json:"status"`
	Artifact      algebra.ArtifactEnvelope  `json:"artifact"`
	Checked       algebra.CheckedExpression `json:"checked"`
	UniqueNodes   int                       `json:"unique_nodes"`
	ExpandedNodes int                       `json:"expanded_nodes"`
	Metadata      []metadataReference       `json:"metadata"`
}

type artifactInspectionOutput struct {
	Status              status.OutcomeKind          `json:"status"`
	ValidationScope     string                      `json:"validation_scope"`
	MetadataVerified    bool                        `json:"metadata_verified"`
	NormalizerVersion   string                      `json:"normalizer_version"`
	RegistrySnapshotID  string                      `json:"registry_snapshot_id"`
	RequiredRegistries  []algebra.RegistryPin       `json:"required_registry_pins"`
	Profile             algebra.RewriteProfile      `json:"profile"`
	ContextID           string                      `json:"context_id"`
	RequestContext      *algebra.RequestContext     `json:"request_context,omitempty"`
	References          []algebra.Declaration       `json:"references"`
	ContextDependencies []algebra.ContextDependency `json:"context_dependencies"`
	Provenance          []provenance.SourceRef      `json:"provenance"`
	ResourceFingerprint string                      `json:"resource_fingerprint"`
	MeaningFingerprint  string                      `json:"meaning_fingerprint"`
	CacheKey            string                      `json:"cache_key"`
	Root                algebra.NodeID              `json:"root"`
	UniqueNodes         int                         `json:"unique_nodes"`
	ExpandedNodes       int                         `json:"expanded_nodes"`
	Checked             algebra.CheckedExpression   `json:"checked"`
}

type naryFactsOutput struct {
	Status          status.OutcomeKind    `json:"status"`
	SnapshotID      string                `json:"snapshot_id"`
	ContextID       string                `json:"context_id"`
	At              string                `json:"at"`
	Referents       []algebra.Declaration `json:"referents"`
	Facts           []facts.NaryFact      `json:"facts"`
	Summary         facts.NarySummary     `json:"summary"`
	Metadata        []metadataReference   `json:"metadata"`
	AbsenceMeans    string                `json:"absence_means"`
	ValidationScope string                `json:"validation_scope"`
}

type catalogExportOutput struct {
	Status          status.OutcomeKind           `json:"status"`
	ValidationScope string                       `json:"validation_scope"`
	Inventory       catalogexport.ExportDocument `json:"inventory"`
}

type algebraEvaluationOutput struct {
	Status          status.OutcomeKind             `json:"status"`
	Result          reason.AlgebraEvaluationResult `json:"result"`
	Metadata        []metadataReference            `json:"metadata"`
	ValidationScope string                         `json:"validation_scope"`
}

type algebraEquivalenceOutput struct {
	Status          status.OutcomeKind        `json:"status"`
	ClaimType       string                    `json:"claim_type"`
	Result          algebra.EquivalenceResult `json:"result"`
	Metadata        []metadataReference       `json:"metadata"`
	ValidationScope string                    `json:"validation_scope"`
}

type algebraProofOutput struct {
	Status          status.OutcomeKind  `json:"status"`
	ClaimType       string              `json:"claim_type"`
	Result          any                 `json:"result"`
	Metadata        []metadataReference `json:"metadata"`
	ValidationScope string              `json:"validation_scope"`
}

type synthesisOutput struct {
	Status          status.OutcomeKind             `json:"status"`
	Mode            string                         `json:"mode"`
	GoalProfile     string                         `json:"goal_profile"`
	ContextID       string                         `json:"context_id"`
	Metadata        []metadataReference            `json:"metadata"`
	Result          synthesis.Result               `json:"result"`
	ValidationScope string                         `json:"validation_scope"`
	ExportRequest   *algebra.ArtifactExportRequest `json:"export_request,omitempty"`
}

type goalVerificationOutput struct {
	Status          status.OutcomeKind               `json:"status"`
	Result          synthesis.GoalVerificationResult `json:"result"`
	Metadata        []metadataReference              `json:"metadata"`
	ValidationScope string                           `json:"validation_scope"`
}

type synthesisSearchOutput struct {
	Status                      status.OutcomeKind     `json:"status"`
	Mode                        string                 `json:"mode"`
	GoalProfile                 string                 `json:"goal_profile"`
	ContextID                   string                 `json:"context_id"`
	Metadata                    []metadataReference    `json:"metadata"`
	Result                      synthesis.SearchResult `json:"result"`
	ValidationScope             string                 `json:"validation_scope"`
	ResponseCandidatesDiscarded int                    `json:"response_candidates_discarded,omitempty"`
}

type algebraInfrastructureError struct{ message string }

func (failure *algebraInfrastructureError) Error() string { return failure.message }

func parseAlgebraArgs(args []string) (algebraOptions, string) {
	var options algebraOptions
	if len(args) < 2 || args[0] != "algebra" {
		return options, "expected algebra command"
	}
	index := 2
	if args[1] == "check" {
		options.command = "check"
	} else if args[1] == "capabilities" {
		options.command = "capabilities"
	} else if args[1] == "build" {
		options.command = "build"
	} else if args[1] == "synthesize" {
		options.command = "synthesize"
	} else if args[1] == "verify-goal" {
		options.command = "verify_goal"
	} else if args[1] == "evaluate" {
		options.command = "evaluate"
	} else if args[1] == "compare" {
		options.command = "compare"
	} else if args[1] == "verify-proof" {
		options.command = "verify_proof"
	} else if args[1] == "normalize" || args[1] == "verify-rewrite" || args[1] == "export" || args[1] == "inspect" || args[1] == "replay" {
		options.command = args[1]
	} else if args[1] == "metadata" {
		if len(args) < 3 || (args[2] != "inspect" && args[2] != "validate") {
			return options, "expected algebra metadata inspect or validate"
		}
		options.command = "metadata_" + args[2]
		index = 3
	} else if args[1] == "facts" {
		if len(args) < 3 || args[2] != "validate" {
			return options, "expected algebra facts validate"
		}
		options.command = "facts_validate"
		index = 3
	} else if args[1] == "catalog" {
		if len(args) < 3 || args[2] != "export" {
			return options, "expected algebra catalog export"
		}
		options.command = "catalog_export"
		index = 3
	} else if args[1] == "binding" {
		if len(args) < 3 || (args[2] != "inspect" && args[2] != "validate" && args[2] != "instantiate") {
			return options, "expected algebra binding inspect, validate, or instantiate"
		}
		options.command = "binding_" + args[2]
		index = 3
	} else if args[1] == "context" {
		if len(args) < 3 || args[2] != "check" {
			return options, "expected algebra context check"
		}
		options.command = "context_check"
		index = 3
	} else {
		return options, "expected algebra check, build, synthesize, verify-goal, evaluate, compare, verify-proof, binding, export, inspect, replay, normalize, verify-rewrite, metadata, facts, or catalog"
	}
	seenJSON, seenInput := false, false
	seenNoCache := false
	seenGoal, seenMode, seenCandidate, seenCertificate := false, false, false, false
	options.searchLimitSet = make(map[string]bool)
	options.verifyLimitSet = make(map[string]bool)
	options.evaluationLimitSet = make(map[string]bool)
	options.comparisonLimitSet = make(map[string]bool)
	seenExpression, seenFacts, seenSelection := false, false, false
	seenLanguagePackage := false
	seenContextID, seenProfile, seenAt := false, false, false
	seenLeft, seenRight, seenProof, seenClaim := false, false, false, false
	for i := index; i < len(args); i++ {
		flag := args[i]
		switch flag {
		case "--json":
			if seenJSON {
				return options, "duplicate --json flag"
			}
			seenJSON = true
			options.json = true
		case "--input":
			if options.command != "check" && options.command != "context_check" && options.command != "build" && options.command != "normalize" && options.command != "verify-rewrite" && options.command != "export" && options.command != "inspect" && options.command != "replay" && options.command != "facts_validate" && options.command != "verify_proof" && options.command != "binding_instantiate" {
				return options, "--input is not valid for this algebra command"
			}
			if seenInput {
				return options, "duplicate --input flag"
			}
			seenInput = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --input"
			}
			i++
			options.input = args[i]
			if options.input == "" {
				return options, "--input value must not be empty"
			}
		case "--expression", "--facts", "--select", "--context-id", "--profile", "--at", "--evaluation-limit":
			allowed := options.command == "evaluate" || options.command == "verify_proof" || ((flag == "--context-id" || flag == "--profile") && options.command == "compare")
			if !allowed {
				return options, flag + " is not valid for this algebra command"
			}
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for " + flag
			}
			i++
			value := args[i]
			switch flag {
			case "--expression":
				if options.command != "evaluate" && options.command != "verify_proof" {
					return options, "--expression is not valid for this command"
				}
				if seenExpression {
					return options, "duplicate --expression flag"
				}
				seenExpression = true
				options.expression = value
			case "--facts":
				if options.command != "evaluate" && options.command != "verify_proof" {
					return options, "--facts is not valid for this command"
				}
				if seenFacts {
					return options, "duplicate --facts flag"
				}
				seenFacts = true
				options.factsInput = value
			case "--select":
				if options.command != "evaluate" && options.command != "verify_proof" {
					return options, "--select is not valid for this command"
				}
				if seenSelection {
					return options, "duplicate --select flag"
				}
				seenSelection = true
				options.selectionSpecified = true
				options.selection = value
			case "--context-id":
				if seenContextID {
					return options, "duplicate --context-id flag"
				}
				seenContextID = true
				options.contextID = value
			case "--profile":
				if seenProfile {
					return options, "duplicate --profile flag"
				}
				seenProfile = true
				options.profile = value
			case "--at":
				if options.command != "evaluate" && options.command != "verify_proof" {
					return options, "--at is not valid for this command"
				}
				if seenAt {
					return options, "duplicate --at flag"
				}
				seenAt = true
				options.at = value
				if value == "" {
					return options, "--at value must be a UTC timestamp"
				}
			case "--evaluation-limit":
				if options.command != "evaluate" && options.command != "verify_proof" {
					return options, "--evaluation-limit is not valid for this command"
				}
				if err := setEvaluationLimit(&options, value); err != "" {
					return options, err
				}
			}
		case "--goal":
			if options.command != "synthesize" && options.command != "verify_goal" && options.command != "verify_proof" {
				return options, "--goal is not valid for this algebra command"
			}
			if seenGoal {
				return options, "duplicate --goal flag"
			}
			seenGoal = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --goal"
			}
			i++
			options.goal = args[i]
			if options.goal == "" {
				return options, "--goal value must not be empty"
			}
		case "--candidate":
			if options.command != "verify_goal" && options.command != "verify_proof" {
				return options, "--candidate is not valid for this algebra command"
			}
			if seenCandidate {
				return options, "duplicate --candidate flag"
			}
			seenCandidate = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --candidate"
			}
			i++
			options.candidate = args[i]
			if options.candidate == "" || options.candidate == "-" {
				return options, "--candidate requires a named file path"
			}
		case "--certificate":
			if options.command != "verify_goal" && options.command != "verify_proof" {
				return options, "--certificate is not valid for this algebra command"
			}
			if seenCertificate {
				return options, "duplicate --certificate flag"
			}
			seenCertificate = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --certificate"
			}
			i++
			options.certificate = args[i]
			if options.certificate == "" || options.certificate == "-" {
				return options, "--certificate requires a named file path"
			}
		case "--mode":
			if options.command != "synthesize" {
				return options, "--mode is only valid for algebra synthesize"
			}
			if seenMode {
				return options, "duplicate --mode flag"
			}
			seenMode = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --mode"
			}
			i++
			options.mode = args[i]
		case "--search-limit":
			if options.command != "synthesize" {
				return options, "--search-limit is only valid for algebra synthesize --mode search"
			}
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --search-limit"
			}
			i++
			if err := setSearchLimit(&options, args[i]); err != "" {
				return options, err
			}
		case "--verify-limit":
			if options.command != "verify_goal" && options.command != "verify_proof" {
				return options, "--verify-limit is only valid for proof verification"
			}
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --verify-limit"
			}
			i++
			if err := setVerifyLimit(&options, args[i]); err != "" {
				return options, err
			}
		case "--comparison-limit":
			if options.command != "compare" && options.command != "verify_proof" {
				return options, "--comparison-limit is only valid for algebra compare or verify-proof equivalence"
			}
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --comparison-limit"
			}
			i++
			if err := setComparisonLimit(&options, args[i]); err != "" {
				return options, err
			}
		case "--claim":
			if options.command != "verify_proof" {
				return options, "--claim is only valid for algebra verify-proof"
			}
			if seenClaim {
				return options, "duplicate --claim flag"
			}
			seenClaim = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --claim"
			}
			i++
			options.claim = args[i]
		case "--left", "--right", "--proof":
			if options.command != "compare" && options.command != "verify_proof" {
				return options, flag + " is not valid for this algebra command"
			}
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for " + flag
			}
			i++
			if args[i] == "" || args[i] == "-" {
				return options, flag + " requires a named file path"
			}
			switch flag {
			case "--left":
				if seenLeft {
					return options, "duplicate --left flag"
				}
				seenLeft, options.left = true, args[i]
			case "--right":
				if seenRight {
					return options, "duplicate --right flag"
				}
				seenRight, options.right = true, args[i]
			case "--proof":
				if seenProof {
					return options, "duplicate --proof flag"
				}
				seenProof, options.proof = true, args[i]
			}
		case "--metadata":
			if options.command == "inspect" {
				return options, "algebra inspect is structural-only and does not accept --metadata"
			}
			if options.command == "catalog_export" {
				return options, "algebra catalog export reads the compiled catalog and does not accept --metadata"
			}
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --metadata"
			}
			i++
			if args[i] == "" {
				return options, "--metadata value must not be empty"
			}
			options.metadata = append(options.metadata, args[i])
		case "--language-package":
			if !strings.HasPrefix(options.command, "binding_") {
				return options, "--language-package is only valid for algebra binding"
			}
			if seenLanguagePackage {
				return options, "duplicate --language-package flag"
			}
			seenLanguagePackage = true
			if i+1 >= len(args) || args[i+1] == "--json" || strings.HasPrefix(args[i+1], "--") {
				return options, "missing value for --language-package"
			}
			i++
			options.languagePackage = args[i]
			if options.languagePackage == "" || options.languagePackage == "-" {
				return options, "--language-package requires an explicit named file path"
			}
		case "--no-cache":
			if options.command != "export" {
				return options, "--no-cache is only valid for algebra export"
			}
			if seenNoCache {
				return options, "duplicate --no-cache flag"
			}
			seenNoCache = true
			options.noCache = true
		default:
			return options, fmt.Sprintf("unknown algebra argument %q", flag)
		}
	}
	if !seenJSON {
		return options, "algebra command requires --json"
	}
	if options.command == "capabilities" && (len(options.metadata) != 0 || options.input != "" || options.expression != "" || options.factsInput != "" || options.selectionSpecified || options.contextID != "" || options.profile != "" || options.at != "") {
		return options, "algebra capabilities accepts only --json"
	}
	if strings.HasPrefix(options.command, "metadata_") && len(options.metadata) == 0 {
		return options, "algebra metadata requires at least one --metadata file"
	}
	if options.command == "facts_validate" && len(options.metadata) == 0 {
		return options, "algebra facts validate requires at least one --metadata file"
	}
	if options.command == "evaluate" {
		if !seenExpression || !seenFacts || !seenContextID || !seenProfile {
			return options, "algebra evaluate requires --expression, --facts, --context-id, and --profile"
		}
		if len(options.metadata) == 0 {
			return options, "algebra evaluate requires at least one --metadata file"
		}
		for _, path := range []string{options.expression, options.factsInput} {
			if path == "" || path == "-" {
				return options, "algebra evaluate inputs must be named file paths"
			}
		}
		if options.selectionSpecified && (options.selection == "" || options.selection == "-") {
			return options, "an explicit --select input must be a named file path"
		}
		if options.contextID == "" || options.profile != reason.AlgebraQueryProfile {
			return options, "algebra evaluate requires a nonempty context ID and profile positive-facts-v1"
		}
	}
	if options.command == "compare" {
		if !seenLeft || !seenRight || !seenContextID || !seenProfile || len(options.metadata) == 0 {
			return options, "algebra compare requires --left, --right, --profile, --context-id, and at least one --metadata file"
		}
		if options.contextID == "" || options.profile == "" {
			return options, "comparison profile path and context ID must not be empty"
		}
		if options.input != "" || options.goal != "" || options.candidate != "" || options.certificate != "" || options.proof != "" || options.expression != "" || options.factsInput != "" || options.selection != "" || options.at != "" || options.claim != "" || len(options.verifyLimitSet) != 0 || len(options.evaluationLimitSet) != 0 {
			return options, "algebra compare received flags from another command profile"
		}
	}
	if options.command == "verify_proof" {
		if !seenClaim || options.claim == "" {
			return options, "algebra verify-proof requires an explicit --claim"
		}
		if options.claim == "equivalence" {
			if !seenLeft || !seenRight || !seenProfile || !seenContextID || !seenCertificate || len(options.metadata) == 0 {
				return options, "equivalence verification requires --left, --right, --profile, --context-id, --certificate, and metadata"
			}
			if options.left == "" || options.right == "" || options.profile == "" || options.contextID == "" {
				return options, "equivalence verification file paths and context ID must not be empty"
			}
			if options.input != "" || options.goal != "" || options.candidate != "" || options.proof != "" || options.expression != "" || options.factsInput != "" || options.selection != "" || options.at != "" || len(options.evaluationLimitSet) != 0 || len(options.verifyLimitSet) != 0 {
				return options, "equivalence verification received flags from another claim"
			}
		} else if options.claim == "rewrite" {
			if !seenInput || len(options.metadata) == 0 {
				return options, "rewrite verification requires --input and metadata"
			}
			if options.left != "" || options.right != "" || options.proof != "" || options.goal != "" || options.candidate != "" || options.certificate != "" || options.expression != "" || options.factsInput != "" || options.selection != "" || options.contextID != "" || options.profile != "" || options.at != "" || len(options.evaluationLimitSet)+len(options.verifyLimitSet)+len(options.comparisonLimitSet) != 0 {
				return options, "rewrite verification received flags from another claim"
			}
		} else if options.claim == "goal" {
			if !seenGoal || !seenCandidate || !seenCertificate || len(options.metadata) == 0 {
				return options, "goal verification requires --goal, --candidate, --certificate, and metadata"
			}
			if options.left != "" || options.right != "" || options.proof != "" || options.input != "" || options.expression != "" || options.factsInput != "" || options.selection != "" || options.contextID != "" || options.profile != "" || options.at != "" || len(options.evaluationLimitSet)+len(options.comparisonLimitSet) != 0 {
				return options, "goal verification received flags from another claim"
			}
		} else if options.claim == "evaluation" {
			if !seenExpression || !seenFacts || !seenContextID || !seenProfile || !seenProof || len(options.metadata) == 0 {
				return options, "evaluation verification requires --expression, --facts, --proof, --context-id, --profile, and metadata"
			}
			if options.expression == "" || options.factsInput == "" || options.proof == "" || options.contextID == "" || options.profile != reason.AlgebraQueryProfile {
				return options, "evaluation verification requires named inputs, a context ID, and profile positive-facts-v1"
			}
			if options.expression == "-" || options.factsInput == "-" || options.proof == "-" || options.selectionSpecified && (options.selection == "" || options.selection == "-") {
				return options, "evaluation verification requires named file paths"
			}
			if options.left != "" || options.right != "" || options.input != "" || options.goal != "" || options.candidate != "" || options.certificate != "" || len(options.verifyLimitSet)+len(options.comparisonLimitSet) != 0 {
				return options, "evaluation verification received flags from another claim"
			}
		}
	}
	if len(options.metadata) > 16 {
		return options, "algebra metadata accepts at most 16 files"
	}
	if strings.HasPrefix(options.command, "binding_") {
		if !seenLanguagePackage {
			return options, "algebra binding requires --language-package <file>"
		}
		if options.command == "binding_validate" && len(options.metadata) == 0 {
			return options, "algebra binding validate requires at least one --metadata file"
		}
		if options.command == "binding_instantiate" && (len(options.metadata) == 0 || !seenInput) {
			return options, "algebra binding instantiate requires --metadata and --input <request.json|->"
		}
		if options.command != "binding_instantiate" && seenInput {
			return options, "--input is only valid for algebra binding instantiate"
		}
		if options.command != "binding_instantiate" && (len(options.searchLimitSet)+len(options.verifyLimitSet)+len(options.evaluationLimitSet)+len(options.comparisonLimitSet) != 0 || options.goal != "" || options.expression != "" || options.factsInput != "" || options.selection != "" || options.contextID != "" || options.profile != "" || options.at != "" || options.claim != "" || options.left != "" || options.right != "" || options.proof != "" || options.candidate != "" || options.certificate != "" || options.mode != "" || options.noCache) {
			return options, "algebra binding received flags from another command profile"
		}
		if options.command == "binding_instantiate" && (options.goal != "" || options.expression != "" || options.factsInput != "" || options.selection != "" || options.contextID != "" || options.profile != "" || options.at != "" || options.claim != "" || options.left != "" || options.right != "" || options.proof != "" || options.candidate != "" || options.certificate != "" || options.mode != "" || options.noCache || len(options.searchLimitSet)+len(options.verifyLimitSet)+len(options.evaluationLimitSet)+len(options.comparisonLimitSet) != 0) {
			return options, "algebra binding instantiate received flags from another command profile"
		}
	}
	if options.command == "build" {
		if len(options.metadata) == 0 {
			return options, "algebra build requires at least one --metadata file"
		}
	}
	if options.command == "context_check" && !seenInput {
		return options, "algebra context check requires --input <document.json>"
	}
	if options.command == "synthesize" {
		if !seenMode || (options.mode != "construct" && options.mode != "search") {
			return options, "algebra synthesize requires --mode construct or search"
		}
		if len(options.searchLimitSet) > 0 && options.mode != "search" {
			return options, "--search-limit is only valid with --mode search"
		}
		if !seenGoal {
			return options, "algebra synthesize requires --goal <file>"
		}
		if len(options.metadata) == 0 {
			return options, "algebra synthesize requires at least one --metadata file"
		}
	}
	if options.command == "verify_goal" {
		if !seenGoal || !seenCandidate || !seenCertificate {
			return options, "algebra verify-goal requires --goal, --candidate, and --certificate named files"
		}
		if options.goal == "-" {
			return options, "--goal for verify-goal requires a named file path"
		}
		if len(options.metadata) == 0 {
			return options, "algebra verify-goal requires at least one --metadata file"
		}
		if len(options.searchLimitSet) != 0 {
			return options, "--search-limit is not valid for algebra verify-goal"
		}
	}
	if (options.command == "normalize" || options.command == "verify-rewrite") && len(options.metadata) == 0 {
		return options, "algebra normalize and verify-rewrite require at least one --metadata file"
	}
	if (options.command == "export" || options.command == "replay") && len(options.metadata) == 0 {
		return options, "algebra export and replay require at least one --metadata file"
	}
	return options, ""
}

func setSearchLimit(options *algebraOptions, assignment string) string {
	name, value, ok := strings.Cut(assignment, "=")
	if !ok || name == "" || value == "" {
		return "--search-limit must be name=value"
	}
	if options.searchLimitSet[name] {
		return fmt.Sprintf("duplicate --search-limit %q", name)
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return fmt.Sprintf("--search-limit %q must use a nonnegative decimal integer", name)
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Sprintf("--search-limit %q is outside the supported integer range", name)
	}
	var destination *int
	switch name {
	case "max_path_length":
		destination = &options.searchLimits.MaxPathLength
	case "max_branch_depth":
		destination = &options.searchLimits.MaxBranchDepth
	case "max_expansions":
		destination = &options.searchLimits.MaxExpansions
	case "max_candidates":
		destination = &options.searchLimits.MaxCandidates
	case "max_trace_entries":
		destination = &options.searchLimits.MaxTraceEntries
	case "max_work_units":
		destination = &options.searchLimits.MaxWorkUnits
	case "max_candidate_bytes":
		destination = &options.searchLimits.MaxCandidateBytes
	case "timeout_millis":
		destination = &options.searchLimits.TimeoutMillis
	default:
		return fmt.Sprintf("unknown --search-limit name %q", name)
	}
	*destination = parsed
	options.searchLimitSet[name] = true
	return ""
}

func setVerifyLimit(options *algebraOptions, assignment string) string {
	name, value, ok := strings.Cut(assignment, "=")
	if !ok || name == "" || value == "" {
		return "--verify-limit must be name=value"
	}
	if options.verifyLimitSet[name] {
		return fmt.Sprintf("duplicate --verify-limit %q", name)
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return fmt.Sprintf("--verify-limit %q must use a nonnegative decimal integer", name)
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Sprintf("--verify-limit %q is outside the supported integer range", name)
	}
	var destination *int
	switch name {
	case "max_certificate_bytes":
		destination = &options.verifyLimits.MaxCertificateBytes
	case "max_replay_steps":
		destination = &options.verifyLimits.MaxReplaySteps
	case "max_work_nodes":
		destination = &options.verifyLimits.MaxWorkNodes
	default:
		return fmt.Sprintf("unknown --verify-limit name %q", name)
	}
	*destination = parsed
	options.verifyLimitSet[name] = true
	return ""
}

func setComparisonLimit(options *algebraOptions, assignment string) string {
	name, value, ok := strings.Cut(assignment, "=")
	if !ok || name == "" || value == "" {
		return "--comparison-limit must be name=value"
	}
	if options.comparisonLimitSet[name] {
		return fmt.Sprintf("duplicate --comparison-limit %q", name)
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return fmt.Sprintf("--comparison-limit %q must use a nonnegative decimal integer", name)
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Sprintf("--comparison-limit %q is outside the supported integer range", name)
	}
	var destination *int
	switch name {
	case "max_work_units":
		destination = &options.comparisonLimits.MaxWorkUnits
	case "max_certificate_bytes":
		destination = &options.comparisonLimits.MaxCertificateBytes
	default:
		return fmt.Sprintf("unknown --comparison-limit name %q", name)
	}
	*destination = parsed
	options.comparisonLimitSet[name] = true
	return ""
}

func setEvaluationLimit(options *algebraOptions, assignment string) string {
	name, value, ok := strings.Cut(assignment, "=")
	if !ok || name == "" || value == "" {
		return "--evaluation-limit must be name=value"
	}
	if options.evaluationLimitSet[name] {
		return fmt.Sprintf("duplicate --evaluation-limit %q", name)
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return fmt.Sprintf("--evaluation-limit %q must use a nonnegative decimal integer", name)
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Sprintf("--evaluation-limit %q is outside the supported integer range", name)
	}
	var destination *int
	switch name {
	case "max_plan_atoms":
		destination = &options.evaluationLimits.MaxPlanAtoms
	case "max_variables":
		destination = &options.evaluationLimits.MaxVariables
	case "max_intermediate_rows":
		destination = &options.evaluationLimits.MaxIntermediateRows
	case "max_results":
		destination = &options.evaluationLimits.MaxResults
	case "max_proof_nodes":
		destination = &options.evaluationLimits.MaxProofNodes
	case "max_proofs_per_answer":
		destination = &options.evaluationLimits.MaxProofsPerAnswer
	case "max_proof_depth":
		destination = &options.evaluationLimits.MaxProofDepth
	case "max_work_units":
		destination = &options.evaluationLimits.MaxWorkUnits
	case "max_output_bytes":
		destination = &options.evaluationLimits.MaxOutputBytes
	case "timeout_millis":
		destination = &options.evaluationLimits.TimeoutMillis
	default:
		return fmt.Sprintf("unknown --evaluation-limit name %q", name)
	}
	*destination = parsed
	options.evaluationLimitSet[name] = true
	return ""
}

func runAlgebra(args []string, input io.Reader, stdout, stderr io.Writer) int {
	options, parseError := parseAlgebraArgs(args)
	if parseError != "" {
		return writeResult(stdout, stderr, algebraDiagnostic(status.OutcomeInvalid, "INVALID_ALGEBRA_ARGUMENTS", parseError, "algebra"))
	}
	if strings.HasPrefix(options.command, "binding_") {
		return runAlgebraBinding(options, input, stdout, stderr)
	}
	if options.command == "capabilities" {
		output := algebraCapabilitiesOutput{Status: status.OutcomeOK, Scopes: map[string]algebraScopeCapability{
			"not":           {CanRepresent: true, CanEvaluate: false},
			"modal":         {CanRepresent: true, CanEvaluate: false, Condition: "opaque possible/necessary/likely wrapper; no modal fact semantics"},
			"quantifier":    {CanRepresent: true, CanEvaluate: false},
			"ask":           {CanRepresent: true, CanEvaluate: true, Condition: "root Ask with a positive Atom/And/Exists body after beta lowering; answer projection is embedded in Ask"},
			"relative_time": {CanRepresent: true, CanEvaluate: false, Condition: "past/present are explicit reference attributes; no mapping to the query instant"},
			"multiplicity":  {CanRepresent: true, CanEvaluate: false, Condition: "one is accepted only in the atomic reference profile; many has no grouping semantics"},
		}, QuestionType: string(algebra.TypeQuestion), Evaluation: algebraScopeCapability{CanRepresent: true, CanEvaluate: true, Condition: "only the bounded positive Ask profile; not arbitrary Question values"}, Scope: "representation and evaluation capability only; missing attributes remain unknown; no world-truth or quantifier-completeness claim"}
		if err := writeJSON(stdout, output); err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra capabilities: %v\n", err)
			return ExitError
		}
		return ExitOK
	}
	if options.command == "metadata_inspect" || options.command == "metadata_validate" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		operation := strings.TrimPrefix(options.command, "metadata_")
		output := metadataCommandOutput{Status: status.OutcomeOK, Operation: operation, Packages: make([]metadataPackageOutput, 0, len(registries))}
		if operation == "validate" {
			output.ValidationScope = "positive-facts-v1 profile, manifest structure, dependency closure, declarations and types, constraints, provenance, and alias definitions; no theorem consistency, world-truth, or owner-acceptance claim"
		}
		for _, registry := range registries {
			output.Packages = append(output.Packages, metadataPackageSummary(registry, operation == "inspect"))
		}
		if err := writeJSON(stdout, output); err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize metadata result: %v\n", err)
			return ExitError
		}
		return ExitOK
	}
	if options.command == "context_check" {
		var registries []*metamodel.Registry
		if len(options.metadata) != 0 {
			var err error
			registries, err = readMetadataRegistries(options.metadata)
			if err != nil {
				return writeAlgebraError(stdout, stderr, err)
			}
		}
		return runAlgebraContextCheck(options, input, stdout, stderr, registries)
	}
	if options.command == "compare" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runAlgebraCompare(options, stdout, stderr, registries)
	}
	if options.command == "verify_proof" {
		switch options.claim {
		case "equivalence", "rewrite", "goal", "evaluation":
		default:
			result := algebraProofOutput{Status: status.OutcomeUnsupported, ClaimType: options.claim,
				Result:   map[string]any{"verdict": "undetermined", "diagnostics": []algebra.Error{{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_PROOF_CLAIM", Path: "claim", Message: "only equivalence, rewrite, goal, and evaluation proof claims are supported"}}},
				Metadata: []metadataReference{}, ValidationScope: "unsupported claim type; no equivalence or other proof interpretation was attempted"}
			return writeCappedAlgebraOutput(stdout, stderr, result, result.Status)
		}
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runAlgebraVerifyProof(options, input, stdout, stderr, registries)
	}
	if options.command == "facts_validate" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runNaryFactsValidate(options, input, stdout, stderr, registries)
	}
	if options.command == "evaluate" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runAlgebraEvaluate(options, stdout, stderr, registries)
	}
	if options.command == "synthesize" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runAlgebraSynthesize(options, input, stdout, stderr, registries)
	}
	if options.command == "verify_goal" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runAlgebraVerifyGoal(options, input, stdout, stderr, registries)
	}
	if options.command == "catalog_export" {
		return runAlgebraCatalogExport(stdout, stderr)
	}
	if options.command == "export" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runArtifactExport(options, input, stdout, stderr, registries)
	}
	if options.command == "inspect" {
		return runArtifactInspect(options, input, stdout, stderr)
	}
	if options.command == "replay" {
		registries, err := readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		return runArtifactReplay(options, input, stdout, stderr, registries)
	}
	var registries []*metamodel.Registry
	if len(options.metadata) != 0 {
		var err error
		registries, err = readMetadataRegistries(options.metadata)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	if options.command == "build" {
		return runAlgebraBuild(options, input, stdout, stderr, registries)
	}
	if options.command == "normalize" {
		return runAlgebraNormalize(options, input, stdout, stderr, registries)
	}
	if options.command == "verify-rewrite" {
		return runAlgebraVerifyRewrite(options, input, stdout, stderr, registries)
	}
	var reader io.Reader
	if options.input == "" || options.input == "-" {
		if input == nil {
			_, _ = fmt.Fprintln(stderr, "INFRASTRUCTURE_ERROR: stdin reader is unavailable")
			return ExitError
		}
		reader = input
	} else {
		file, err := os.Open(options.input)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not open algebra input: %v\n", err)
			return ExitError
		}
		defer file.Close()
		reader = file
	}
	data, err := io.ReadAll(io.LimitReader(reader, algebra.MaxDocumentBytes+1))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not read algebra input: %v\n", err)
		return ExitError
	}
	var document algebra.Document
	var checked algebra.CheckedExpression
	var canonical []byte
	if len(options.metadata) == 0 {
		document, err = algebra.Decode(data)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		checked, err = algebra.Check(document)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		canonical, err = algebra.Encode(document)
	} else {
		document, err = algebra.DecodeDocumentJSON(data)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		checked, document, err = metamodel.CheckDocument(document, registries)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		canonical, err = algebra.Encode(document)
	}
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraOutput{
		Status: checked.Status, Checked: checked,
		CanonicalDocument: json.RawMessage(canonical),
	}
	for _, registry := range registries {
		output.Metadata = append(output.Metadata, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
	}
	if checked.Status == status.OutcomeIncomplete {
		output.Diagnostics = []status.Diagnostic{{
			Code: "EXPRESSION_INCOMPLETE", Message: "typed holes remain unresolved",
			Path: "expression", Severity: status.SeverityWarning,
		}}
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra result: %v\n", err)
		return ExitError
	}
	return exitCode(checked.Status)
}

func runAlgebraBinding(options algebraOptions, input io.Reader, stdout, stderr io.Writer) int {
	packageBytes, err := readNamedAlgebraFile(options.languagePackage, language.MaxSemanticLanguagePackageBytes, "semantic language package")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	pack, err := language.LoadSemanticPackage(packageBytes, nil)
	if err != nil {
		return writeSemanticLanguageError(stdout, stderr, err)
	}
	packageOutput := summarizeSemanticBindingPackage(pack, packageBytes)
	if options.command == "binding_inspect" {
		metadata := []metadataReference{}
		if len(options.metadata) != 0 {
			registries, err := readMetadataRegistries(options.metadata)
			if err != nil {
				return writeAlgebraError(stdout, stderr, err)
			}
			for _, registry := range registries {
				metadata = append(metadata, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
			}
		}
		output := semanticBindingInspectOutput{
			Status: status.OutcomeOK, Package: packageOutput, MetadataVerified: false, Metadata: metadata,
			ValidationScope: "structural inspection of the explicitly supplied language package, authored frames, and tagged targets; selected metadata and world truth were not checked",
		}
		return writeCappedAlgebraOutput(stdout, stderr, output, status.OutcomeOK)
	}
	registries, err := readMetadataRegistries(options.metadata)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	selected, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	metadata := make([]metadataReference, 0, len(registries))
	for _, registry := range registries {
		metadata = append(metadata, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
	}
	if options.command == "binding_validate" {
		targets, err := pack.ValidateTargets(selected)
		if err != nil {
			return writeSemanticLanguageError(stdout, stderr, err)
		}
		resultStatus := status.OutcomeOK
		for _, target := range targets {
			if len(target.Unresolved) != 0 {
				resultStatus = status.OutcomeIncomplete
				break
			}
		}
		output := semanticBindingValidationOutput{
			Status: resultStatus, Package: packageOutput, Metadata: metadata, MetadataVerified: true,
			Targets:         targets,
			ValidationScope: "tagged targets checked against the explicitly selected approved primitive declarations; authorization is reported separately; no world-truth or owner-acceptance claim",
		}
		return writeCappedAlgebraOutput(stdout, stderr, output, resultStatus)
	}
	requestBytes, err := readAlgebraInput(options, input, language.MaxSemanticBindingRequestBytes, "semantic binding request")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request, err := language.DecodeSemanticBindingRequest(requestBytes)
	if err != nil {
		return writeSemanticLanguageError(stdout, stderr, err)
	}
	rewrite, err := metamodel.NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := language.InstantiateSemanticBinding(pack, selected, rewrite, request)
	if err != nil {
		return writeSemanticLanguageError(stdout, stderr, err)
	}
	output := struct {
		Status           status.OutcomeKind                   `json:"status"`
		Result           language.SemanticInstantiationResult `json:"result"`
		Metadata         []metadataReference                  `json:"metadata"`
		RegistrySnapshot string                               `json:"registry_snapshot_id"`
		ValidationScope  string                               `json:"validation_scope"`
	}{
		Status: result.Status, Result: result, Metadata: metadata, RegistrySnapshot: rewrite.SnapshotID,
		ValidationScope: "selected typed-target instantiation and independent goal verification when applicable; no facts, world truth, or language-quality claim",
	}
	return writeCappedAlgebraOutput(stdout, stderr, output, result.Status)
}

func summarizeSemanticBindingPackage(pack *language.SemanticLanguagePackage, packageBytes []byte) semanticBindingPackageOutput {
	digest := sha256.Sum256(packageBytes)
	output := semanticBindingPackageOutput{
		ID: pack.ID(), Version: pack.Version(), Language: string(pack.Language()),
		Fingerprint: "sha256:" + hex.EncodeToString(digest[:]), Provenance: pack.Provenance(),
		Frames: []semanticBindingFrameOutput{}, Targets: []semanticBindingTargetOutput{},
	}
	for _, frame := range pack.Frames() {
		row := semanticBindingFrameOutput{
			ID: string(frame.ID), SenseID: frame.SenseID.String(), SemanticRoles: make([]string, 0, len(frame.SemanticRoles)),
			Arguments: make([]semanticBindingArgumentOutput, 0, len(frame.Arguments)), AllowedFamilyIDs: make([]string, 0, len(frame.AllowedFamilyIDs)),
			Provenance: frame.Provenance,
		}
		for _, role := range frame.SemanticRoles {
			row.SemanticRoles = append(row.SemanticRoles, string(role))
		}
		for _, argument := range frame.Arguments {
			row.Arguments = append(row.Arguments, semanticBindingArgumentOutput{
				ID: argument.ID, Role: string(argument.Role), Required: argument.Required,
				GrammaticalFunction: argument.GrammaticalFunction, Case: argument.Case, Adposition: argument.Adposition,
			})
		}
		for _, family := range frame.AllowedFamilyIDs {
			row.AllowedFamilyIDs = append(row.AllowedFamilyIDs, family.String())
		}
		output.Frames = append(output.Frames, row)
	}
	for _, binding := range pack.LexiconIndex().Document().Bindings {
		if binding.Target == nil {
			continue
		}
		row := semanticBindingTargetOutput{
			SenseID: binding.SenseID.String(), Mapping: string(binding.Mapping), Status: string(binding.Status),
			Authorized: binding.TargetExecutionAuthorized(), Provenance: binding.Provenance,
		}
		data := &semanticBindingTargetData{
			Kind: string(binding.Target.Kind), Profile: binding.Target.Profile,
			RolePorts: make([]semanticBindingRolePort, 0, len(binding.Target.RolePorts)),
		}
		data.SymbolRef = string(binding.Target.SymbolRef)
		if binding.Target.GoalSpec != nil {
			data.Goal = binding.Target.GoalSpec
		}
		for _, port := range binding.Target.RolePorts {
			data.RolePorts = append(data.RolePorts, semanticBindingRolePort{FrameID: port.FrameID, Role: port.Role, PortID: port.PortID, Type: port.Type})
		}
		row.Target = data
		output.Targets = append(output.Targets, row)
	}
	return output
}

func writeSemanticLanguageError(stdout, stderr io.Writer, err error) int {
	var resultError interface{ Result() status.Result }
	if errors.As(err, &resultError) {
		result := resultError.Result()
		if len(result.Diagnostics) != 0 {
			diagnostic := result.Diagnostics[0]
			return writeResult(stdout, stderr, algebraDiagnostic(result.Status, diagnostic.Code, diagnostic.Message, diagnostic.Path))
		}
		return writeResult(stdout, stderr, result)
	}
	return writeAlgebraError(stdout, stderr, err)
}

func runAlgebraContextCheck(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, algebra.MaxDocumentBytes, "algebra context check")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	document, err := algebra.DecodeDocumentJSON(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if document.RequestContext == nil {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "MISSING_REQUEST_CONTEXT", Path: "request_context", Message: "algebra context check requires an explicit request_context in the document"})
	}
	var checked algebra.CheckedExpression
	if len(registries) == 0 {
		checked, err = algebra.Check(document)
	} else {
		checked, document, err = metamodel.CheckDocument(document, registries)
	}
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	assessment, err := algebra.CheckRequestContext(*document.RequestContext, document)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	canonical, err := algebra.Encode(document)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraContextCheckOutput{
		Status: checked.Status, Checked: checked, CanonicalDocument: json.RawMessage(canonical),
		Assessment: assessment, Metadata: metadataReferences(registries),
		ValidationScope: "typed document and request-local context structure; mention status derives only from explicit resolution evidence and conditions are preserved without evaluating their truth",
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra context assessment: %v\\n", err)
		return ExitError
	}
	return exitCode(output.Status)
}

func runAlgebraSynthesize(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	goalInput := options
	goalInput.input = options.goal
	data, err := readAlgebraInput(goalInput, input, synthesis.MaxGoalInputBytes, "algebra synthesize goal")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal, err := synthesis.DecodeGoal(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goalLimits, err := synthesis.NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal.Limits = goalLimits
	if options.mode == "search" {
		return runAlgebraSearch(goal, options.searchLimits, stdout, stderr, registries)
	}
	primitiveContext, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion, Limits: goalLimits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := synthesis.Construct(primitiveContext, goal)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	outputStatus := result.Status
	if result.Status == status.OutcomeOK && result.Checked != nil && result.CanonicalDocument != nil {
		rewriteEnvironment, envErr := metamodel.NormalizationEnvironment(algebra.Document{
			SchemaVersion: algebra.SchemaVersion, References: goal.References, RequestContext: goal.RequestContext, Limits: goalLimits.TermLimits,
		}, registries)
		if envErr != nil {
			return writeAlgebraError(stdout, stderr, envErr)
		}
		certificate, _, certErr := synthesis.CreateSynthesisCertificate(
			primitiveContext, rewriteEnvironment, goal, goal, []synthesis.ChoiceAssignment{},
			synthesis.CertificateOriginConstruct, result, synthesis.GoalVerificationLimits{TermLimits: goalLimits.TermLimits},
		)
		if certErr != nil {
			return writeAlgebraError(stdout, stderr, certErr)
		}
		verification, verifyErr := synthesis.VerifyGoalCertificate(
			primitiveContext, rewriteEnvironment, goal, *result.CanonicalDocument, certificate,
			synthesis.GoalVerificationLimits{TermLimits: goalLimits.TermLimits},
		)
		if verifyErr != nil {
			return writeAlgebraError(stdout, stderr, verifyErr)
		}
		result.Certificate = &certificate
		result.Verification = &verification
		result.FullyValidated = goalVerificationAccepted(verification)
		if result.FullyValidated {
			result.Status = status.OutcomeOK
		} else if verification.Status == status.OutcomeOK {
			outputStatus = status.OutcomeInvalid
		} else {
			outputStatus = verification.Status
		}
	}
	output := synthesisOutput{
		Status: outputStatus, Mode: "construct", GoalProfile: goal.Profile, ContextID: goal.ContextID,
		Metadata: metadataReferences(registries), Result: result,
		ValidationScope: "finite positive-facts construction with independent goal-constraint, operation-replay, and type verification; not world truth, search uniqueness, or language approval",
	}
	if result.FullyValidated && result.Status == status.OutcomeOK && result.Checked != nil && result.CanonicalDocument != nil {
		output.ExportRequest = &algebra.ArtifactExportRequest{
			SchemaVersion: algebra.SchemaVersion, Document: *result.CanonicalDocument,
			Profile:   algebra.RewriteProfile{ID: "dse-construct-artifact-v1", Rules: []algebra.RewriteRule{}},
			ContextID: goal.ContextID, ContextDependencies: []algebra.ContextDependency{},
			Provenance: append([]provenance.SourceRef(nil), goal.Provenance...),
		}
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra synthesis result: %v\n", err)
		return ExitError
	}
	return exitCode(outputStatus)
}

const maxAlgebraEvaluationResponseBytes = 16 << 20

func runAlgebraEvaluate(options algebraOptions, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	limits, err := reason.NormalizeAlgebraEvaluationLimits(options.evaluationLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	readNamed := func(path string, maximum int, description string) ([]byte, error) {
		file, openErr := os.Open(path)
		if openErr != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not open %s input %q: %v", description, path, openErr)}
		}
		defer file.Close()
		data, readErr := io.ReadAll(io.LimitReader(file, int64(maximum)+1))
		if readErr != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not read %s input %q: %v", description, path, readErr)}
		}
		if len(data) > maximum {
			return nil, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "JSON_BYTE_LIMIT", Path: description, Message: fmt.Sprintf("input exceeds %d byte limit", maximum)}
		}
		return data, nil
	}
	expressionBytes, err := readNamed(options.expression, algebra.MaxDocumentBytes, "expression")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	source, err := algebra.DecodeDocumentJSON(expressionBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if source.RequestContext != nil && source.RequestContext.ID != options.contextID {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "QUERY_REQUEST_CONTEXT_ID_MISMATCH", Path: "request_context.id", Message: "source request context must match the explicit query context ID"})
	}
	askSource := source.Expression != nil && source.Expression.Kind == algebra.TermAsk
	if askSource && options.selectionSpecified {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "ASK_SELECTION_FORBIDDEN", Path: "--select", Message: "Ask answer variables are the sole projection authority"})
	}
	if !askSource && !options.selectionSpecified {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "QUERY_SELECTION_REQUIRED", Path: "--select", Message: "non-Ask queries require an explicit answer selection"})
	}
	factsBytes, err := readNamed(options.factsInput, facts.MaxNaryInputBytes, "facts")
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	factInput, err := facts.DecodeNaryInput(factsBytes)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	var selection []reason.AlgebraSelection
	if options.selectionSpecified {
		selectionBytes, err := readNamed(options.selection, algebra.MaxDocumentBytes, "selection")
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		selection, err = reason.DecodeAlgebraSelections(selectionBytes)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	if factInput.ContextID != options.contextID {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "FACT_CONTEXT_MISMATCH", Path: "facts.context_id", Message: "--context-id must match the explicit fact input context"})
	}
	selectedAt := factInput.At
	if options.at != "" {
		flagTime, timeErr := parseEvaluationInstant(options.at, "--at")
		if timeErr != nil {
			return writeAlgebraError(stdout, stderr, timeErr)
		}
		if factInput.At != "" {
			factTime, parseErr := parseEvaluationInstant(factInput.At, "facts.at")
			if parseErr != nil {
				return writeNaryFactsError(stdout, stderr, parseErr)
			}
			if !flagTime.Equal(factTime) {
				return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "FACT_INSTANT_MISMATCH", Path: "--at", Message: "--at conflicts with the explicit fact input instant"})
			}
		} else {
			selectedAt = options.at
			factInput.At = selectedAt
		}
	}
	schema, err := metamodel.FactsSchema(registries)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	view, err := facts.BuildNary(schema, factInput)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	environment, err := metamodel.NormalizationEnvironment(algebra.Document{
		SchemaVersion: algebra.SchemaVersion, Parameters: source.Parameters,
		References: source.References, RequestContext: source.RequestContext, Limits: limits.TermLimits,
	}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request := reason.AlgebraQueryRequest{
		SchemaVersion: reason.AlgebraQuerySchemaVersion, Profile: options.profile,
		ContextID: options.contextID, At: selectedAt, Select: selection,
	}
	plan, err := reason.CompileAlgebraQuery(context.Background(), schema, environment, source, request, limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := reason.EvaluateAlgebraQuery(context.Background(), plan, view, limits)
	if err != nil {
		return writeNaryFactsOrAlgebraError(stdout, stderr, err)
	}
	output := algebraEvaluationOutput{
		Status: result.Status, Result: result, Metadata: metadataReferences(registries),
		ValidationScope: "finite positive query evaluation over the supplied typed fact snapshot and explicit known-reference domain; negative evidence is reported separately, absence remains unknown, and no world-completeness claim is made",
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if len(encoded)+1 > maxAlgebraEvaluationResponseBytes {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "EVALUATION_RESPONSE_BYTES", Path: "result", Message: "evaluation response envelope exceeds the 16 MiB output limit"})
	}
	if _, err := stdout.Write(append(encoded, '\n')); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write algebra evaluation result: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

func parseEvaluationInstant(value, path string) (time.Time, error) {
	if !strings.HasSuffix(value, "Z") {
		return time.Time{}, &algebra.Error{Status: status.OutcomeInvalid, Code: "INVALID_EVALUATION_INSTANT", Path: path, Message: "instant must be RFC3339 UTC with a Z suffix"}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, &algebra.Error{Status: status.OutcomeInvalid, Code: "INVALID_EVALUATION_INSTANT", Path: path, Message: "instant must be RFC3339 UTC with a Z suffix"}
	}
	return parsed.UTC(), nil
}

func writeNaryFactsOrAlgebraError(stdout, stderr io.Writer, err error) int {
	var typed *facts.StatusError
	if errors.As(err, &typed) {
		return writeNaryFactsError(stdout, stderr, err)
	}
	return writeAlgebraError(stdout, stderr, err)
}

func goalVerificationAccepted(result synthesis.GoalVerificationResult) bool {
	return result.Status == status.OutcomeOK && result.Disposition == synthesis.GoalVerificationDispositionChecked &&
		result.Complete && result.TypeCorrect && result.RefinementChecked && result.ReplayChecked &&
		result.CertificateValid && result.GoalSatisfied
}

func runAlgebraVerifyGoal(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	verifyLimits, err := synthesis.NormalizeGoalVerificationLimits(options.verifyLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goalOptions := options
	goalOptions.input = options.goal
	goalBytes, err := readAlgebraInput(goalOptions, input, synthesis.MaxGoalInputBytes, "algebra verify-goal goal")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal, err := synthesis.DecodeGoal(goalBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goalLimits, err := synthesis.NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal.Limits = goalLimits

	candidateOptions := options
	candidateOptions.input = options.candidate
	candidateBytes, err := readAlgebraInput(candidateOptions, input, algebra.MaxDocumentBytes, "algebra verify-goal candidate")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	candidate, err := algebra.DecodeDocumentJSON(candidateBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	certificateOptions := options
	certificateOptions.input = options.certificate
	certificateBytes, err := readAlgebraInput(certificateOptions, input, verifyLimits.MaxCertificateBytes, "algebra verify-goal certificate")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	certificate, err := synthesis.DecodeSynthesisCertificate(certificateBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}

	primitiveContext, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion, Limits: goalLimits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	rewriteEnvironment, err := metamodel.NormalizationEnvironment(algebra.Document{
		SchemaVersion: algebra.SchemaVersion, References: goal.References, RequestContext: goal.RequestContext, Limits: goalLimits.TermLimits,
	}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	verification, err := synthesis.VerifyGoalCertificate(primitiveContext, rewriteEnvironment, goal, candidate, certificate, verifyLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if verification.Status == status.OutcomeOK && !goalVerificationAccepted(verification) {
		verification.Status = status.OutcomeInvalid
		verification.Disposition = synthesis.GoalVerificationDispositionRejected
		verification.CertificateValid = false
		verification.Diagnostics = append(verification.Diagnostics, algebra.Error{
			Status: status.OutcomeInvalid, Code: "GOAL_VERIFICATION_INCOMPLETE",
			Path: "result", Message: "all independent verification flags are required for acceptance",
		})
	}
	output := goalVerificationOutput{
		Status: verification.Status, Result: verification, Metadata: metadataReferences(registries),
		ValidationScope: "independent finite positive-facts structural checking and operation replay; not world truth, uniqueness, or language approval",
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize goal verification result: %v\n", err)
		return ExitError
	}
	return exitCode(output.Status)
}

func runAlgebraSearch(goal synthesis.GoalSpec, requested synthesis.SearchLimits, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	goalLimits, err := synthesis.NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal.Limits = goalLimits
	primitiveContext, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion, Limits: goalLimits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	rewriteEnvironment, err := metamodel.NormalizationEnvironment(algebra.Document{
		SchemaVersion:  algebra.SchemaVersion,
		References:     goal.References,
		RequestContext: goal.RequestContext,
		Limits:         goalLimits.TermLimits,
	}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := synthesis.Search(context.Background(), primitiveContext, rewriteEnvironment, goal, requested)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := synthesisSearchOutput{
		Status: result.Status, Mode: "search", GoalProfile: goal.Profile, ContextID: goal.ContextID,
		Metadata: metadataReferences(registries), Result: result,
		ValidationScope: "finite declared positive-facts path profile; independent goal-constraint and operation-replay verification plus alpha identity; not world truth, speaker intent, uniqueness, or language approval",
	}
	return writeSynthesisSearchOutput(stdout, stderr, output)
}

const maxSynthesisSearchResponseBytes = 16 << 20

func writeSynthesisSearchOutput(stdout, stderr io.Writer, output synthesisSearchOutput) int {
	encode := func() ([]byte, error) { return json.Marshal(output) }
	encoded, err := encode()
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	for len(encoded)+1 > maxSynthesisSearchResponseBytes && len(output.Result.Candidates) > 0 {
		last := output.Result.Candidates[len(output.Result.Candidates)-1]
		candidateBytes, marshalErr := json.Marshal(last)
		if marshalErr != nil {
			return writeAlgebraError(stdout, stderr, marshalErr)
		}
		output.Result.Candidates = output.Result.Candidates[:len(output.Result.Candidates)-1]
		output.ResponseCandidatesDiscarded++
		output.Result.Usage.Candidates = len(output.Result.Candidates)
		output.Result.Usage.CandidateBytes -= len(candidateBytes)
		if output.Result.Usage.CandidateBytes < 0 {
			return writeAlgebraError(stdout, stderr, &algebra.Error{
				Status: status.OutcomeError, Code: "SEARCH_CANDIDATE_BYTE_ACCOUNTING",
				Path: "result.usage.candidate_bytes", Message: "candidate byte accounting became inconsistent",
			})
		}
		output.Result.Status = status.OutcomeResourceLimit
		output.Status = status.OutcomeResourceLimit
		output.Result.Outcome = synthesis.SearchOutcomeIncomplete
		output.Result.Complete = false
		output.Result.UniqueWithinDeclaredSpace = false
		output.Result.StopReason = synthesis.SearchStopReason("max_response_bytes")
		output.Result.ExhaustedBounds = appendUniqueSearchBound(output.Result.ExhaustedBounds, output.Result.StopReason)
		encoded, err = encode()
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	if len(encoded)+1 > maxSynthesisSearchResponseBytes {
		return writeAlgebraError(stdout, stderr, &algebra.Error{
			Status: status.OutcomeResourceLimit, Code: "SEARCH_RESPONSE_BYTES",
			Path: "result", Message: "search response envelope exceeds the 16 MiB output limit",
		})
	}
	if _, err := stdout.Write(append(encoded, '\n')); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write algebra search result: %v\n", err)
		return ExitError
	}
	return exitCode(output.Status)
}

func appendUniqueSearchBound(bounds []synthesis.SearchStopReason, reason synthesis.SearchStopReason) []synthesis.SearchStopReason {
	for _, existing := range bounds {
		if existing == reason {
			return bounds
		}
	}
	return append(bounds, reason)
}

func runAlgebraCatalogExport(stdout, stderr io.Writer) int {
	inventory, err := catalogexport.Export()
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := catalogExportOutput{
		Status:          status.OutcomeOK,
		ValidationScope: "inert catalog inventory; symbols remain opaque and pending; no operational approval, world-truth, or lexical-coverage claim",
		Inventory:       inventory,
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize catalog export: %v\n", err)
		return ExitError
	}
	if len(encoded)+1 > 16<<20 {
		return writeAlgebraError(stdout, stderr, &algebra.Error{
			Status: status.OutcomeResourceLimit, Code: "CATALOG_EXPORT_BYTES",
			Path: "inventory", Message: "catalog export envelope exceeds the 16 MiB output limit",
		})
	}
	if _, err := stdout.Write(append(encoded, '\n')); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write catalog export: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func runAlgebraNormalize(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, algebra.MaxDocumentBytes, "algebra normalize")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request, err := algebra.DecodeNormalizationRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	environment, err := metamodel.NormalizationEnvironment(request.Document, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.Normalize(environment, request.Profile, request.Document.Expression, request.Budget)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraNormalizeOutput{Status: result.Status, Result: result, Metadata: metadataReferences(registries)}
	if err := writeJSON(stdout, output); err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	return exitCode(result.Status)
}

func runNaryFactsValidate(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, facts.MaxNaryInputBytes, "algebra facts validate")
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	request, err := facts.DecodeNaryInput(data)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	schema, err := metamodel.FactsSchema(registries)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	view, err := facts.BuildNary(schema, request)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	// The request's explicit query instant is passed through. Empty means only
	// timeless evidence applies; the CLI never chooses the current time.
	summary, err := view.SummaryAt(view.At())
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	output := naryFactsOutput{
		Status: status.OutcomeOK, SnapshotID: view.SnapshotID(), ContextID: view.ContextID(), At: view.At(),
		Referents: view.Referents(), Facts: view.Facts(), Summary: summary,
		Metadata: metadataReferences(registries), AbsenceMeans: "unknown",
		ValidationScope: "explicit n-ary ground evidence; not world-truth or owner approval",
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize n-ary facts result: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func runAlgebraVerifyRewrite(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, algebra.MaxRewriteEnvelopeBytes, "algebra verify-rewrite")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	envelope, err := algebra.DecodeRewriteVerificationRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request := envelope.Request
	environment, err := metamodel.NormalizationEnvironment(request.Document, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.VerifyRewriteCertificate(environment, request.Profile, request.Document.Expression, envelope.Certificate, request.Budget)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraRewriteVerificationOutput{Status: result.Status, Result: result, Metadata: metadataReferences(registries)}
	if err := writeJSON(stdout, output); err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	return exitCode(result.Status)
}

func readAlgebraInput(options algebraOptions, input io.Reader, maximum int, description string) ([]byte, error) {
	var reader io.Reader
	if options.input == "" || options.input == "-" {
		if input == nil {
			return nil, &algebraInfrastructureError{message: "stdin reader is unavailable"}
		}
		reader = input
	} else {
		file, err := os.Open(options.input)
		if err != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not open %s input: %v", description, err)}
		}
		defer file.Close()
		reader = file
	}
	data, err := io.ReadAll(io.LimitReader(reader, int64(maximum)+1))
	if err != nil {
		return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not read %s input: %v", description, err)}
	}
	if len(data) > maximum {
		return nil, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "JSON_BYTE_LIMIT", Path: "input", Message: fmt.Sprintf("input exceeds %d byte limit", maximum)}
	}
	return data, nil
}

func runArtifactExport(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, algebra.MaxArtifactExportRequestBytes, "algebra export")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request, err := algebra.DecodeArtifactExportRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	environment, err := metamodel.ArtifactEnvironment(registries, request.Profile, request.ContextID, request.ContextDependencies, request.Provenance, request.Document.RequestContext)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	var cache *algebra.ArtifactCache
	if !options.noCache {
		cache, err = algebra.NewArtifactCache(0, 0)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	artifact, hit, err := algebra.ExportArtifact(environment, request.Document, cache)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	checked := artifact.Arena().Checked()
	output := artifactExportOutput{
		Status: checked.Status, Artifact: artifact.Envelope(), Checked: checked,
		UniqueNodes: artifact.Arena().Count(), ExpandedNodes: checked.NodeCount,
		Interning: !options.noCache, CacheHit: hit, Metadata: metadataReferences(registries),
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize artifact export: %v\n", err)
		return ExitError
	}
	return exitCode(checked.Status)
}

func runArtifactInspect(options algebraOptions, input io.Reader, stdout, stderr io.Writer) int {
	data, err := readAlgebraInput(options, input, algebra.MaxArtifactEnvelopeBytes, "algebra inspect")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	envelope, err := algebra.DecodeArtifactEnvelope(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	// The scoped envelope marshaler checks graph integrity and its self-reported
	// receipts against the envelope's declared content. This remains explicitly
	// structural and does not validate metadata approval or request authority.
	if _, err := json.Marshal(envelope); err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	context := algebra.Document{
		SchemaVersion: algebra.SchemaVersion, Sorts: envelope.Sorts, Signatures: envelope.Signatures,
		Parameters: envelope.Parameters, References: envelope.References, RequestContext: envelope.RequestContext, Closed: envelope.Closed, Limits: envelope.Limits,
	}
	arena, err := algebra.NewArena(context, envelope.Root, envelope.Nodes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	checked := arena.Checked()
	document, err := arena.Document()
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := artifactInspectionOutput{
		Status:            checked.Status,
		ValidationScope:   "structural artifact integrity against declared content only; reference attributes and request-context values are displayed as unverified claims; selected metadata, approval, identity authority, and external resources are not verified",
		MetadataVerified:  false,
		NormalizerVersion: envelope.NormalizerVersion, RegistrySnapshotID: envelope.RegistrySnapshotID,
		RequiredRegistries: append([]algebra.RegistryPin{}, envelope.Registries...), Profile: envelope.Profile,
		ContextID: envelope.ContextID, RequestContext: document.RequestContext, References: append([]algebra.Declaration{}, document.References...), ContextDependencies: append([]algebra.ContextDependency{}, envelope.ContextDependencies...),
		Provenance:          append([]provenance.SourceRef{}, envelope.Provenance...),
		ResourceFingerprint: envelope.ResourceFingerprint, MeaningFingerprint: envelope.MeaningFingerprint, CacheKey: envelope.CacheKey,
		Root: arena.Root(), UniqueNodes: arena.Count(), ExpandedNodes: checked.NodeCount, Checked: checked,
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize artifact inspection: %v\n", err)
		return ExitError
	}
	return exitCode(checked.Status)
}

func runArtifactReplay(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, algebra.MaxArtifactReplayRequestBytes, "algebra replay")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request, err := algebra.DecodeArtifactReplayRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	expected, err := metamodel.ArtifactEnvironment(registries, request.Profile, request.ContextID, request.ContextDependencies, request.Provenance, request.RequestContext)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	artifact, err := algebra.ReplayArtifact(expected, request.Artifact, request.Limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	checked := artifact.Arena().Checked()
	output := artifactReplayOutput{
		Status: checked.Status, Artifact: artifact.Envelope(), Checked: checked,
		UniqueNodes: artifact.Arena().Count(), ExpandedNodes: checked.NodeCount,
		Metadata: metadataReferences(registries),
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize artifact replay: %v\n", err)
		return ExitError
	}
	return exitCode(checked.Status)
}

func metadataReferences(registries []*metamodel.Registry) []metadataReference {
	refs := make([]metadataReference, 0, len(registries))
	for _, registry := range registries {
		refs = append(refs, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
	}
	return refs
}

func runAlgebraBuild(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	if len(registries) == 0 {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "MISSING_METADATA", Path: "metadata", Message: "algebra build requires at least one metadata package"})
	}
	var reader io.Reader
	if options.input == "" || options.input == "-" {
		if input == nil {
			_, _ = fmt.Fprintln(stderr, "INFRASTRUCTURE_ERROR: stdin reader is unavailable")
			return ExitError
		}
		reader = input
	} else {
		file, err := os.Open(options.input)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not open algebra build input: %v\n", err)
			return ExitError
		}
		defer file.Close()
		reader = file
	}
	data, err := io.ReadAll(io.LimitReader(reader, algebra.MaxDocumentBytes+1))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not read algebra build input: %v\n", err)
		return ExitError
	}
	request, err := algebra.DecodeBuildRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	context, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.Build(context, request)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if result.CanonicalDocument != nil && result.Expression != nil {
		admission := *result.CanonicalDocument
		admission.Sorts = nil
		admission.Signatures = nil
		checked, hydrated, err := metamodel.CheckDocument(admission, registries)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
		result.Status = checked.Status
		result.Checked = &checked
		for i := range hydrated.Signatures {
			if hydrated.Signatures[i].Ports == nil {
				hydrated.Signatures[i].Ports = []algebra.Port{}
			}
		}
		result.CanonicalDocument = &hydrated
	}
	resultKind := ""
	for _, record := range result.Trace {
		if record.ID == result.ResultID {
			resultKind = record.ResultKind
			break
		}
	}
	output := algebraBuildOutput{BuildResult: result, ResultKind: resultKind}
	if resultKind == "substitution" {
		substitution := result.Substitution
		if substitution == nil {
			substitution = algebra.Substitution{}
		}
		output.Substitution = &substitution
	}
	for _, registry := range registries {
		output.Metadata = append(output.Metadata, metadataReference{ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint()})
	}
	if result.Status == status.OutcomeIncomplete && result.Checked != nil && len(result.Checked.Holes) > 0 {
		output.Diagnostics = []status.Diagnostic{{Code: "EXPRESSION_INCOMPLETE", Message: "typed holes remain unresolved", Path: "result.expression", Severity: status.SeverityWarning}}
	}
	if err := writeJSON(stdout, output); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize algebra build result: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

func readMetadataRegistries(paths []string) ([]*metamodel.Registry, error) {
	if len(paths) == 0 {
		return nil, &algebra.Error{Status: status.OutcomeInvalid, Code: "MISSING_METADATA", Path: "metadata", Message: "at least one metadata file is required"}
	}
	if len(paths) > 16 {
		return nil, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: "metadata", Message: "metadata file count exceeds limit 16"}
	}
	inputs := make([][]byte, 0, len(paths))
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not open metadata file %q: %v", path, err)}
		}
		data, readErr := io.ReadAll(io.LimitReader(file, algebra.MaxDocumentBytes+1))
		closeErr := file.Close()
		if readErr != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not read metadata file %q: %v", path, readErr)}
		}
		if closeErr != nil {
			return nil, &algebraInfrastructureError{message: fmt.Sprintf("could not close metadata file %q: %v", path, closeErr)}
		}
		inputs = append(inputs, data)
	}
	registries, err := metamodel.LoadPackages(inputs)
	if err != nil {
		return nil, err
	}
	return registries, nil
}

func metadataPackageSummary(registry *metamodel.Registry, includeSymbols bool) metadataPackageOutput {
	document := registry.Document()
	result := metadataPackageOutput{
		ID: registry.ID(), Version: registry.Version(), Profile: document.Profile,
		ResourceFingerprint: registry.ResourceFingerprint(), FormalDenotationID: registry.FormalDenotationID(),
		SortCount: len(document.Sorts), SymbolCount: len(document.Symbols),
		Provenance: append([]provenance.SourceRef(nil), document.Provenance...),
	}
	for _, symbol := range document.Symbols {
		switch symbol.Kind {
		case metamodel.SymbolPrimitive:
			result.PrimitiveCount++
		case metamodel.SymbolAlias:
			result.AliasCount++
		case metamodel.SymbolOpaqueUnformalized:
			result.OpaqueCount++
		}
		if symbol.ReviewStatus == metamodel.ReviewPending {
			result.PendingCount++
		} else if symbol.ReviewStatus == metamodel.ReviewApproved {
			result.ApprovedCount++
		}
		if includeSymbols {
			result.Symbols = append(result.Symbols, metadataSymbolOutput{ID: symbol.ID, Kind: symbol.Kind, Label: symbol.Label, Description: symbol.Description, ReviewStatus: symbol.ReviewStatus, Provenance: append([]provenance.SourceRef(nil), symbol.Provenance...)})
		}
	}
	return result
}

func writeAlgebraError(stdout, stderr io.Writer, err error) int {
	var infrastructure *algebraInfrastructureError
	if errors.As(err, &infrastructure) {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: %s\n", infrastructure.message)
		return ExitError
	}
	var typed *algebra.Error
	if errors.As(err, &typed) {
		return writeResult(stdout, stderr, algebraDiagnostic(typed.Status, typed.Code, typed.Message, typed.Path))
	}
	_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: algebra check failed unexpectedly: %v\n", err)
	return ExitError
}

func writeNaryFactsError(stdout, stderr io.Writer, err error) int {
	var typed *facts.StatusError
	if errors.As(err, &typed) {
		result := typed.Result()
		if len(result.Diagnostics) > 0 {
			diagnostic := result.Diagnostics[0]
			return writeResult(stdout, stderr, algebraDiagnostic(result.Status, diagnostic.Code, diagnostic.Message, diagnostic.Path))
		}
		return writeResult(stdout, stderr, result)
	}
	return writeAlgebraError(stdout, stderr, err)
}

func runAlgebraCompare(options algebraOptions, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	leftBytes, err := readNamedAlgebraFile(options.left, algebra.MaxDocumentBytes, "comparison left")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	rightBytes, err := readNamedAlgebraFile(options.right, algebra.MaxDocumentBytes, "comparison right")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	profileBytes, err := readNamedAlgebraFile(options.profile, algebra.MaxDocumentBytes, "comparison profile")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if len(leftBytes)+len(rightBytes)+len(profileBytes) > algebra.MaxEquivalenceBytes {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "EQUIVALENCE_INPUT_BYTES", Path: "inputs", Message: "combined comparison inputs exceed the 16 MiB request cap"})
	}
	left, err := algebra.DecodeDocumentJSON(leftBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	right, err := algebra.DecodeDocumentJSON(rightBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	profile, err := algebra.DecodeRewriteProfileJSON(profileBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	limits, err := algebra.NormalizeEquivalenceLimits(options.comparisonLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	parameters, err := mergeComparisonDeclarations(left.Parameters, right.Parameters, "parameters")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	references, err := mergeComparisonDeclarations(left.References, right.References, "references")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	environment, err := metamodel.NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion, Parameters: parameters, References: references, RequestContext: left.RequestContext, Limits: limits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.CompareExpressions(environment, left, right, profile, options.contextID, limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraEquivalenceOutput{Status: result.Status, ClaimType: algebra.EquivalenceClaimProfile, Result: result, Metadata: metadataReferences(registries), ValidationScope: "positive equivalence only relative to the exact selected rewrite profile, metadata snapshot, context ID, and the two supplied expressions; different normal forms remain undetermined"}
	return writeCappedAlgebraOutput(stdout, stderr, output, output.Status)
}

func mergeComparisonDeclarations(left, right []algebra.Declaration, path string) ([]algebra.Declaration, error) {
	byID := make(map[string]algebra.Type, len(left)+len(right))
	for _, declaration := range append(append([]algebra.Declaration{}, left...), right...) {
		if !validAlgebraCLIText(declaration.ID, 256) {
			return nil, &algebra.Error{Status: status.OutcomeInvalid, Code: "INVALID_COMPARISON_DECLARATION", Path: path, Message: "source declaration ID is empty or exceeds its bound"}
		}
		if existing, ok := byID[declaration.ID]; ok {
			if !algebra.TypeEqual(existing, declaration.Type) {
				return nil, &algebra.Error{Status: status.OutcomeInvalid, Code: "COMPARISON_DECLARATION_TYPE_CONFLICT", Path: path + "." + declaration.ID, Message: "the same parameter or reference ID has different types across comparison inputs"}
			}
			continue
		}
		byID[declaration.ID] = declaration.Type
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]algebra.Declaration, 0, len(ids))
	for _, id := range ids {
		result = append(result, algebra.Declaration{ID: id, Type: byID[id]})
	}
	return result, nil
}

func validAlgebraCLIText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value)
}

func readNamedAlgebraFile(path string, maximum int, description string) ([]byte, error) {
	if path == "" || path == "-" {
		return nil, &algebra.Error{Status: status.OutcomeInvalid, Code: "NAMED_INPUT_REQUIRED", Path: description, Message: "this input requires a named file path"}
	}
	options := algebraOptions{input: path}
	return readAlgebraInput(options, nil, maximum, description)
}

func writeCappedAlgebraOutput(stdout, stderr io.Writer, output any, kind status.OutcomeKind) int {
	encoded, err := json.Marshal(output)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if len(encoded)+1 > maxAlgebraEvaluationResponseBytes {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "PROOF_RESPONSE_BYTES", Path: "result", Message: "complete proof response exceeds the 16 MiB output limit"})
	}
	if _, err := stdout.Write(append(encoded, '\n')); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write algebra proof response: %v\n", err)
		return ExitError
	}
	return exitCode(kind)
}

func runAlgebraVerifyProof(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	switch options.claim {
	case "equivalence":
		return runAlgebraVerifyEquivalence(options, stdout, stderr, registries)
	case "rewrite":
		return runAlgebraVerifyRewriteClaim(options, input, stdout, stderr, registries)
	case "goal":
		return runAlgebraVerifyGoalClaim(options, input, stdout, stderr, registries)
	case "evaluation":
		return runAlgebraVerifyEvaluation(options, stdout, stderr, registries)
	default:
		return ExitUnsupported
	}
}

func runAlgebraVerifyEquivalence(options algebraOptions, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	leftBytes, err := readNamedAlgebraFile(options.left, algebra.MaxDocumentBytes, "equivalence left")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	rightBytes, err := readNamedAlgebraFile(options.right, algebra.MaxDocumentBytes, "equivalence right")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	profileBytes, err := readNamedAlgebraFile(options.profile, algebra.MaxDocumentBytes, "equivalence profile")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	limits, err := algebra.NormalizeEquivalenceLimits(options.comparisonLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	certificateBytes, err := readNamedAlgebraFile(options.certificate, limits.MaxCertificateBytes, "equivalence certificate")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if len(leftBytes)+len(rightBytes)+len(profileBytes)+len(certificateBytes) > algebra.MaxEquivalenceBytes {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "EQUIVALENCE_INPUT_BYTES", Path: "inputs", Message: "combined equivalence proof inputs exceed the 16 MiB request cap"})
	}
	left, err := algebra.DecodeDocumentJSON(leftBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	right, err := algebra.DecodeDocumentJSON(rightBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	profile, err := algebra.DecodeRewriteProfileJSON(profileBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	certificate, err := algebra.DecodeEquivalenceCertificate(certificateBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	parameters, err := mergeComparisonDeclarations(left.Parameters, right.Parameters, "parameters")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	references, err := mergeComparisonDeclarations(left.References, right.References, "references")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	environment, err := metamodel.NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion, Parameters: parameters, References: references, RequestContext: left.RequestContext, Limits: limits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.VerifyEquivalenceCertificate(environment, left, right, profile, options.contextID, certificate, limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraProofOutput{Status: result.Status, ClaimType: algebra.EquivalenceClaimProfile, Result: result, Metadata: metadataReferences(registries), ValidationScope: "independent positive profile-equivalence certificate verification; no general non-equivalence, entailment, or countermodel claim"}
	return writeCappedAlgebraOutput(stdout, stderr, output, output.Status)
}

func runAlgebraVerifyRewriteClaim(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	data, err := readAlgebraInput(options, input, algebra.MaxRewriteEnvelopeBytes, "rewrite proof request")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	envelope, err := algebra.DecodeRewriteVerificationRequest(data)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request := envelope.Request
	environment, err := metamodel.NormalizationEnvironment(request.Document, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := algebra.VerifyRewriteCertificate(environment, request.Profile, request.Document.Expression, envelope.Certificate, request.Budget)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	output := algebraProofOutput{Status: result.Status, ClaimType: "rewrite", Result: result, Metadata: metadataReferences(registries), ValidationScope: "independent verification of one submitted rewrite certificate under its explicit licensed profile"}
	return writeCappedAlgebraOutput(stdout, stderr, output, output.Status)
}

func runAlgebraVerifyGoalClaim(options algebraOptions, input io.Reader, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	verifyLimits, err := synthesis.NormalizeGoalVerificationLimits(options.verifyLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goalBytes, err := readNamedAlgebraFile(options.goal, synthesis.MaxGoalInputBytes, "goal")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal, err := synthesis.DecodeGoal(goalBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goalLimits, err := synthesis.NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	goal.Limits = goalLimits
	candidateBytes, err := readNamedAlgebraFile(options.candidate, algebra.MaxDocumentBytes, "candidate")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	candidate, err := algebra.DecodeDocumentJSON(candidateBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	certificateBytes, err := readNamedAlgebraFile(options.certificate, verifyLimits.MaxCertificateBytes, "goal certificate")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	certificate, err := synthesis.DecodeSynthesisCertificate(certificateBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	primitiveContext, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion, Limits: goalLimits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	rewriteEnvironment, err := metamodel.NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion, References: goal.References, RequestContext: goal.RequestContext, Limits: goalLimits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	result, err := synthesis.VerifyGoalCertificate(primitiveContext, rewriteEnvironment, goal, candidate, certificate, verifyLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if result.Status == status.OutcomeOK && !goalVerificationAccepted(result) {
		result.Status = status.OutcomeInvalid
		result.Disposition = synthesis.GoalVerificationDispositionRejected
		result.CertificateValid = false
		result.Diagnostics = append(result.Diagnostics, algebra.Error{Status: status.OutcomeInvalid, Code: "GOAL_VERIFICATION_INCOMPLETE", Path: "result", Message: "all independent verification flags are required for acceptance"})
	}
	output := algebraProofOutput{Status: result.Status, ClaimType: "goal", Result: result, Metadata: metadataReferences(registries), ValidationScope: "independent finite positive-facts structural checking and operation replay; not world truth, uniqueness, or language approval"}
	return writeCappedAlgebraOutput(stdout, stderr, output, output.Status)
}

func runAlgebraVerifyEvaluation(options algebraOptions, stdout, stderr io.Writer, registries []*metamodel.Registry) int {
	limits, err := reason.NormalizeAlgebraEvaluationLimits(options.evaluationLimits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	expressionBytes, err := readNamedAlgebraFile(options.expression, algebra.MaxDocumentBytes, "expression")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	factBytes, err := readNamedAlgebraFile(options.factsInput, facts.MaxNaryInputBytes, "facts")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	var selectionBytes []byte
	if options.selectionSpecified {
		selectionBytes, err = readNamedAlgebraFile(options.selection, algebra.MaxDocumentBytes, "selection")
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	proofBytes, err := readNamedAlgebraFile(options.proof, reason.MaxAlgebraOutputBytes, "evaluation proof")
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if len(expressionBytes)+len(factBytes)+len(selectionBytes)+len(proofBytes) > algebra.MaxEquivalenceBytes {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "EVALUATION_PROOF_INPUT_BYTES", Path: "inputs", Message: "combined evaluation proof inputs exceed the 16 MiB request cap"})
	}
	source, err := algebra.DecodeDocumentJSON(expressionBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if source.RequestContext != nil && source.RequestContext.ID != options.contextID {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "QUERY_REQUEST_CONTEXT_ID_MISMATCH", Path: "request_context.id", Message: "source request context must match the explicit query context ID"})
	}
	askSource := source.Expression != nil && source.Expression.Kind == algebra.TermAsk
	if askSource && options.selectionSpecified {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "ASK_SELECTION_FORBIDDEN", Path: "--select", Message: "Ask answer variables are the sole projection authority"})
	}
	if !askSource && !options.selectionSpecified {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "QUERY_SELECTION_REQUIRED", Path: "--select", Message: "non-Ask queries require an explicit answer selection"})
	}
	factInput, err := facts.DecodeNaryInput(factBytes)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	var selection []reason.AlgebraSelection
	if options.selectionSpecified {
		selection, err = reason.DecodeAlgebraSelections(selectionBytes)
		if err != nil {
			return writeAlgebraError(stdout, stderr, err)
		}
	}
	claim, err := reason.DecodeAlgebraEvaluationProof(proofBytes)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	if factInput.ContextID != options.contextID {
		return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "FACT_CONTEXT_MISMATCH", Path: "facts.context_id", Message: "--context-id must match the explicit fact input context"})
	}
	selectedAt := factInput.At
	if options.at != "" {
		flagTime, timeErr := parseEvaluationInstant(options.at, "--at")
		if timeErr != nil {
			return writeAlgebraError(stdout, stderr, timeErr)
		}
		if factInput.At != "" {
			factTime, parseErr := parseEvaluationInstant(factInput.At, "facts.at")
			if parseErr != nil {
				return writeNaryFactsError(stdout, stderr, parseErr)
			}
			if !flagTime.Equal(factTime) {
				return writeAlgebraError(stdout, stderr, &algebra.Error{Status: status.OutcomeInvalid, Code: "FACT_INSTANT_MISMATCH", Path: "--at", Message: "--at conflicts with the explicit fact input instant"})
			}
		} else {
			selectedAt, factInput.At = options.at, options.at
		}
	}
	schema, err := metamodel.FactsSchema(registries)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	view, err := facts.BuildNary(schema, factInput)
	if err != nil {
		return writeNaryFactsError(stdout, stderr, err)
	}
	environment, err := metamodel.NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion, Parameters: source.Parameters, References: source.References, RequestContext: source.RequestContext, Limits: limits.TermLimits}, registries)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	request := reason.AlgebraQueryRequest{SchemaVersion: reason.AlgebraQuerySchemaVersion, Profile: options.profile, ContextID: options.contextID, At: selectedAt, Select: selection}
	plan, err := reason.CompileAlgebraQuery(context.Background(), schema, environment, source, request, limits)
	if err != nil {
		return writeAlgebraError(stdout, stderr, err)
	}
	verification, err := reason.VerifyAlgebraEvaluationProof(context.Background(), plan, view, claim, limits)
	if err != nil {
		return writeNaryFactsOrAlgebraError(stdout, stderr, err)
	}
	output := algebraProofOutput{Status: verification.Status, ClaimType: "evaluation", Result: verification, Metadata: metadataReferences(registries), ValidationScope: "independent checking of supplied positive answer witnesses against the compiled query and active fact snapshot; completeness is never verified"}
	return writeCappedAlgebraOutput(stdout, stderr, output, output.Status)
}

func algebraDiagnostic(kind status.OutcomeKind, code, message, path string) status.Result {
	message = strings.ToValidUTF8(message, "�")
	path = strings.ToValidUTF8(path, "�")
	if len(path) > status.MaxDiagnosticPathBytes {
		path = truncateAlgebraText(path, status.MaxDiagnosticPathBytes)
		const truncationNotice = " (diagnostic path was truncated)"
		message = truncateAlgebraText(message, status.MaxDiagnosticMessageBytes-len(truncationNotice)) + truncationNotice
	} else {
		message = truncateAlgebraText(message, status.MaxDiagnosticMessageBytes)
	}
	return diagnosticResult(kind, code, message, path)
}

func truncateAlgebraText(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	const marker = "..."
	limit := maxBytes - len(marker)
	for limit > 0 && !utf8.RuneStart(value[limit]) {
		limit--
	}
	return value[:limit] + marker
}

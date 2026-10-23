package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

const parseRootCategory construction.SlotCategory = "clause"
const parserMaxTokens = 256

// Parse matches the supplied phrase evidence against every analysis-capable
// compiled plan for the lattice language. It retains all bounded derivations.
func Parse(lattice analyzer.LexicalLattice, candidates []PhraseCandidate, context ParseContext, index *Index, budget Budget) (Forest, error) {
	if err := budget.Validate(); err != nil {
		return Forest{}, err
	}
	if index == nil || len(index.plans) == 0 {
		return Forest{}, newParserError("index", "parser index is required")
	}
	if err := validateLattice(lattice); err != nil {
		return Forest{}, err
	}
	if err := validateParseContext(context); err != nil {
		return Forest{}, err
	}
	forest := Forest{Language: lattice.Language, Outcome: status.OutcomeOK, Diagnostics: append([]status.Diagnostic(nil), lattice.Diagnostics...)}
	if len(forest.Diagnostics) > status.MaxDiagnostics-2 {
		forest.Diagnostics = forest.Diagnostics[:status.MaxDiagnostics-2]
		forest.Diagnostics = append(forest.Diagnostics, diag("LATTICE_DIAGNOSTICS_TRUNCATED", "additional lexical diagnostics were omitted to preserve parser outcome details", "lattice.diagnostics"))
	}
	if lattice.Outcome != "" && lattice.Outcome != status.OutcomeOK {
		forest.Outcome = lattice.Outcome
	}
	if len(candidates) > budget.MaxExpansions {
		forest.Outcome = status.OutcomeIncomplete
		forest.Diagnostics = appendDiagnosticOnce(forest.Diagnostics, diag("RESOURCE_LIMIT", "candidate inventory exceeds parser expansion budget", "candidates"))
		return forest, nil
	}
	readingIndex := indexLatticeReadings(lattice)
	byID := make(map[string]int, len(candidates))
	headSenses := make(map[string]map[ids.SenseID]bool)
	evidenceItems := 0
	for i := range candidates {
		evidenceItems += len(candidates[i].ReadingIDs) + len(candidates[i].SenseIDs) + len(candidates[i].Features)
		if evidenceItems > 65536 {
			return Forest{}, newCompileLimitError("candidate.evidence", evidenceItems, 65536)
		}
		if err := validateCandidate(candidates[i], i, lattice, byID, readingIndex); err != nil {
			return Forest{}, err
		}
		byID[candidates[i].ID] = i
		if candidates[i].Category == construction.CategoryVP {
			headSenses[candidates[i].ID] = candidateHeadSenseSet(candidates[i], lattice)
		}
	}
	if len(lattice.Tokens) == 0 {
		if forest.Outcome == status.OutcomeOK {
			forest.Outcome = status.OutcomeUnknown
		}
		forest.Diagnostics = appendDiagnosticOnce(forest.Diagnostics, diag("EMPTY_LATTICE", "no tokens are available to parse", ""))
		return forest, nil
	}
	if len(index.byLanguage[lattice.Language]) == 0 {
		forest.Outcome = status.OutcomeUnsupported
		forest.Diagnostics = appendDiagnosticOnce(forest.Diagnostics, diag("NO_LANGUAGE_PLANS", "no compiled parser plans for lattice language", string(lattice.Language)))
		forest.Unmatched = unmatchedCandidates(candidates, "no compiled plans for lattice language")
		return forest, nil
	}
	planIDs := plansStartingAtZero(index, lattice, candidates)
	if len(planIDs) == 0 {
		forest.Outcome = status.OutcomeUnknown
		forest.Diagnostics = appendDiagnosticOnce(forest.Diagnostics, diag("NO_START_PLAN", "no compiled plan can begin with token zero evidence", string(lattice.Language)))
		forest.Unmatched = unmatchedCandidates(candidates, "no compiled plan starts with this token evidence")
		return forest, nil
	}
	state := matcher{forest: &forest, lattice: lattice, candidates: candidates, byID: byID, headSenses: headSenses, context: context, budget: budget, semantic: map[string]bool{}, used: map[string]bool{}, rootByKey: map[string]int{}, phraseByKey: map[string]int{}, unsupportedHeads: map[string]bool{}, matchedHeadRanges: map[[2]int]bool{}}
	for _, pi := range planIDs {
		if !state.expand() {
			break
		}
		if pi < 0 || pi >= len(index.plans) {
			return Forest{}, newParserError("index.by_language", "compiled plan index is invalid")
		}
		state.matchPlan(index.plans[pi])
	}
	if state.exhausted {
		state.markIncomplete("RESOURCE_LIMIT", "parser search budget exhausted")
	}
	state.finishUnmatched()
	if len(forest.Roots) == 0 && forest.Outcome == status.OutcomeOK {
		forest.Outcome = status.OutcomeUnknown
		forest.Diagnostics = append(forest.Diagnostics, diag("NO_PARSE", "no complete construction matched the supplied evidence", ""))
	}
	if forest.Outcome == status.OutcomeOK && state.semanticCount > 1 {
		forest.Outcome = status.OutcomeAmbiguous
		forest.Diagnostics = append(forest.Diagnostics, diag("AMBIGUOUS_PARSE", "multiple sense and role interpretations remain", ""))
	}
	return forest, nil
}

type assignment struct {
	candidateIDs []string
	gap          *WHGapEvidence
	role         meaning.Role
}
type partial struct {
	slots      map[string]assignment
	unresolved []string
	used       map[string]bool
}
type matcher struct {
	forest               *Forest
	lattice              analyzer.LexicalLattice
	candidates           []PhraseCandidate
	byID                 map[string]int
	headSenses           map[string]map[ids.SenseID]bool
	context              ParseContext
	budget               Budget
	expansions           int
	retainedAlternatives int
	exhausted            bool
	semantic             map[string]bool
	semanticCount        int
	used                 map[string]bool
	rootByKey            map[string]int
	phraseByKey          map[string]int
	unsupportedHeads     map[string]bool
	matchedHeadRanges    map[[2]int]bool
}

func validateLattice(l analyzer.LexicalLattice) error {
	if !validParserLanguage(l.Language) {
		return newParserError("lattice.language", "lattice language is not a canonical bounded language tag")
	}
	if len(l.Tokens) > parserMaxTokens {
		return newCompileLimitError("lattice.tokens", len(l.Tokens), parserMaxTokens)
	}
	if l.Outcome != "" && !l.Outcome.Valid() {
		return newParserError("lattice.outcome", "lattice outcome is invalid")
	}
	if len(l.Diagnostics) > status.MaxDiagnostics {
		return newCompileLimitError("lattice.diagnostics", len(l.Diagnostics), status.MaxDiagnostics)
	}
	for _, d := range l.Diagnostics {
		if err := d.Validate(); err != nil {
			return newParserError("lattice.diagnostics", err.Error())
		}
	}
	lastEnd := -1
	readingCount := 0
	for i, node := range l.Tokens {
		if node.Index != i || node.Token.Surface == "" || node.Token.Normalized == "" || !utf8.ValidString(node.Token.Surface) || !utf8.ValidString(node.Token.Normalized) || node.Token.Span.Start < 0 || node.Token.Span.End > status.DefaultBounds().MaxInputBytes || node.Token.Span.End <= node.Token.Span.Start || node.Token.Span.Start < lastEnd || node.Token.Span.End-node.Token.Span.Start != len(node.Token.Surface) || !validTokenKind(node.Token.Kind) {
			return newParserError(fmt.Sprintf("lattice.tokens[%d]", i), "token index or source span is invalid")
		}
		lastEnd = node.Token.Span.End
		readingCount += len(node.Readings)
		if readingCount > 65536 {
			return newCompileLimitError("lattice.readings", readingCount, 65536)
		}
		for _, reading := range node.Readings {
			if reading.Reading.ID == "" || len(reading.Reading.ID) > 256 {
				return newParserError(fmt.Sprintf("lattice.tokens[%d].readings", i), "reading ID is empty or exceeds its bound")
			}
		}
	}
	return nil
}

func validTokenKind(kind text.TokenKind) bool {
	return kind == text.TokenWord || kind == text.TokenPunctuation || kind == text.TokenOther
}

func validateParseContext(context ParseContext) error {
	if len(context.WHGaps) > 256 {
		return newCompileLimitError("context.wh_gaps", len(context.WHGaps), 256)
	}
	if len(context.InformationStructure) > 256 {
		return newCompileLimitError("context.information_structure", len(context.InformationStructure), 256)
	}
	seen := map[string]bool{}
	for i, g := range context.WHGaps {
		if !validRoleName(string(g.Role)) || !validMeaningNodeID(string(g.ScopeNodeID)) || len(g.SlotID) > construction.MaxFamilyTextBytes || g.SlotID == "" {
			return newParserError(fmt.Sprintf("context.wh_gaps[%d]", i), "WH gap requires bounded slot, canonical role, and scope node IDs")
		}
		key := g.SlotID + "\x00" + string(g.Role) + "\x00" + string(g.ScopeNodeID)
		if seen[key] {
			return newParserError(fmt.Sprintf("context.wh_gaps[%d]", i), "duplicate WH gap evidence")
		}
		seen[key] = true
	}
	for role, value := range context.InformationStructure {
		if !validRoleName(string(role)) || (value != construction.InformationTopic && value != construction.InformationFocus && value != construction.InformationGiven && value != construction.InformationNew) {
			return newParserError("context.information_structure", "information structure requires canonical roles and known values")
		}
	}
	return nil
}

func unmatchedCandidates(candidates []PhraseCandidate, reason string) []UnmatchedCandidate {
	result := make([]UnmatchedCandidate, 0, len(candidates))
	for _, c := range candidates {
		result = append(result, UnmatchedCandidate{CandidateID: c.ID, TokenStart: c.TokenStart, TokenEnd: c.TokenEnd, Reason: reason})
	}
	return result
}

func validRoleName(value string) bool {
	if value == "" || len(value) > 64 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
func validMeaningNodeID(value string) bool {
	if value == "" || len(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func validConstructionToken(value string) bool {
	if value == "" || len(value) > construction.MaxFamilyTextBytes || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
func validFeatureToken(value string) bool {
	if value == "" || len(value) > features.MaxTokenBytes || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func validParserLanguage(language features.LanguageID) bool {
	value := string(language)
	if len(value) == 0 || len(value) > construction.MaxLanguageBytes || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, character := range part {
			if !(character >= 'a' && character <= 'z') && !(index > 0 && character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

type latticeReadingRef struct {
	tokenIndex int
	candidate  analyzer.ReadingCandidate
}

func indexLatticeReadings(l analyzer.LexicalLattice) map[string][]latticeReadingRef {
	result := make(map[string][]latticeReadingRef)
	for _, node := range l.Tokens {
		for _, reading := range node.Readings {
			result[reading.Reading.ID] = append(result[reading.Reading.ID], latticeReadingRef{tokenIndex: node.Index, candidate: reading})
		}
	}
	return result
}

func validateCandidate(c PhraseCandidate, i int, l analyzer.LexicalLattice, seen map[string]int, latticeReadings map[string][]latticeReadingRef) error {
	path := fmt.Sprintf("candidates[%d]", i)
	if c.ID == "" || len(c.ID) > 256 || strings.TrimSpace(c.ID) != c.ID {
		return newParserError(path+".id", "candidate ID is empty or malformed")
	}
	if _, ok := seen[c.ID]; ok {
		return newParserError(path+".id", "candidate ID is duplicated")
	}
	if !c.Category.Valid() || c.TokenStart < 0 || c.TokenEnd <= c.TokenStart || c.TokenEnd > len(l.Tokens) {
		return newParserError(path, "candidate category or token range is invalid")
	}
	if c.Category == construction.CategoryVP {
		if c.HeadTokenIndex < c.TokenStart || c.HeadTokenIndex >= c.TokenEnd {
			return newParserError(path+".head_token_index", "VP candidate head token is outside its span")
		}
	} else if c.HeadTokenIndex != 0 {
		return newParserError(path+".head_token_index", "only VP candidates may specify a head token")
	}
	if c.Span.Start != l.Tokens[c.TokenStart].Token.Span.Start || c.Span.End != l.Tokens[c.TokenEnd-1].Token.Span.End {
		return newParserError(path+".span", "candidate byte span does not match its half-open token range")
	}
	if len(c.ReadingIDs) == 0 || len(c.ReadingIDs) > 256 || len(c.SenseIDs) > 256 || len(c.Features) > features.MaxBundleValues {
		return newCompileLimitError(path+".evidence", len(c.ReadingIDs)+len(c.SenseIDs)+len(c.Features), 576)
	}
	seenRead := map[string]bool{}
	selectedReadings := make(map[string]analyzer.ReadingCandidate, len(c.ReadingIDs))
	for _, id := range c.ReadingIDs {
		refs, ok := latticeReadings[id]
		var r analyzer.ReadingCandidate
		found := false
		for _, ref := range refs {
			if ref.tokenIndex >= c.TokenStart && ref.tokenIndex < c.TokenEnd {
				r = ref.candidate
				found = true
				break
			}
		}
		if !ok || !found || seenRead[id] {
			return newParserError(path+".reading_ids", "candidate references a missing or duplicate lattice reading")
		}
		seenRead[id] = true
		selectedReadings[id] = r
		if c.LexemeID.String() != "" && r.Reading.LexemeID != c.LexemeID.String() {
			return newParserError(path+".lexeme_id", "candidate lexeme disagrees with its reading evidence")
		}
	}
	seenSense := map[ids.SenseID]bool{}
	for _, sid := range c.SenseIDs {
		if sid.String() == "" || seenSense[sid] {
			return newParserError(path+".sense_ids", "candidate contains an empty or duplicate sense ID")
		}
		seenSense[sid] = true
		found := false
		for _, rid := range c.ReadingIDs {
			for _, s := range selectedReadings[rid].Senses {
				if s.Sense.ID == sid {
					found = true
				}
			}
		}
		if !found {
			return newParserError(path+".sense_ids", "candidate references a sense absent from its reading evidence")
		}
	}
	for feature, value := range c.Features {
		if !validFeatureToken(string(feature)) || !validFeatureToken(string(value)) {
			return newParserError(path+".features", "candidate contains invalid feature evidence")
		}
	}
	if c.CaseKnown && !validConstructionToken(c.Case) {
		return newParserError(path+".case", "known case evidence must name a canonical case")
	}
	if !c.CaseKnown && c.Case != "" {
		return newParserError(path+".case", "unknown case evidence cannot carry a case value")
	}
	if featureCase, exists := c.Features[features.FeatureID("case")]; exists && c.CaseKnown && string(featureCase) != c.Case {
		return newParserError(path+".features.case", "case feature disagrees with the dedicated case evidence")
	}
	if _, exists := c.Features[features.FeatureID("case")]; exists && !c.CaseKnown {
		return newParserError(path+".features.case", "case feature evidence requires matching known dedicated case evidence")
	}
	return nil
}

func (m *matcher) expand() bool {
	m.expansions++
	if m.expansions > m.budget.MaxExpansions {
		m.exhausted = true
		return false
	}
	return true
}

func (m *matcher) matchPlan(p compiledPlan) {
	if len(p.sequence) == 0 {
		return
	}
	initial := partial{slots: map[string]assignment{}, unresolved: nil, used: map[string]bool{}}
	m.walk(p, 0, 0, initial)
}

func plansStartingAtZero(index *Index, lattice analyzer.LexicalLattice, candidates []PhraseCandidate) []int {
	if len(lattice.Tokens) == 0 {
		return nil
	}
	seen := make(map[int]bool)
	seenCategories := map[construction.SlotCategory]bool{}
	for _, c := range candidates {
		if c.TokenStart != 0 {
			continue
		}
		if seenCategories[c.Category] {
			continue
		}
		seenCategories[c.Category] = true
		for _, pi := range index.byFirstSymbol[planLookupKey{Language: lattice.Language, Category: c.Category}] {
			seen[pi] = true
		}
	}
	terminal := canonicalTerminal(lattice.Tokens[0].Token)
	for _, pi := range index.byFirstSymbol[planLookupKey{Language: lattice.Language, Terminal: terminal}] {
		seen[pi] = true
	}
	result := make([]int, 0, len(seen))
	for pi := range seen {
		result = append(result, pi)
	}
	sort.Ints(result)
	return result
}

func (m *matcher) walk(p compiledPlan, si, ti int, part partial) {
	if !m.expand() {
		return
	}
	if si == len(p.sequence) {
		if ti != len(m.lattice.Tokens) {
			return
		}
		if !m.checkPlan(p, &part) {
			return
		}
		m.retain(p, part)
		return
	}
	symbol := p.sequence[si]
	if symbol.Slot == nil {
		if ti < len(m.lattice.Tokens) && symbolMatchesInput(symbol, canonicalTerminal(m.lattice.Tokens[ti].Token)) {
			m.walk(p, si+1, ti+1, clonePartial(part))
		}
		if symbol.Optional {
			m.walk(p, si+1, ti, clonePartial(part))
		}
		return
	}
	slot := *symbol.Slot
	role := bindingRole(p, slot.ID)
	for _, c := range m.candidates {
		if !m.expand() {
			return
		}
		if c.TokenStart != ti || c.Category != slot.Category {
			continue
		}
		if c.TokenEnd > len(m.lattice.Tokens) {
			continue
		}
		if slot.ID == p.headSlotID && (!containsSense(c.SenseIDs, p.permission.SenseID) || !m.headSenses[c.ID][p.permission.SenseID]) {
			m.noteUnsupportedHead(c)
			continue
		}
		next := clonePartial(part)
		if !m.caseCompatible(p, slot, c, &next) {
			continue
		}
		next.slots[slot.ID] = assignment{candidateIDs: []string{c.ID}, role: role}
		next.used[c.ID] = true
		m.walk(p, si+1, c.TokenEnd, next)
	}
	// A cardinality maximum greater than one permits adjacent candidates to fill
	// a slot. Minimum is enforced below; no surface position supplies a role.
	if slot.Cardinality.Maximum > 1 {
		m.walkMulti(p, si, ti, slot, role, part)
	}
	if gap := m.findGap(slot.ID, role); gap != nil {
		next := clonePartial(part)
		next.slots[slot.ID] = assignment{gap: gap, role: role}
		m.walk(p, si+1, ti, next)
	}
	if slot.Optionality == construction.SlotOptional && slot.Cardinality.Minimum == 0 {
		next := clonePartial(part)
		next.slots[slot.ID] = assignment{role: role}
		m.walk(p, si+1, ti, next)
	}
}

func symbolMatchesInput(symbol compiledSymbol, value string) bool {
	if value == symbol.Terminal {
		return true
	}
	for _, accepted := range symbol.AcceptedInputs {
		if value == accepted {
			return true
		}
	}
	return false
}

func canonicalTerminal(token text.Token) string {
	if token.Folded != "" {
		return token.Folded
	}
	return token.Normalized
}

func (m *matcher) walkMulti(p compiledPlan, si, ti int, slot construction.SlotSpec, role meaning.Role, part partial) {
	var gather func(int, []string, partial)
	gather = func(cursor int, ids []string, current partial) {
		if !m.expand() {
			return
		}
		count := len(ids)
		if count >= slot.Cardinality.Minimum && count > 1 {
			next := clonePartial(current)
			next.slots[slot.ID] = assignment{candidateIDs: append([]string(nil), ids...), role: role}
			for _, id := range ids {
				next.used[id] = true
			}
			m.walk(p, si+1, cursor, next)
		}
		if count == slot.Cardinality.Maximum {
			return
		}
		for _, c := range m.candidates {
			if !m.expand() {
				return
			}
			if c.TokenStart != cursor || c.Category != slot.Category {
				continue
			}
			already := false
			for _, prior := range ids {
				if prior == c.ID {
					already = true
					break
				}
			}
			if already {
				continue
			}
			if slot.ID == p.headSlotID && (!containsSense(c.SenseIDs, p.permission.SenseID) || !m.headSenses[c.ID][p.permission.SenseID]) {
				m.noteUnsupportedHead(c)
				continue
			}
			nextPart := clonePartial(current)
			if !m.caseCompatible(p, slot, c, &nextPart) {
				continue
			}
			nextIDs := append(append([]string(nil), ids...), c.ID)
			gather(c.TokenEnd, nextIDs, nextPart)
		}
	}
	gather(ti, nil, part)
}

func (m *matcher) caseCompatible(p compiledPlan, slot construction.SlotSpec, c PhraseCandidate, part *partial) bool {
	wanted := slot.Case
	for _, b := range p.plan.ArgumentBindings {
		if b.SlotID == slot.ID && b.CasePolicy != "" {
			if wanted != "" && wanted != b.CasePolicy {
				return false
			}
			wanted = b.CasePolicy
		}
	}
	for _, constraint := range p.plan.Constraints {
		if constraint.Kind == construction.ConstraintGovern && constraint.Target.SlotID == slot.ID {
			if wanted != "" && wanted != constraint.CasePolicy {
				return false
			}
			wanted = constraint.CasePolicy
		}
	}
	if wanted == "" {
		return true
	}
	if !c.CaseKnown {
		part.unresolved = appendUnique(part.unresolved, slot.ID)
		return true
	}
	return c.Case == wanted
}

func (m *matcher) findGap(slot string, role meaning.Role) *WHGapEvidence {
	for i := range m.context.WHGaps {
		g := &m.context.WHGaps[i]
		if g.SlotID == slot && role != "" && g.Role == role && g.ScopeNodeID != "" {
			copy := *g
			return &copy
		}
	}
	return nil
}

func bindingRole(p compiledPlan, slot string) meaning.Role {
	for _, b := range p.plan.ArgumentBindings {
		if b.SlotID == slot {
			return meaning.Role(b.Role)
		}
	}
	return ""
}

func (m *matcher) checkPlan(p compiledPlan, part *partial) bool {
	for _, sym := range p.sequence {
		if sym.Slot == nil {
			continue
		}
		a, exists := part.slots[sym.Slot.ID]
		if !exists {
			return false
		}
		n := len(a.candidateIDs)
		if n < sym.Slot.Cardinality.Minimum || n > sym.Slot.Cardinality.Maximum {
			if a.gap == nil || n != 0 || sym.Slot.Cardinality.Minimum > 1 {
				return false
			}
			part.unresolved = appendUnique(part.unresolved, sym.Slot.ID)
		}
	}
	for _, condition := range p.variant.Conditions {
		switch condition.Kind {
		case construction.ConditionFeatureEquals:
			a := part.slots[condition.SlotID]
			value, known := commonFeature(m.candidates, m.byID, a.candidateIDs, condition.Feature)
			if !known {
				part.unresolved = appendUnique(part.unresolved, condition.SlotID)
			} else if value != condition.Value {
				return false
			}
		case construction.ConditionInformationStructure:
			value, known := m.context.InformationStructure[condition.Role]
			if !known {
				part.unresolved = appendUnique(part.unresolved, "condition:information_structure:"+string(condition.Role))
			} else if value != condition.Info {
				return false
			}
		}
	}
	for _, c := range p.plan.Constraints {
		switch c.Kind {
		case construction.ConstraintPrecede:
			left, lok := part.slots[c.BeforeSlot]
			right, rok := part.slots[c.AfterSlot]
			if lok && rok && len(left.candidateIDs) > 0 && len(right.candidateIDs) > 0 {
				if m.candidates[m.byID[left.candidateIDs[0]]].TokenStart >= m.candidates[m.byID[right.candidateIDs[0]]].TokenStart {
					return false
				}
			}
		case construction.ConstraintAgree:
			ok, known := agreement(m.candidates, m.byID, p, part, c)
			if !known {
				part.unresolved = appendUnique(part.unresolved, resolveConstraintSlot(p, c.Target), resolveConstraintSlot(p, c.Controller))
			} else if !ok && c.Strength == construction.ConstraintHard {
				return false
			}
		case construction.ConstraintRequire:
			ok, known := predicateValue(m.candidates, m.byID, p, part, *c.Condition)
			if !known {
				for _, sym := range p.sequence {
					if sym.Slot != nil {
						part.unresolved = appendUnique(part.unresolved, sym.Slot.ID)
					}
				}
			} else if !ok {
				return false
			}
		case construction.ConstraintOptional:
			if _, present := part.slots[c.OptionalSlot]; !present {
				return false
			}
		}
	}
	for _, s := range p.family.SlotSpecs {
		if s.Adposition != "" {
			if a := part.slots[s.ID]; len(a.candidateIDs) > 0 && !construction.VariantEmitsAdposition(p.variant, p.family, s.ID, s.Adposition) {
				part.unresolved = appendUnique(part.unresolved, s.ID)
			}
		}
	}
	return true
}

func (m *matcher) retain(p compiledPlan, part partial) {
	if len(m.forest.Nodes) >= m.budget.MaxNodes || m.retainedAlternatives >= m.budget.MaxAlternatives {
		m.exhausted = true
		return
	}
	rootKey := fmt.Sprintf("%d:%d", 0, len(m.lattice.Tokens))
	neededNodes := 0
	if _, exists := m.rootByKey[rootKey]; !exists {
		neededNodes++
	}
	phraseNeeded := map[string]bool{}
	for slotID, a := range part.slots {
		if len(a.candidateIDs) == 0 {
			continue
		}
		first := m.candidates[m.byID[a.candidateIDs[0]]]
		last := m.candidates[m.byID[a.candidateIDs[len(a.candidateIDs)-1]]]
		key := fmt.Sprintf("%s:%d:%d", p.slots[slotID].Category, first.TokenStart, last.TokenEnd)
		if _, exists := m.phraseByKey[key]; !exists && !phraseNeeded[key] {
			phraseNeeded[key] = true
			neededNodes++
		}
	}
	if len(m.forest.Nodes)+neededNodes > m.budget.MaxNodes {
		m.exhausted = true
		return
	}
	rootIndex, ok := m.rootByKey[rootKey]
	if !ok {
		span := text.Span{}
		if len(m.lattice.Tokens) > 0 {
			span = text.Span{Start: m.lattice.Tokens[0].Token.Span.Start, End: m.lattice.Tokens[len(m.lattice.Tokens)-1].Token.Span.End}
		}
		if len(m.forest.Nodes) >= m.budget.MaxNodes {
			m.exhausted = true
			return
		}
		rootIndex = len(m.forest.Nodes)
		m.rootByKey[rootKey] = rootIndex
		m.forest.Nodes = append(m.forest.Nodes, PackedNode{ID: NodeID("root:" + rootKey), Start: 0, End: len(m.lattice.Tokens), Span: span, Category: parseRootCategory})
		m.forest.Roots = append(m.forest.Roots, NodeID("root:"+rootKey))
	}
	bindings := make([]SlotBinding, 0, len(part.slots))
	keys := make([]string, 0, len(part.slots))
	for id := range part.slots {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		a := part.slots[id]
		binding := SlotBinding{SlotID: id, Role: a.role, WHGap: a.gap}
		if len(a.candidateIDs) > 0 {
			cat := p.slots[id].Category
			first := m.candidates[m.byID[a.candidateIDs[0]]]
			last := m.candidates[m.byID[a.candidateIDs[len(a.candidateIDs)-1]]]
			phraseKey := fmt.Sprintf("%s:%d:%d", cat, first.TokenStart, last.TokenEnd)
			pi, exists := m.phraseByKey[phraseKey]
			if !exists {
				if len(m.forest.Nodes) >= m.budget.MaxNodes {
					m.exhausted = true
					return
				}
				span := text.Span{Start: first.Span.Start, End: last.Span.End}
				pi = len(m.forest.Nodes)
				m.phraseByKey[phraseKey] = pi
				m.forest.Nodes = append(m.forest.Nodes, PackedNode{ID: NodeID("phrase:" + phraseKey), Start: first.TokenStart, End: last.TokenEnd, Span: span, Category: cat})
			}
			node := &m.forest.Nodes[pi]
			node.CandidateIDs = appendUnique(node.CandidateIDs, a.candidateIDs...)
			binding.CandidateIDs = append([]string(nil), a.candidateIDs...)
			binding.NodeIDs = []NodeID{node.ID}
			for _, cid := range a.candidateIDs {
				m.used[cid] = true
			}
			if id == p.headSlotID {
				m.matchedHeadRanges[[2]int{first.TokenStart, last.TokenEnd}] = true
			}
		}
		bindings = append(bindings, binding)
	}
	alt := ParseAlternative{PlanID: p.plan.ID, FamilyID: p.family.ID, VariantID: p.variant.ID, Permission: p.permission, Slots: bindings, Unresolved: append([]string(nil), part.unresolved...), DerivationID: derivationID(p, bindings)}
	root := &m.forest.Nodes[rootIndex]
	root.Alternatives = append(root.Alternatives, alt)
	m.retainedAlternatives++
	sem := semanticKey(p, bindings)
	if !m.semantic[sem] {
		m.semantic[sem] = true
		m.semanticCount++
	}
	if len(part.unresolved) > 0 {
		m.forest.Outcome = status.OutcomeIncomplete
		m.forest.Diagnostics = appendDiagnosticOnce(m.forest.Diagnostics, diag("UNRESOLVED_PARSE_EVIDENCE", "parse retained with unresolved feature or licensing evidence", ""))
	}
}

func (m *matcher) noteUnsupportedHead(c PhraseCandidate) {
	if m.unsupportedHeads == nil {
		m.unsupportedHeads = map[string]bool{}
	}
	m.unsupportedHeads[c.ID] = true
}
func (m *matcher) finishUnmatched() {
	used := m.used
	if used == nil {
		used = map[string]bool{}
	}
	usedNouns := make(map[string]bool)
	for _, candidate := range m.candidates {
		if used[candidate.ID] && candidate.Category == construction.CategoryNP {
			usedNouns[fmt.Sprintf("%d:%d:%s:%s", candidate.TokenStart, candidate.TokenEnd, candidate.LexemeID, senseSetKey(candidate.SenseIDs))] = true
		}
	}
	for _, c := range m.candidates {
		if used[c.ID] {
			continue
		}
		if c.Category == construction.CategoryNP && usedNouns[fmt.Sprintf("%d:%d:%s:%s", c.TokenStart, c.TokenEnd, c.LexemeID, senseSetKey(c.SenseIDs))] {
			// A different case reading for the same noun and sense is
			// rejected by the frame, not an unresolved lexical meaning.
			continue
		}
		reason := "not selected by a retained complete derivation"
		if m.unsupportedHeads[c.ID] {
			reason = "candidate lacks the plan's licensed predicate sense"
			if m.overlapsMatchedHead(c.TokenStart, c.TokenEnd) {
				m.markIncomplete("UNSUPPORTED_COMPETING_HEAD", "overlapping predicate evidence lacks an authorized plan sense")
			}
		}
		m.forest.Unmatched = append(m.forest.Unmatched, UnmatchedCandidate{CandidateID: c.ID, TokenStart: c.TokenStart, TokenEnd: c.TokenEnd, Reason: reason})
	}
}

func senseSetKey(senses []ids.SenseID) string {
	values := make([]string, len(senses))
	for index, sense := range senses {
		values[index] = sense.String()
	}
	sort.Strings(values)
	return strings.Join(values, "\x00")
}
func (m *matcher) overlapsMatchedHead(start, end int) bool {
	for span := range m.matchedHeadRanges {
		if start < span[1] && span[0] < end {
			return true
		}
	}
	return false
}
func (m *matcher) markIncomplete(code, message string) {
	m.forest.Outcome = status.OutcomeIncomplete
	m.forest.Diagnostics = appendDiagnosticOnce(m.forest.Diagnostics, diag(code, message, "parser"))
}

func clonePartial(p partial) partial {
	n := partial{slots: make(map[string]assignment, len(p.slots)), unresolved: append([]string(nil), p.unresolved...), used: make(map[string]bool, len(p.used))}
	for k, v := range p.slots {
		v.candidateIDs = append([]string(nil), v.candidateIDs...)
		n.slots[k] = v
	}
	for k, v := range p.used {
		n.used[k] = v
	}
	return n
}
func containsSense(values []ids.SenseID, value ids.SenseID) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func candidateHeadSenseSet(candidate PhraseCandidate, lattice analyzer.LexicalLattice) map[ids.SenseID]bool {
	result := map[ids.SenseID]bool{}
	if candidate.HeadTokenIndex < 0 || candidate.HeadTokenIndex >= len(lattice.Tokens) {
		return result
	}
	readingIDs := make(map[string]bool, len(candidate.ReadingIDs))
	for _, id := range candidate.ReadingIDs {
		readingIDs[id] = true
	}
	for _, reading := range lattice.Tokens[candidate.HeadTokenIndex].Readings {
		if !readingIDs[reading.Reading.ID] {
			continue
		}
		for _, sense := range reading.Senses {
			result[sense.Sense.ID] = true
		}
	}
	return result
}
func appendUnique(values []string, more ...string) []string {
	seen := map[string]bool{}
	for _, v := range values {
		seen[v] = true
	}
	for _, v := range more {
		if v != "" && !seen[v] {
			values = append(values, v)
			seen[v] = true
		}
	}
	return values
}
func commonFeature(candidates []PhraseCandidate, byID map[string]int, ids []string, feature features.FeatureID) (features.FeatureValueID, bool) {
	var value features.FeatureValueID
	for i, id := range ids {
		c := candidates[byID[id]]
		v, ok := c.Features[feature]
		if !ok || i > 0 && value != v {
			return "", false
		}
		value = v
	}
	return value, len(ids) > 0
}
func agreement(candidates []PhraseCandidate, byID map[string]int, plan compiledPlan, p *partial, c construction.CompiledConstraint) (bool, bool) {
	leftSlot := resolveConstraintSlot(plan, c.Controller)
	rightSlot := resolveConstraintSlot(plan, c.Target)
	if leftSlot == "" || rightSlot == "" {
		return false, false
	}
	left, ok1 := valueAt(candidates, byID, p.slots[leftSlot], c.Controller.Feature)
	right, ok2 := valueAt(candidates, byID, p.slots[rightSlot], c.Target.Feature)
	if !ok1 || !ok2 {
		return false, false
	}
	return left == right, true
}
func resolveConstraintSlot(p compiledPlan, source construction.SourcePath) string {
	if source.SlotID != "" {
		return source.SlotID
	}
	for _, binding := range p.plan.ArgumentBindings {
		if binding.ArgumentID == source.ArgumentID {
			return binding.SlotID
		}
	}
	return ""
}
func valueAt(candidates []PhraseCandidate, byID map[string]int, a assignment, f features.FeatureID) (features.FeatureValueID, bool) {
	if len(a.candidateIDs) == 0 {
		return "", false
	}
	var value features.FeatureValueID
	for i, id := range a.candidateIDs {
		c := candidates[byID[id]]
		v, ok := c.Features[f]
		if !ok {
			return "", false
		}
		if i > 0 && value != v {
			return "", false
		}
		value = v
	}
	return value, true
}
func predicateValue(candidates []PhraseCandidate, byID map[string]int, plan compiledPlan, p *partial, pred construction.Predicate) (bool, bool) {
	if pred.Kind == construction.PredicateSlotPresent {
		a, ok := p.slots[pred.SlotID]
		return ok && len(a.candidateIDs) > 0, ok
	}
	if pred.Kind == construction.PredicateFeatureEquals {
		slotID := resolveConstraintSlot(plan, pred.Source)
		a, ok := p.slots[slotID]
		if !ok {
			return false, false
		}
		v, known := valueAt(candidates, byID, a, pred.Source.Feature)
		return v == pred.Value, known
	}
	return false, false
}
func derivationID(p compiledPlan, slots []SlotBinding) string {
	var b strings.Builder
	b.WriteString(string(p.plan.ID))
	b.WriteByte('|')
	b.WriteString(p.permission.SenseID.String())
	for _, s := range slots {
		b.WriteByte('|')
		b.WriteString(s.SlotID)
		b.WriteByte('=')
		for _, id := range s.CandidateIDs {
			b.WriteString(id)
			b.WriteByte(',')
		}
		if s.WHGap != nil {
			b.WriteString("wh:")
			b.WriteString(string(s.WHGap.ScopeNodeID))
		}
	}
	digest := sha256.Sum256([]byte(b.String()))
	return "derivation:" + hex.EncodeToString(digest[:])
}
func semanticKey(p compiledPlan, slots []SlotBinding) string {
	var b strings.Builder
	b.WriteString(p.permission.SenseID.String())
	b.WriteByte('/')
	b.WriteString(string(p.permission.FrameID))
	for _, s := range slots {
		b.WriteByte('|')
		b.WriteString(string(s.Role))
		b.WriteByte('=')
		for _, id := range s.CandidateIDs {
			b.WriteString(id)
			b.WriteByte(',')
		}
		if s.WHGap != nil {
			b.WriteString("|wh:")
			b.WriteString(string(s.WHGap.Role))
			b.WriteByte(':')
			b.WriteString(string(s.WHGap.ScopeNodeID))
		}
	}
	return b.String()
}
func diag(code, message, path string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityWarning}
}
func appendDiagnosticOnce(values []status.Diagnostic, d status.Diagnostic) []status.Diagnostic {
	for _, v := range values {
		if v.Code == d.Code && v.Path == d.Path {
			return values
		}
	}
	if len(values) < status.MaxDiagnostics {
		return append(values, d)
	}
	if isPriorityParserDiagnostic(d.Code) {
		for i := len(values) - 1; i >= 0; i-- {
			if !isPriorityParserDiagnostic(values[i].Code) {
				values[i] = d
				return values
			}
		}
	}
	return values
}

func isPriorityParserDiagnostic(code string) bool {
	switch code {
	case "RESOURCE_LIMIT", "NO_LANGUAGE_PLANS", "NO_START_PLAN", "NO_PARSE", "UNRESOLVED_PARSE_EVIDENCE", "UNSUPPORTED_COMPETING_HEAD":
		return true
	default:
		return false
	}
}

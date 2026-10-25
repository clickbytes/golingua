package metamodel

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	maxPackages         = 16
	maxDeclarations     = 512
	maxIDBytes          = 256
	maxLabelBytes       = 1024
	maxDescriptionBytes = 2048
)

// Compile validates one manifest against the explicitly supplied resolved
// dependency closure (direct and transitive snapshots), then creates its
// immutable registry snapshot. Unrelated snapshots are rejected.
func Compile(document Document, dependencySnapshots []*Registry) (*Registry, error) {
	closure, direct, err := snapshotClosure(document, dependencySnapshots)
	if err != nil {
		return nil, err
	}
	context := make([]Document, 0, len(closure)+1)
	context = append(context, document)
	for _, registry := range closure {
		context = append(context, registry.Document())
	}
	if err := validateDocuments(context); err != nil {
		return nil, err
	}
	if err := validateDeclarationTypes(context); err != nil {
		return nil, err
	}
	if err := validatePackageReferences(document, context); err != nil {
		return nil, err
	}
	if err := validateAliasCycles(context); err != nil {
		return nil, err
	}
	if err := validateAliases(context); err != nil {
		return nil, err
	}
	return newValidatedRegistry(document, direct)
}

// LoadPackages decodes and validates an explicitly supplied package set.
// It returns no active prefix if any package or dependency fails.
func LoadPackages(inputs [][]byte) ([]*Registry, error) {
	if len(inputs) == 0 {
		return nil, invalid("EMPTY_PACKAGE_SET", "packages", "at least one package is required")
	}
	if len(inputs) > maxPackages {
		return nil, resource("RESOURCE_LIMIT", "packages", fmt.Sprintf("package count exceeds limit %d", maxPackages))
	}
	documents := make([]Document, 0, len(inputs))
	byID := make(map[string]Document, len(inputs))
	for i, input := range inputs {
		document, err := Decode(input)
		if err != nil {
			return nil, err
		}
		if _, exists := byID[document.ID]; exists {
			return nil, invalid("DUPLICATE_PACKAGE_ID", fmt.Sprintf("packages[%d].id", i), "package ID is duplicated")
		}
		byID[document.ID] = document
		documents = append(documents, document)
	}
	if err := validatePackageGraph(documents, byID); err != nil {
		return nil, err
	}
	if err := validateDocuments(documents); err != nil {
		return nil, err
	}
	for _, document := range documents {
		context := append([]Document{document}, dependencyClosureDocuments(document, byID)...)
		if err := validateDeclarationTypes(context); err != nil {
			return nil, err
		}
		if err := validatePackageReferences(document, context); err != nil {
			return nil, err
		}
	}
	if err := validateAliasCycles(documents); err != nil {
		return nil, err
	}
	if err := validateAliases(documents); err != nil {
		return nil, err
	}

	order := topologicalDocuments(documents, byID)
	registries := make(map[string]*Registry, len(order))
	result := make([]*Registry, 0, len(order))
	for _, document := range order {
		deps := make([]*Registry, 0, len(document.Dependencies))
		for _, dependency := range document.Dependencies {
			deps = append(deps, registries[dependency.ID])
		}
		registry, err := newValidatedRegistry(document, deps)
		if err != nil {
			return nil, invalid("REGISTRY_CONSTRUCTION_FAILED", "packages", err.Error())
		}
		registries[document.ID] = registry
		result = append(result, registry)
	}
	return result, nil
}

// CheckDocument hydrates a source-neutral algebra expression from explicitly
// selected registries and invokes the ordinary algebra checker.
func CheckDocument(document algebra.Document, registries []*Registry) (algebra.CheckedExpression, algebra.Document, error) {
	if len(registries) == 0 {
		checked, err := algebra.Check(document)
		return checked, document, err
	}
	hydrated, symbols, err := declarationContext(document, registries)
	if err != nil {
		return algebra.CheckedExpression{}, algebra.Document{}, err
	}
	if err := checkSelectedAtoms(document.Expression, symbols); err != nil {
		return algebra.CheckedExpression{}, algebra.Document{}, err
	}
	checked, err := algebra.Check(hydrated)
	return checked, hydrated, err
}

// DeclarationContext returns a checked declaration environment suitable for
// validating a real expression separately. It never synthesizes an AST node.
// Registry snapshots must be explicitly supplied and previously validated.
func DeclarationContext(document algebra.Document, registries []*Registry) (algebra.Document, error) {
	hydrated, _, err := declarationContext(document, registries)
	return hydrated, err
}

func declarationContext(document algebra.Document, registries []*Registry) (algebra.Document, map[string]Symbol, error) {
	if len(registries) == 0 {
		if err := algebra.CheckDeclarations(document); err != nil {
			return algebra.Document{}, nil, err
		}
		return document, map[string]Symbol{}, nil
	}
	if len(document.Sorts) != 0 || len(document.Signatures) != 0 {
		return algebra.Document{}, nil, invalid("INLINE_DECLARATIONS_WITH_METADATA", "document", "inline sorts and signatures are forbidden when metadata registries are selected")
	}
	if len(registries) > maxPackages {
		return algebra.Document{}, nil, resource("RESOURCE_LIMIT", "registries", "too many registries selected")
	}
	combinedSorts := make(map[string]algebra.SortDeclaration)
	combinedSymbols := make(map[string]Symbol)
	for i, registry := range registries {
		if registry == nil {
			return algebra.Document{}, nil, invalid("NIL_REGISTRY", fmt.Sprintf("registries[%d]", i), "registry is nil")
		}
		if registry.ID() == "" || registry.Version() == "" || registry.ResourceFingerprint() == "" || registry.FormalDenotationID() == "" {
			return algebra.Document{}, nil, invalid("UNVALIDATED_REGISTRY", fmt.Sprintf("registries[%d]", i), "registry has no validated identity and fingerprints")
		}
		for _, sortDecl := range registry.document.Sorts {
			if prior, exists := combinedSorts[sortDecl.ID]; exists {
				if prior != sortDecl {
					return algebra.Document{}, nil, invalid("SORT_COLLISION", "registries", "selected registries declare conflicting sort IDs")
				}
				return algebra.Document{}, nil, invalid("DUPLICATE_SORT_ID", "registries", "selected registries contain a duplicate sort ID")
			}
			combinedSorts[sortDecl.ID] = sortDecl
		}
		for id, symbol := range registry.symbols {
			if _, exists := combinedSymbols[id]; exists {
				return algebra.Document{}, nil, invalid("DUPLICATE_SYMBOL_ID", "registries", "selected registries contain a duplicate symbol ID")
			}
			combinedSymbols[id] = symbol
		}
	}
	for id := range combinedSorts {
		if _, conflict := combinedSymbols[id]; conflict {
			return algebra.Document{}, nil, invalid("SORT_SYMBOL_COLLISION", "registries", "sort and symbol IDs must not collide")
		}
	}
	if len(combinedSorts)+len(combinedSymbols)+len(document.Parameters)+len(document.References) > maxDeclarations {
		return algebra.Document{}, nil, resource("RESOURCE_LIMIT", "registries", "aggregate declaration count exceeds 512")
	}
	hydrated := document
	hydrated.Sorts = make([]algebra.SortDeclaration, 0, len(combinedSorts))
	for _, declaration := range combinedSorts {
		hydrated.Sorts = append(hydrated.Sorts, declaration)
	}
	sort.Slice(hydrated.Sorts, func(i, j int) bool { return hydrated.Sorts[i].ID < hydrated.Sorts[j].ID })
	hydrated.Signatures = make([]algebra.Signature, 0, len(combinedSymbols))
	for id, symbol := range combinedSymbols {
		if symbol.Kind != SymbolPrimitive || symbol.ReviewStatus != ReviewApproved {
			continue
		}
		hydrated.Signatures = append(hydrated.Signatures, algebra.Signature{ID: id, Label: symbol.Label, Ports: append([]algebra.Port(nil), symbol.Ports...)})
	}
	sort.Slice(hydrated.Signatures, func(i, j int) bool { return hydrated.Signatures[i].ID < hydrated.Signatures[j].ID })
	if err := algebra.CheckDeclarations(hydrated); err != nil {
		return algebra.Document{}, nil, err
	}
	return hydrated, combinedSymbols, nil
}

func validateDocuments(documents []Document) error {
	if len(documents) > maxPackages {
		return resource("RESOURCE_LIMIT", "packages", "package count exceeds limit 16")
	}
	seenPackage := make(map[string]struct{}, len(documents))
	seenRecord := make(map[string]string)
	aggregateDeclarations := 0
	for docIndex, document := range documents {
		base := fmt.Sprintf("packages[%d]", docIndex)
		if document.SchemaVersion != DocumentSchemaVersion {
			return invalid("INVALID_SCHEMA_VERSION", base+".schema_version", "unsupported manifest version")
		}
		if !validText(document.ID, maxIDBytes) {
			return invalid("INVALID_PACKAGE_ID", base+".id", "package ID is invalid")
		}
		if !validText(document.Version, maxIDBytes) {
			return invalid("INVALID_PACKAGE_VERSION", base+".version", "package version is invalid")
		}
		if document.Profile != PositiveFactsProfile {
			return unsupported("UNSUPPORTED_PROFILE", base+".profile", "manifest profile is not supported")
		}
		if _, exists := seenPackage[document.ID]; exists {
			return invalid("DUPLICATE_PACKAGE_ID", base+".id", "package ID is duplicated")
		}
		seenPackage[document.ID] = struct{}{}
		if len(document.Sorts)+len(document.Symbols) > maxDeclarations {
			return resource("RESOURCE_LIMIT", base, "package sort and symbol count exceeds 512")
		}
		aggregateDeclarations += len(document.Sorts) + len(document.Symbols)
		if aggregateDeclarations > maxDeclarations {
			return resource("RESOURCE_LIMIT", "packages", "aggregate sort and symbol count exceeds 512")
		}
		if len(document.Dependencies) > 16 {
			return resource("RESOURCE_LIMIT", base+".dependencies", "dependency count exceeds 16")
		}
		if len(document.Provenance) > 512 {
			return resource("RESOURCE_LIMIT", base+".provenance", "provenance count exceeds 512")
		}
		for i, dependency := range document.Dependencies {
			path := fmt.Sprintf("%s.dependencies[%d]", base, i)
			if !validText(dependency.ID, maxIDBytes) || !validText(dependency.Version, maxIDBytes) || !validText(dependency.Fingerprint, 128) {
				return invalid("INVALID_DEPENDENCY", path, "dependency ID, version, and fingerprint must be bounded nonempty text")
			}
		}
		if len(document.Axioms) != 0 {
			return unsupported("UNSUPPORTED_AXIOM", base+".axioms", "positive-facts-v1 does not execute axioms")
		}
		if err := validateSources(document.Provenance, base+".provenance"); err != nil {
			return err
		}
		if len(document.Provenance) == 0 {
			return invalid("EMPTY_PROVENANCE", base+".provenance", "package provenance is required")
		}
		for i, sortDecl := range document.Sorts {
			path := fmt.Sprintf("%s.sorts[%d]", base, i)
			if !validText(sortDecl.ID, maxIDBytes) {
				return invalid("INVALID_SORT_ID", path+".id", "sort ID is invalid")
			}
			if !validOptionalText(sortDecl.Label, maxLabelBytes) {
				return invalid("INVALID_LABEL", path+".label", "sort label is invalid")
			}
			if prior, ok := seenRecord[sortDecl.ID]; ok {
				return invalid("DUPLICATE_RECORD_ID", path+".id", "sort or symbol ID collides with "+prior)
			}
			seenRecord[sortDecl.ID] = "sort"
		}
		for i, symbol := range document.Symbols {
			path := fmt.Sprintf("%s.symbols[%d]", base, i)
			if !validText(symbol.ID, maxIDBytes) {
				return invalid("INVALID_SYMBOL_ID", path+".id", "symbol ID is invalid")
			}
			if prior, ok := seenRecord[symbol.ID]; ok {
				return invalid("DUPLICATE_RECORD_ID", path+".id", "sort or symbol ID collides with "+prior)
			}
			seenRecord[symbol.ID] = "symbol"
			if !validOptionalText(symbol.Label, maxLabelBytes) {
				return invalid("INVALID_LABEL", path+".label", "symbol label is invalid")
			}
			if !validOptionalText(symbol.Description, maxDescriptionBytes) {
				return invalid("INVALID_DESCRIPTION", path+".description", "symbol description is invalid")
			}
			if symbol.ReviewStatus != ReviewPending && symbol.ReviewStatus != ReviewApproved {
				return invalid("INVALID_REVIEW_STATUS", path+".review_status", "review_status must be pending or approved")
			}
			if len(symbol.Provenance) == 0 {
				return invalid("EMPTY_PROVENANCE", path+".provenance", "symbol provenance is required")
			}
			if len(symbol.Provenance) > 512 {
				return resource("RESOURCE_LIMIT", path+".provenance", "provenance count exceeds 512")
			}
			if err := validateSources(symbol.Provenance, path+".provenance"); err != nil {
				return err
			}
			switch symbol.Kind {
			case SymbolPrimitive:
				if symbol.Alias != nil {
					return invalid("IRRELEVANT_FIELD", path+".alias", "primitive cannot have an alias")
				}
				if symbol.Ports == nil {
					return invalid("MISSING_FIELD", path+".ports", "primitive requires ports")
				}
			case SymbolAlias:
				if symbol.Alias == nil || symbol.Ports == nil {
					return invalid("MISSING_FIELD", path, "alias requires ports and alias expression")
				}
			case SymbolOpaqueUnformalized:
				if len(symbol.Ports) != 0 || len(symbol.Constraints) != 0 || symbol.Alias != nil {
					return invalid("IRRELEVANT_FIELD", path, "opaque symbol cannot carry executable structure")
				}
			default:
				return unsupported("UNSUPPORTED_SYMBOL_KIND", path+".kind", "symbol kind is unsupported")
			}
			if len(symbol.Ports) > 16 {
				return resource("RESOURCE_LIMIT", path+".ports", "port count exceeds 16")
			}
			if len(symbol.Constraints) > 512 {
				return resource("RESOURCE_LIMIT", path+".constraints", "constraint count exceeds 512")
			}
			seenRoles := make(map[string]struct{}, len(symbol.Ports))
			for j, port := range symbol.Ports {
				portPath := fmt.Sprintf("%s.ports[%d]", path, j)
				if !validText(port.Role, maxIDBytes) {
					return invalid("INVALID_ROLE_ID", portPath+".role", "role ID is invalid")
				}
				if _, duplicate := seenRoles[port.Role]; duplicate {
					return invalid("DUPLICATE_ROLE_ID", portPath+".role", "role is duplicated")
				}
				seenRoles[port.Role] = struct{}{}
				if port.Type.Kind == algebra.TypeArrow {
					return invalid("INVALID_RELATION_PORT_TYPE", portPath+".type", "relation port must have a base sort type")
				}
			}
			if err := validateConstraints(symbol, path); err != nil {
				return err
			}
		}
		if len(document.TypicalEdges) > 512 {
			return resource("RESOURCE_LIMIT", base+".typical_edges", "typical edge count exceeds 512")
		}
		for i, edge := range document.TypicalEdges {
			path := fmt.Sprintf("%s.typical_edges[%d]", base, i)
			if !validText(edge.From, maxIDBytes) || !validText(edge.To, maxIDBytes) || !validText(edge.Kind, 64) {
				return invalid("INVALID_TYPICAL_EDGE", path, "edge endpoints and kind must be bounded nonempty text")
			}
		}
	}
	return nil
}

func validateDeclarationTypes(context []Document) error {
	sorts := make([]algebra.SortDeclaration, 0)
	signatures := make([]algebra.Signature, 0)
	for _, document := range context {
		sorts = append(sorts, document.Sorts...)
		for _, symbol := range document.Symbols {
			if symbol.Kind == SymbolPrimitive || symbol.Kind == SymbolAlias {
				signatures = append(signatures, algebra.Signature{ID: symbol.ID, Label: symbol.Label, Ports: append([]algebra.Port(nil), symbol.Ports...)})
			}
		}
	}
	return algebra.CheckDeclarations(algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: sorts, Signatures: signatures})
}

func validateConstraints(symbol Symbol, path string) error {
	for i, constraint := range symbol.Constraints {
		constraintPath := fmt.Sprintf("%s.constraints[%d]", path, i)
		switch constraint.Kind {
		case ConstraintArity:
			if constraint.Count == nil {
				return invalid("MISSING_FIELD", constraintPath+".count", "arity constraint requires count")
			}
			if *constraint.Count < 0 || *constraint.Count > 16 {
				return invalid("INVALID_CONSTRAINT_VALUE", constraintPath+".count", "arity count must be between 0 and 16")
			}
			if *constraint.Count != len(symbol.Ports) {
				return invalid("ARITY_CONSTRAINT_MISMATCH", constraintPath+".count", "arity constraint must equal the declared port count")
			}
			if constraint.Role != "" || constraint.Type != nil {
				return invalid("IRRELEVANT_FIELD", constraintPath, "arity constraint cannot carry role or type")
			}
		case ConstraintPortType:
			if constraint.Role == "" || constraint.Type == nil {
				return invalid("MISSING_FIELD", constraintPath, "port_type constraint requires role and type")
			}
			if constraint.Count != nil {
				return invalid("IRRELEVANT_FIELD", constraintPath+".count", "port_type constraint cannot carry count")
			}
			found := false
			for _, port := range symbol.Ports {
				if port.Role == constraint.Role {
					found = true
					if !algebra.TypeEqual(port.Type, *constraint.Type) {
						return invalid("PORT_TYPE_CONSTRAINT_MISMATCH", constraintPath+".type", "constraint type differs from declared port type")
					}
					break
				}
			}
			if !found {
				return invalid("UNKNOWN_CONSTRAINT_ROLE", constraintPath+".role", "constraint role is not a declared port")
			}
		default:
			return unsupported("UNSUPPORTED_CONSTRAINT", constraintPath+".kind", "constraint kind is unsupported")
		}
	}
	return nil
}

func validateSources(sources []provenance.SourceRef, path string) error {
	seen := make(map[string]struct{}, len(sources))
	for i, source := range sources {
		if _, exists := seen[source.ID]; exists {
			return invalid("DUPLICATE_SOURCE_ID", fmt.Sprintf("%s[%d].id", path, i), "provenance source ID is duplicated")
		}
		seen[source.ID] = struct{}{}
		if err := source.Validate(); err != nil {
			return invalid("INVALID_PROVENANCE", fmt.Sprintf("%s[%d]", path, i), err.Error())
		}
	}
	return nil
}

func validatePackageGraph(documents []Document, byID map[string]Document) error {
	if err := validateDocuments(documents); err != nil {
		return err
	}
	state := make(map[string]uint8, len(documents))
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return invalid("DEPENDENCY_CYCLE", "packages", "package dependency graph contains a cycle")
		case 2:
			return nil
		}
		state[id] = 1
		document := byID[id]
		seen := make(map[string]struct{}, len(document.Dependencies))
		for i, dependency := range document.Dependencies {
			path := fmt.Sprintf("package[%s].dependencies[%d]", id, i)
			if _, duplicate := seen[dependency.ID]; duplicate {
				return invalid("DUPLICATE_DEPENDENCY", path+".id", "dependency ID is duplicated")
			}
			seen[dependency.ID] = struct{}{}
			depDoc, exists := byID[dependency.ID]
			if !exists {
				return invalid("MISSING_DEPENDENCY", path+".id", "dependency package was not supplied")
			}
			if depDoc.Version != dependency.Version {
				return invalid("DEPENDENCY_VERSION_MISMATCH", path+".version", "dependency version does not match supplied package")
			}
			fingerprint, err := documentResourceFingerprint(depDoc)
			if err != nil {
				return invalid("DEPENDENCY_FINGERPRINT_UNAVAILABLE", path+".fingerprint", err.Error())
			}
			if fingerprint != dependency.Fingerprint {
				return invalid("DEPENDENCY_FINGERPRINT_MISMATCH", path+".fingerprint", "dependency resource fingerprint does not match supplied package")
			}
			if err := visit(dependency.ID); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func documentResourceFingerprint(document Document) (string, error) {
	copyDocument, err := cloneDocument(document)
	if err != nil {
		return "", err
	}
	canonical, err := canonicalResource(copyDocument)
	if err != nil {
		return "", err
	}
	return digest("metamodel-resource-v1", canonical), nil
}

func dependencyClosureDocuments(document Document, byID map[string]Document) []Document {
	seen := map[string]bool{document.ID: true}
	queue := make([]string, 0, len(document.Dependencies))
	for _, dep := range document.Dependencies {
		queue = append(queue, dep.ID)
	}
	out := make([]Document, 0)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		dep, exists := byID[id]
		if !exists {
			continue
		}
		out = append(out, dep)
		for _, nested := range dep.Dependencies {
			queue = append(queue, nested.ID)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func topologicalDocuments(documents []Document, byID map[string]Document) []Document {
	ids := make([]string, 0, len(documents))
	for _, document := range documents {
		ids = append(ids, document.ID)
	}
	sort.Strings(ids)
	seen := make(map[string]bool, len(ids))
	out := make([]Document, 0, len(ids))
	var visit func(string)
	visit = func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true
		doc := byID[id]
		deps := append([]Dependency(nil), doc.Dependencies...)
		sort.Slice(deps, func(i, j int) bool { return deps[i].ID < deps[j].ID })
		for _, dep := range deps {
			visit(dep.ID)
		}
		out = append(out, doc)
	}
	for _, id := range ids {
		visit(id)
	}
	return out
}

func snapshotClosure(document Document, snapshots []*Registry) ([]*Registry, []*Registry, error) {
	if len(snapshots)+1 > maxPackages {
		return nil, nil, resource("RESOURCE_LIMIT", "dependency_snapshots", "package closure exceeds limit 16")
	}
	byID := make(map[string]*Registry, len(snapshots))
	for i, snapshot := range snapshots {
		if snapshot == nil {
			return nil, nil, invalid("NIL_REGISTRY", fmt.Sprintf("dependency_snapshots[%d]", i), "registry is nil")
		}
		if _, exists := byID[snapshot.ID()]; exists {
			return nil, nil, invalid("DUPLICATE_DEPENDENCY", "dependency_snapshots", "dependency package ID is duplicated")
		}
		byID[snapshot.ID()] = snapshot
	}
	needed := make(map[string]bool)
	direct := make([]*Registry, 0, len(document.Dependencies))
	queue := make([]string, 0, len(document.Dependencies))
	directSeen := make(map[string]struct{}, len(document.Dependencies))
	for _, dep := range document.Dependencies {
		if _, exists := directSeen[dep.ID]; exists {
			return nil, nil, invalid("DUPLICATE_DEPENDENCY", "dependencies", "dependency ID is duplicated")
		}
		directSeen[dep.ID] = struct{}{}
		queue = append(queue, dep.ID)
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if needed[id] {
			continue
		}
		needed[id] = true
		registry, exists := byID[id]
		if !exists {
			return nil, nil, invalid("MISSING_DEPENDENCY", "dependencies", "required dependency snapshot was not supplied")
		}
		for _, dep := range registry.document.Dependencies {
			nested, ok := byID[dep.ID]
			if !ok || nested.Version() != dep.Version || nested.ResourceFingerprint() != dep.Fingerprint {
				return nil, nil, invalid("DEPENDENCY_PIN_MISMATCH", "dependency_snapshots", "transitive dependency snapshot does not match its pin")
			}
			queue = append(queue, dep.ID)
		}
	}
	if len(needed) != len(byID) {
		return nil, nil, invalid("UNDECLARED_DEPENDENCY_SNAPSHOT", "dependency_snapshots", "an unrelated registry snapshot was supplied")
	}
	for _, declaration := range document.Dependencies {
		registry := byID[declaration.ID]
		if registry.Version() != declaration.Version || registry.ResourceFingerprint() != declaration.Fingerprint {
			return nil, nil, invalid("DEPENDENCY_PIN_MISMATCH", "dependencies", "dependency snapshot does not match its version and fingerprint pin")
		}
		direct = append(direct, registry)
	}
	closure := make([]*Registry, 0, len(byID))
	for id := range needed {
		closure = append(closure, byID[id])
	}
	sort.Slice(closure, func(i, j int) bool { return closure[i].ID() < closure[j].ID() })
	return closure, direct, nil
}

func validatePackageReferences(document Document, context []Document) error {
	allowedSymbols := make(map[string]Symbol)
	for _, candidate := range context {
		if candidate.ID != document.ID && !isInDependencyClosure(document, candidate.ID, context) {
			continue
		}
		for _, symbol := range candidate.Symbols {
			allowedSymbols[symbol.ID] = symbol
		}
	}
	for i, edge := range document.TypicalEdges {
		if _, ok := allowedSymbols[edge.From]; !ok {
			return invalid("UNKNOWN_TYPICAL_EDGE_SYMBOL", fmt.Sprintf("typical_edges[%d].from", i), "typical edge source is not a symbol in the dependency closure")
		}
		if _, ok := allowedSymbols[edge.To]; !ok {
			return invalid("UNKNOWN_TYPICAL_EDGE_SYMBOL", fmt.Sprintf("typical_edges[%d].to", i), "typical edge target is not a symbol in the dependency closure")
		}
	}
	for _, symbol := range document.Symbols {
		if symbol.Alias == nil {
			continue
		}
		if err := walkAtoms(symbol.Alias.Expression, func(id, path string) error {
			if target, ok := allowedSymbols[id]; !ok {
				return invalid("UNKNOWN_SYMBOL", path, "alias references an undeclared symbol")
			} else if target.Kind == SymbolOpaqueUnformalized {
				return unsupported("OPAQUE_SYMBOL_USE", path, "opaque symbol has no executable signature")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func isInDependencyClosure(document Document, target string, context []Document) bool {
	byID := make(map[string]Document, len(context))
	for _, candidate := range context {
		byID[candidate.ID] = candidate
	}
	seen := make(map[string]bool)
	queue := make([]string, 0, len(document.Dependencies))
	for _, dep := range document.Dependencies {
		queue = append(queue, dep.ID)
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == target {
			return true
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		if doc, ok := byID[id]; ok {
			for _, dep := range doc.Dependencies {
				queue = append(queue, dep.ID)
			}
		}
	}
	return false
}

func validateAliasCycles(documents []Document) error {
	aliases := make(map[string]Symbol)
	for _, document := range documents {
		for _, symbol := range document.Symbols {
			if symbol.Kind == SymbolAlias {
				aliases[symbol.ID] = symbol
			}
		}
	}
	state := make(map[string]uint8, len(aliases))
	var visit func(string) error
	visit = func(id string) error {
		if state[id] == 1 {
			return invalid("ALIAS_CYCLE", "symbols["+id+"]", "alias reference graph contains a cycle")
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		symbol := aliases[id]
		if err := walkAtoms(symbol.Alias.Expression, func(target, _ string) error {
			if _, ok := aliases[target]; ok {
				return visit(target)
			}
			return nil
		}); err != nil {
			return err
		}
		state[id] = 2
		return nil
	}
	ids := make([]string, 0, len(aliases))
	for id := range aliases {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func validateAliases(documents []Document) error {
	byID := make(map[string]Document, len(documents))
	for _, document := range documents {
		byID[document.ID] = document
	}
	for _, document := range documents {
		context := append([]Document{document}, dependencyClosureDocuments(document, byID)...)
		sorts := make([]algebra.SortDeclaration, 0)
		signatures := make([]algebra.Signature, 0)
		for _, candidate := range context {
			sorts = append(sorts, candidate.Sorts...)
			for _, declared := range candidate.Symbols {
				if declared.Kind == SymbolPrimitive || declared.Kind == SymbolAlias {
					signatures = append(signatures, algebra.Signature{ID: declared.ID, Label: declared.Label, Ports: append([]algebra.Port(nil), declared.Ports...)})
				}
			}
		}
		for index, symbol := range document.Symbols {
			if symbol.Kind != SymbolAlias {
				continue
			}
			path := fmt.Sprintf("package[%s].symbols[%d].alias", document.ID, index)
			if err := validateAlias(symbol, path, sorts, signatures); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateAlias(symbol Symbol, path string, sorts []algebra.SortDeclaration, signatures []algebra.Signature) error {
	alias := symbol.Alias
	if alias == nil || alias.Expression == nil {
		return invalid("MISSING_ALIAS_EXPRESSION", path+".expression", "alias expression is required")
	}
	if len(alias.Interface) != len(symbol.Ports) {
		return invalid("ALIAS_INTERFACE_ARITY", path+".interface", "alias interface must map every port exactly once")
	}
	seenPorts, seenBinders := make(map[string]bool), make(map[string]bool)
	term := alias.Expression
	for i, port := range symbol.Ports {
		binding := alias.Interface[i]
		if binding.Port != port.Role {
			return invalid("ALIAS_INTERFACE_ORDER", fmt.Sprintf("%s.interface[%d].port", path, i), "alias interface must follow declared port order")
		}
		if seenPorts[binding.Port] || seenBinders[binding.Binder] {
			return invalid("ALIAS_INTERFACE_DUPLICATE", fmt.Sprintf("%s.interface[%d]", path, i), "port and binder mappings must be one-to-one")
		}
		seenPorts[binding.Port], seenBinders[binding.Binder] = true, true
		if term == nil || term.Kind != algebra.TermLambda {
			return invalid("ALIAS_LAMBDA_INTERFACE", path+".expression", "alias must begin with one lambda per declared port")
		}
		if term.ID != binding.Binder {
			return invalid("ALIAS_LAMBDA_BINDER", fmt.Sprintf("%s.interface[%d].binder", path, i), "interface binder does not match the corresponding leading lambda")
		}
		if term.Type == nil || !algebra.TypeEqual(*term.Type, port.Type) {
			return invalid("ALIAS_LAMBDA_TYPE", fmt.Sprintf("%s.expression.type", path), "leading lambda type differs from the corresponding port type")
		}
		term = term.Body
	}
	if term == nil {
		return invalid("MISSING_ALIAS_BODY", path+".expression", "alias expression has no body")
	}
	resultType, err := checkAliasExpression(alias.Expression, sorts, signatures, path)
	if err != nil {
		return err
	}
	expected := algebra.Type{Kind: algebra.TypeProp}
	for i := len(symbol.Ports) - 1; i >= 0; i-- {
		previous := expected
		argument := symbol.Ports[i].Type
		expected = algebra.Type{Kind: algebra.TypeArrow, From: &argument, To: &previous}
	}
	if !algebra.TypeEqual(resultType, expected) {
		return invalid("ALIAS_RESULT_TYPE", path+".expression", "alias must denote a curried relation from its ordered ports to Prop")
	}
	return nil
}

func checkAliasExpression(expression *algebra.Term, sorts []algebra.SortDeclaration, signatures []algebra.Signature, path string) (algebra.Type, error) {
	result, err := algebra.Check(algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: sorts, Signatures: signatures, Closed: true, Expression: expression})
	if err != nil {
		return algebra.Type{}, err
	}
	if result.Status != status.OutcomeOK || len(result.Holes) != 0 || len(result.FreeParameters) != 0 {
		return algebra.Type{}, invalid("INCOMPLETE_ALIAS", path, "alias must be closed and contain no holes")
	}
	return result.Type, nil
}

func checkSelectedAtoms(term *algebra.Term, symbols map[string]Symbol) error {
	return walkAtoms(term, func(id, path string) error {
		symbol, ok := symbols[id]
		if !ok {
			// Leave undeclared IDs to algebra.Check so its UNKNOWN_SYMBOL
			// diagnostic remains authoritative.
			return nil
		}
		if symbol.Kind != SymbolPrimitive || symbol.ReviewStatus != ReviewApproved {
			return unsupported("SYMBOL_NOT_OPERATIONAL", path, "only approved primitive symbols are available for expression checking")
		}
		return nil
	})
}

func walkAtoms(term *algebra.Term, visit func(id, path string) error) error {
	active := make(map[*algebra.Term]bool)
	nodes := 0
	var walk func(*algebra.Term, string, int) error
	walk = func(value *algebra.Term, path string, depth int) error {
		if value == nil {
			return nil
		}
		if depth > algebra.DefaultMaxDepth {
			return resource("RESOURCE_LIMIT", path, "term depth exceeds algebra checker bound")
		}
		if active[value] {
			return invalid("CYCLIC_TERM", path, "term graph contains a pointer cycle")
		}
		nodes++
		if nodes > algebra.DefaultMaxNodes {
			return resource("RESOURCE_LIMIT", path, "term node count exceeds algebra checker bound")
		}
		active[value] = true
		defer delete(active, value)
		if value.Kind == algebra.TermAtom {
			if err := visit(value.Symbol, path+".symbol"); err != nil {
				return err
			}
		}
		for i, arg := range value.Args {
			if err := walk(arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1); err != nil {
				return err
			}
		}
		for i, item := range value.Operands {
			if err := walk(item, fmt.Sprintf("%s.operands[%d]", path, i), depth+1); err != nil {
				return err
			}
		}
		if err := walk(value.Body, path+".body", depth+1); err != nil {
			return err
		}
		if err := walk(value.Function, path+".function", depth+1); err != nil {
			return err
		}
		return walk(value.Argument, path+".argument", depth+1)
	}
	return walk(term, "expression", 1)
}

func validText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value)
}
func validOptionalText(value string, maximum int) bool {
	return len(value) <= maximum && utf8.ValidString(value)
}
func invalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}
func unsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}
func resource(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: code, Path: path, Message: message}
}

// Package bundled turns the authored Go catalogs into a validated data snapshot.
// It preserves every lexical binding status and its original provenance.
package bundled

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

const DatasetID = "golingua-authored"

// Candidate builds the exact CORE and PL/EN lexical snapshot used by the
// runtime store. It never changes approval status or adds language rules.
func Candidate() (staging.Candidate, error) {
	entries, err := catalog.Entries()
	if err != nil {
		return staging.Candidate{}, err
	}
	concepts := make([]core.ConceptDefinition, len(entries))
	for index, entry := range entries {
		concepts[index] = entry.Definition
	}
	graph, err := catalog.Compile()
	if err != nil {
		return staging.Candidate{}, err
	}
	compiledRelations := graph.RelationTypes()
	relations := make([]core.RelationDefinition, len(compiledRelations))
	for index, relation := range compiledRelations {
		relations[index] = core.RelationDefinition{RelationType: relation}
	}
	polish, err := pl.LexicalDocument()
	if err != nil {
		return staging.Candidate{}, err
	}
	english, err := en.LexicalDocument()
	if err != nil {
		return staging.Candidate{}, err
	}
	lexical := lexicon.Document{SchemaVersion: lexicon.SchemaVersion}
	lexical.Lexemes = append(append(lexical.Lexemes, polish.Lexemes...), english.Lexemes...)
	lexical.Senses = append(append(lexical.Senses, polish.Senses...), english.Senses...)
	lexical.Bindings = append(append(lexical.Bindings, polish.Bindings...), english.Bindings...)
	lexiconBytes, err := lexicon.Encode(lexical, graph, lexicon.DefaultBudget())
	if err != nil {
		return staging.Candidate{}, fmt.Errorf("encode authored lexicon: %w", err)
	}
	conceptBytes, err := json.Marshal(core.ConceptDocument{SchemaVersion: 1, Relations: relations, Concepts: concepts})
	if err != nil {
		return staging.Candidate{}, fmt.Errorf("encode authored CORE: %w", err)
	}
	files := []staging.File{
		{ID: "concepts", Path: "authored.concepts.json", SourceID: "golingua-core-catalog", Contents: conceptBytes},
		{ID: "lexicon", Path: "authored.lexicon.json", SourceID: "golingua-lexical-catalogs", Contents: lexiconBytes},
	}
	manifest := importer.Manifest{
		SchemaVersion: importer.ManifestSchemaVersion, DatasetID: DatasetID, Version: "2026-10-23",
		Sources: []provenance.SourceRef{
			{ID: "golingua-core-catalog", Path: "internal/catalog", Version: "2026-10-23", License: "unknown", Author: "golingua contributors"},
			{ID: "golingua-lexical-catalogs", Path: "languages", Version: "2026-10-23", License: "unknown", Author: "golingua contributors"},
		},
	}
	for _, file := range files {
		digest := sha256.Sum256(file.Contents)
		manifest.Files = append(manifest.Files, importer.ManifestFile{
			ID: file.ID, Path: file.Path, SourceID: file.SourceID,
			ChecksumSHA256: hex.EncodeToString(digest[:]), SizeBytes: int64(len(file.Contents)),
		})
	}
	digest, err := staging.ComputeDigest(manifest, files)
	if err != nil {
		return staging.Candidate{}, err
	}
	candidate := staging.Candidate{Manifest: manifest, Files: files, Digest: digest}
	if _, err := pack.Compile(candidate); err != nil {
		return staging.Candidate{}, fmt.Errorf("validate authored pack: %w", err)
	}
	return candidate, nil
}

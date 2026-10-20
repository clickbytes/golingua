# 09 — LanguagePlugin jako niezależny, deklaratywny pakiet PL/EN

**Status:** PROPOSED; implementacja [G03-06 #36](https://github.com/clickbytes/golingua/issues/36), [G04-06 #43](https://github.com/clickbytes/golingua/issues/43), [G05 #6](https://github.com/clickbytes/golingua/issues/6).

## Rozdzielenie uprawnień

`Core/Meaning/Reasoner` nie importują `languages/pl` i `languages/en`. Każdy język jest pakietem, który **deklaruje** i udowadnia:
- `language code, package version, source provenance`;
- własne profile cech form/syntactic/surface, orthography/tokenization policies;
- inventory `MorphLexeme/Sense/SemanticBinding`, `Paradigm/Exception`, `ValencyFrame`;
- `ConstructionFamily/Variant/Operator` i `ConstructionBinding`;
- rozłączne, wersjonowane `capabilities` dla operacji `morphAnalyze`, `morphGenerate`, `sentenceAnalyze`, `sentenceRealize`, `dialogueActRealize` itd., oraz `evidence case IDs`.

`LanguagePlugin` to dane i compiled metadata, **nie** arbitralne `.so` ładowane przez `plugin.Open`, skrypt importowany z JSON lub mutowalny global singleton. Go registry/loader jest wspólny, a potencjalny język TR/IT/FI/AR to rozszerzenie profili i plans, nie osobne rozgałęzienie w CORE.

## Format manifestu — proponowana semantyka

```json
{
  "schemaVersion":"proposed-0",
  "language":"pl",
  "packageId":"golingua.pl",
  "sourceDigest":"pending-review",
  "capabilities":[
    {
      "operation":"sentenceAnalyze",
      "constructionId":"pl.transitive.active",
      "status":"proposed",
      "evidence":[]
    }
  ]
}
```

Dopóki status nie jest `implemented` poparty istniejącym, zatwierdzonym PASS fixture na danym pack SHA, CLI nie może reklamować capability. `proposed`/`blocked`/`notRun` to **nie** synonimy „gotowe”. Nie wpisuj do finalnego manifestu `sourceDigest=pending-review` jako poprawnego hash, to tylko przykład składni.

Wiadomości, pytania i pytania zagnieżdżone wymagają osobnych wrapper/proposition constraints: `question.yes-no` w manifest ≠ gwarancja, że każdy event można pytać i realizować.

## Kontrakt runtime Go — konceptualne funkcje

```go
// Pseudotypy i nazwy do doprecyzowania w G03-06; NIE są obecnie implementowane.
type PackageReader interface {
    Language() LanguageID
    Capabilities() []Capability
    LookupLexemes(lemma string) []LexemeID
    AnalyzeMorph(surface string, budget Budget) ([]MorphReading, error)
    GenerateMorph(id LexemeID, features FeatureBundle, budget Budget) ([]Form, error)
    ConstructionPlans() []ConstructionPlanID
}
```

`Engine` jest separatorem ścieżek:
- source plugin odczytuje tekst i proponuje MeaningGraph;
- neutralny CORE waliduje meaning; reasoner opcjonalnie planuje nowy meaning;
- target plugin realizuje wyłącznie w zatwierdzonym zakresie.

Brak operacji → jawny `unsupported`. Nie dodawać ukrytego „best guess EN string” jako obejścia dla brakującego polskiego realizer.

## Language-local feature contracts

Wspólna techniczna reprezentacja `FeatureBundle` należy do neutralnego pakietu `internal/features`; **zakres dozwolonych nazw i wartości należy do pluginu**. Jeden język może mieć osiem przypadków, inny zero jawnych przypadków w rzeczownikach, trzeci controller niebędący subject. Nie rozszerzać CORE o `gender=masculine` tylko dlatego, że język wymaga takiej cechy przy uzgodnieniu.

Wprowadź operator delegacji `govern(slot,casePolicy)`, `agree(target,controller,keys)`, `selectAllomorph(conditions)` i `linearize(variant)` w planach zwalidowanych przed runtime. Polityka czasów i liczebników również jest deklaracją języka, nie if(language=="pl") w Engine.

## Granice i walidacja loadera

1. Syntactic schema i digest files, wersje i max sizes.
2. Referencje do rzeczywistych SenseID/CoreRef, ValencyFrame, Family/Variant/Procedure/Paradigm; brakujące → INVALID.
3. Mapowania semantyczne wyłącznie approved przy translatable target; unrelated/candidate nie czynią ekwiwalentu.
4. Plugin claims są filtrowane przez rzeczywiste compiled plans oraz EvidenceRecords.
5. Input JSON jest danymi; nigdy kodem do wykonania, nie autoload z sieci.
6. Nie importuj wszystkich języków do jednej globalnej listy cech; pakiet PL nie może mutować feature inventory EN.
7. Immutable compiled pack; source reload wyłącznie przez nową walidowaną kompilację.

## Testy kontraktowe

- Manifest z zadeklarowanym `sentenceRealize` bez construction plan lub reviewed evidence → FAIL.
- Dane PL przekazane do EN profile → FAIL.
- Dwa różne SenseID identycznej powierzchni → oba wyniki analiz dostępne, nie „first wins”.
- Nieobsługiwana kombinacja formy/wariantu czy pytania → typed unsupported, nie silent fallback.
- Źródłowe i docelowe role syntaktyczne różnią się (np. PL dat, EN PP), ale MeaningGraph ma tę samą rolę semantyczną → PASS, jeśli independently approved.

# Quantity, time, and multiword expressions

The Go APIs in `internal/construction` keep three representations separate:

- `QuantityPlan` requires a validated `Meaning` quantifier node and an NP
  plan. Its semantic quantifier is independent of the optional grammatical
  feature bundle, which may describe number, case, or gender only when the
  language profile declares those dimensions. The caller supplies the explicit
  quantifier target and NP plan; absent an approved sense-to-concept binding,
  the compiler does not assert that they denote the same entity.
- `TemporalPlan` represents an explicit calendar date, clock time, duration,
  or relative day. It does not store or infer grammatical tense. Clock timezone
  offsets remain optional; relative days require an explicit Meaning anchor.
- `MWEPlan` references one `LexicalSense` and ordered components. Fixed and
  inflected components reference lexemes; inflected components also require
  explicit feature values; variable components reference a compiled NP plan.
  Literal composition is a separate construction and is never a fallback for
  a missing lexicalized sense.

`languages/{pl,en}/constructions/quantity-time-mwe.json` lists the bounded
structural kinds and language feature dimensions currently accepted by the
compilers. These are schema capabilities, not approved quantity readings,
time expressions, idioms, or surface forms. No PL/EN lexicalized MWE entries,
valency mappings, or phrase realization providers are installed. Therefore
the APIs validate caller supplied data but the application does not generate
quantity, time, or MWE text from these profiles.

# #344 observational report

The existing `scripts/report-language-matrix.py` remains a presentation tool over declared cases and `go test -v` output. It keeps check outcomes, runtime adapter executions, authored I inputs, observed graphs/traces, catalogue metadata, gaps, and control tests distinct. Coverage target values describe sample planning only and do not gate acceptance.

JSON is decoded with `json.JSONDecoder.raw_decode` immediately after known markers (`matrix_inventory=`, `matrix_snapshot=`, `matrix_sense_association=`, `observation=`, and existing direct `result=`). Source case IDs remain the fallback inventory, so missing or truncated runtime metadata leaves declared records visible as NOT RUN.

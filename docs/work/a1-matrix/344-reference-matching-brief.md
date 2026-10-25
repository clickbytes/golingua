# M8 reference matching prerequisite

This bounded change generalizes request-context matching for authored Meaning fixtures with one or more anchored mentions. The historical `IdentificationMeaning` fixture type remains the context carrier; existing one-referent `IdentificationMatches` behavior stays strict.

`ContextMatches` checks each candidate request context against the authored graph and context. Anchors constrain the same alpha-node bijection that matches the graph, so a speaker/addressee anchor cannot move to a different semantic role even when both anchor values remain present. Mention order, node names, graph node order, and root order are not identity. Resolved referents, evidence IDs, missing/extra mentions, and wrong anchors are rejected.

`MatchAllowedCandidate` resolves graph and context together for each allowed MeaningID. Analysis uses each matched variant's own context. Integration preflight records validated source candidate IDs and checks every target alternative against the exact source MeaningID that produced it. Context-free matching retains its existing path. No language cases, product grammar, or engine behavior are added by this prerequisite.

# Semantic binding pilot

This small, explicit pilot exercises four existing PL/EN perception senses and two separate EN technical naming senses. It is executable metadata for the bounded typed-target API; it does not add parsing, translation, sentence routes, or morphology generation.

The selected primitives are new, agent-reviewed formal declarations for this pilot only: `pilot.vision_event(Event, Entity, Entity)`, `pilot.auditory_event(Event, Entity, Entity)`, and a neutral ordered relation `pilot.r(Entity, Entity)`. The visual and auditory predicates have the same type shape but distinct symbols. Entity-valued stimulus references are admitted by these target mappings; an Event-valued stimulus is outside this pilot profile. The frame exposes only `experiencer` and `stimulus`; the predicate's `event` port remains an explicit input reference supplied through `port_references`, never an invented frame role.

Each perception request supplies event `E1` and participants `A` and `B`, maps the two authored frame roles explicitly, and supplies the event port explicitly. The four resulting expressions therefore share the event/participant structure; visual and auditory expressions remain distinct by symbol, and swapping role assignments changes the expression.

The two English `link` senses are deliberately technical naming conventions, not broad claims about the natural word “link.” `direct` names one `pilot.r` edge; `indirect` is a separate goal constrained to a two-edge path with an existential middle entity. No axiom licenses collapsing that path to one edge. The samples require explicit request references and do not assert any facts.

`source-records.json` preserves the four original reviewed CORE bindings and their lexical/frame source records separately from the new target bindings. The pilot uses new target provenance and does not alter or promote the compiled CORE catalog. The formal declarations and target mappings are agent-reviewed for this bounded executable example; owner linguistic acceptance is **NOT EVALUATED**. Existing lexical review records are retained as provenance, not reused as approval for the new formal mappings.

All examples are declarative input files. Package loading and typed instantiation do not imply natural-language parsing, translation quality, world truth, or general coverage of event-valued perception stimuli.

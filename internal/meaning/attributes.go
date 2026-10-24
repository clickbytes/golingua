package meaning

// TemporalFrame is the language-neutral temporal reference frame of an
// event or state. The zero value means that the frame is unresolved.
type TemporalFrame string

const (
	TemporalFramePast    TemporalFrame = "past"
	TemporalFramePresent TemporalFrame = "present"
)

// Valid reports whether frame is one of the explicitly supported frames.
func (frame TemporalFrame) Valid() bool {
	return frame == TemporalFramePast || frame == TemporalFramePresent
}

// Multiplicity describes whether a referent denotes one or multiple
// individuals. The zero value leaves multiplicity unresolved.
type Multiplicity string

const (
	MultiplicityOne  Multiplicity = "one"
	MultiplicityMany Multiplicity = "many"
)

// Valid reports whether multiplicity is explicitly supported.
func (multiplicity Multiplicity) Valid() bool {
	return multiplicity == MultiplicityOne || multiplicity == MultiplicityMany
}

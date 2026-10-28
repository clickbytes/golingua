package valency

import (
	"sort"

	"github.com/clickbytes/golingua/internal/ids"
)

// Index is an immutable, deterministic view of validated sense frames,
// approved construction bindings, and separately stored pending candidates.
type Index struct {
	frames     []ValencyFrame
	bindings   []ConstructionBinding
	candidates []ConstructionCandidate
	bySenseID  map[string][]int
	byFamilyID map[string][]int
}

// BuildIndex validates frames and bindings before compiling the two lookup
// paths. Pending/unapproved bindings are rejected and never become licensed
// runtime data.
func BuildIndex(frames []ValencyFrame, bindings []ConstructionBinding) (*Index, error) {
	return BuildIndexWithCandidates(frames, bindings, nil)
}

// BuildIndexWithCandidates compiles approved runtime bindings and pending
// construction candidates into separate immutable collections. Candidates
// are validated for references but never enter Bindings() or runtime paths.
func BuildIndexWithCandidates(frames []ValencyFrame, bindings []ConstructionBinding, candidates []ConstructionCandidate) (*Index, error) {
	if len(frames) == 0 {
		return nil, invalid("EMPTY_VALENCY_INDEX", "frames", "valency index requires at least one frame")
	}
	if len(frames) > MaxFrames {
		return nil, resourceLimit("frames", len(frames), MaxFrames)
	}
	if len(bindings) > MaxBindings {
		return nil, resourceLimit("bindings", len(bindings), MaxBindings)
	}
	if len(candidates) > MaxBindings {
		return nil, resourceLimit("construction_candidates", len(candidates), MaxBindings)
	}
	canonicalFrames := make([]ValencyFrame, len(frames))
	framePositions := make(map[string]int, len(frames))
	bySense := make(map[string][]int)
	indexItems := 0
	for index, frame := range frames {
		if err := frame.Validate(); err != nil {
			return nil, err
		}
		indexItems += len(frame.SemanticRoles) + len(frame.Arguments) + len(frame.SelectionalConstraints) + len(frame.AllowedFamilyIDs)
		for _, constraint := range frame.SelectionalConstraints {
			indexItems += len(constraint.AllowedConceptIDs) + len(constraint.ExcludedConceptIDs) + len(constraint.AllowedKinds)
		}
		if indexItems > MaxIndexItems {
			return nil, resourceLimit("index_items", indexItems, MaxIndexItems)
		}
		key := string(frame.ID)
		if _, exists := framePositions[key]; exists {
			return nil, invalid("DUPLICATE_VALENCY_FRAME_ID", "frames", "frame ID is duplicated")
		}
		canonicalFrames[index] = canonicalFrame(frame)
		framePositions[key] = index
		bySense[frame.SenseID.String()] = append(bySense[frame.SenseID.String()], index)
	}
	canonicalBindings := make([]ConstructionBinding, len(bindings))
	byFamily := make(map[string][]int)
	seenBindings := make(map[string]struct{}, len(bindings))
	for index, binding := range bindings {
		if err := validateBindingStructure(binding); err != nil {
			return nil, err
		}
		indexItems += 4
		if indexItems > MaxIndexItems {
			return nil, resourceLimit("index_items", indexItems, MaxIndexItems)
		}
		framePosition, exists := framePositions[string(binding.ValencyFrameID)]
		if !exists {
			return nil, unknown("UNKNOWN_VALENCY_FRAME", "bindings["+itoa(index)+"].valency_frame_id", "construction binding refers to no indexed frame")
		}
		frame := canonicalFrames[framePosition]
		if binding.SenseID != frame.SenseID {
			return nil, invalid("SENSE_FRAME_IDENTITY_MISMATCH", "bindings["+itoa(index)+"].sense_id", "binding sense must match the referenced frame")
		}
		if !containsConstruction(frame.AllowedFamilyIDs, binding.FamilyID) {
			return nil, unsupported("FAMILY_NOT_LICENSED_BY_FRAME", "bindings["+itoa(index)+"].family_id", "frame does not license the bound construction family")
		}
		key := bindingKey(binding)
		if _, exists := seenBindings[key]; exists {
			return nil, invalid("DUPLICATE_CONSTRUCTION_BINDING", "bindings", "construction binding is duplicated")
		}
		seenBindings[key] = struct{}{}
		canonicalBindings[index] = cloneBinding(binding)
		byFamily[binding.FamilyID.String()] = append(byFamily[binding.FamilyID.String()], index)
	}
	canonicalCandidates := make([]ConstructionCandidate, len(candidates))
	seenCandidates := make(map[string]struct{}, len(candidates))
	for index, candidate := range candidates {
		if err := validateCandidateStructure(candidate); err != nil {
			return nil, err
		}
		indexItems += 4
		if indexItems > MaxIndexItems {
			return nil, resourceLimit("index_items", indexItems, MaxIndexItems)
		}
		framePosition, exists := framePositions[string(candidate.ValencyFrameID)]
		if !exists {
			return nil, unknown("UNKNOWN_CANDIDATE_VALENCY_FRAME", "candidates["+itoa(index)+"].valency_frame_id", "construction candidate refers to no indexed frame")
		}
		frame := canonicalFrames[framePosition]
		if candidate.SenseID != frame.SenseID {
			return nil, invalid("CANDIDATE_SENSE_FRAME_IDENTITY_MISMATCH", "candidates["+itoa(index)+"].sense_id", "candidate sense must match the referenced frame")
		}
		if !containsConstruction(frame.AllowedFamilyIDs, candidate.FamilyID) {
			return nil, unsupported("CANDIDATE_FAMILY_NOT_LICENSED_BY_FRAME", "candidates["+itoa(index)+"].family_id", "frame does not list the proposed construction family")
		}
		bindingLike := ConstructionBinding{SenseID: candidate.SenseID, ValencyFrameID: candidate.ValencyFrameID, FamilyID: candidate.FamilyID, VariantID: candidate.VariantID}
		key := bindingKey(bindingLike)
		if _, exists := seenCandidates[key]; exists {
			return nil, invalid("DUPLICATE_CONSTRUCTION_CANDIDATE", "candidates", "construction candidate is duplicated")
		}
		if _, exists := seenBindings[key]; exists {
			return nil, invalid("CANDIDATE_ALREADY_APPROVED", "candidates", "an approved binding already exists for this route")
		}
		seenCandidates[key] = struct{}{}
		canonicalCandidates[index] = cloneCandidate(candidate)
	}
	orderFrames(canonicalFrames)
	orderBindings(canonicalBindings)
	orderCandidates(canonicalCandidates)
	// Rebuild positions after sorting so all index values refer to canonical
	// backing slices and accessor order is stable regardless of input order.
	framePositions = make(map[string]int, len(canonicalFrames))
	bySense = make(map[string][]int)
	for index, frame := range canonicalFrames {
		framePositions[string(frame.ID)] = index
		bySense[frame.SenseID.String()] = append(bySense[frame.SenseID.String()], index)
	}
	for senseID, positions := range bySense {
		sort.Slice(positions, func(i, j int) bool {
			return canonicalFrames[positions[i]].ID < canonicalFrames[positions[j]].ID
		})
		bySense[senseID] = positions
	}
	byFamily = make(map[string][]int)
	for index, binding := range canonicalBindings {
		byFamily[binding.FamilyID.String()] = append(byFamily[binding.FamilyID.String()], index)
	}
	return &Index{frames: canonicalFrames, bindings: canonicalBindings, candidates: canonicalCandidates, bySenseID: bySense, byFamilyID: byFamily}, nil
}

// FramesBySenseID returns every frame for a sense in stable ID order. It does
// not select a frame or merge multiple senses.
func (index *Index) FramesBySenseID(senseID ids.SenseID) []ValencyFrame {
	if index == nil || !validSenseID(senseID) {
		return []ValencyFrame{}
	}
	positions := index.bySenseID[senseID.String()]
	result := make([]ValencyFrame, len(positions))
	for offset, position := range positions {
		result[offset] = cloneFrame(index.frames[position])
	}
	return result
}

// BindingsByFamilyID returns all approved bindings for a family in stable
// sense/frame/variant order. Each binding still identifies its own sense.
func (index *Index) BindingsByFamilyID(familyID ids.ConstructionID) []ConstructionBinding {
	if index == nil || !validConstructionID(familyID) {
		return []ConstructionBinding{}
	}
	positions := index.byFamilyID[familyID.String()]
	result := make([]ConstructionBinding, len(positions))
	for offset, position := range positions {
		result[offset] = cloneBinding(index.bindings[position])
	}
	return result
}

// Frames returns all compiled frames in stable FrameID order.
func (index *Index) Frames() []ValencyFrame {
	if index == nil {
		return []ValencyFrame{}
	}
	result := make([]ValencyFrame, len(index.frames))
	for position, frame := range index.frames {
		result[position] = cloneFrame(frame)
	}
	return result
}

// Bindings returns all approved compiled bindings in stable identity order.
func (index *Index) Bindings() []ConstructionBinding {
	if index == nil {
		return []ConstructionBinding{}
	}
	result := make([]ConstructionBinding, len(index.bindings))
	for position, binding := range index.bindings {
		result[position] = cloneBinding(binding)
	}
	return result
}

// Candidates returns pending review proposals in stable identity order.
// They are not construction bindings and do not authorize runtime use.
func (index *Index) Candidates() []ConstructionCandidate {
	if index == nil {
		return []ConstructionCandidate{}
	}
	return append([]ConstructionCandidate(nil), index.candidates...)
}

func bindingKey(binding ConstructionBinding) string {
	return binding.SenseID.String() + "\x00" + string(binding.ValencyFrameID) + "\x00" + binding.FamilyID.String() + "\x00" + binding.VariantID.String()
}

func orderFrames(frames []ValencyFrame) {
	sort.Slice(frames, func(i, j int) bool {
		if frames[i].ID != frames[j].ID {
			return frames[i].ID < frames[j].ID
		}
		return frames[i].SenseID.String() < frames[j].SenseID.String()
	})
}

func orderBindings(bindings []ConstructionBinding) {
	sort.Slice(bindings, func(i, j int) bool {
		left, right := bindings[i], bindings[j]
		if left.SenseID.String() != right.SenseID.String() {
			return left.SenseID.String() < right.SenseID.String()
		}
		if left.ValencyFrameID != right.ValencyFrameID {
			return left.ValencyFrameID < right.ValencyFrameID
		}
		if left.FamilyID.String() != right.FamilyID.String() {
			return left.FamilyID.String() < right.FamilyID.String()
		}
		return left.VariantID.String() < right.VariantID.String()
	})
}

func orderCandidates(candidates []ConstructionCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.SenseID.String() != right.SenseID.String() {
			return left.SenseID.String() < right.SenseID.String()
		}
		if left.ValencyFrameID != right.ValencyFrameID {
			return left.ValencyFrameID < right.ValencyFrameID
		}
		if left.FamilyID.String() != right.FamilyID.String() {
			return left.FamilyID.String() < right.FamilyID.String()
		}
		return left.VariantID.String() < right.VariantID.String()
	})
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}

package importer

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/valency"
)

func validateValencyFrames(frames []valency.ValencyFrame, familyPositions map[string]int, packages []construction.ConstructionPackage, language string) error {
	for frameIndex, frame := range frames {
		if err := frame.Validate(); err != nil {
			return err
		}
		for familyIndex, familyID := range frame.AllowedFamilyIDs {
			packageIndex, exists := familyPositions[familyID.String()]
			if !exists {
				return unknown("UNKNOWN_ALLOWED_FAMILY", fmt.Sprintf("valency_frames[%d].allowed_family_ids[%d]", frameIndex, familyIndex), "valency frame refers to no construction package family")
			}
			if string(packages[packageIndex].Language) != language {
				return unsupported("FRAME_FAMILY_LANGUAGE_MISMATCH", fmt.Sprintf("valency_frames[%d].allowed_family_ids[%d]", frameIndex, familyIndex), "valency frame and family language must match")
			}
		}
	}
	return nil
}

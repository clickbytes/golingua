package pack

import "github.com/clickbytes/golingua/internal/staging"

// LoadActive revalidates and compiles the exact snapshot selected by the
// existing store pointer. It performs no snapshot activation or writes.
func LoadActive(store *staging.Store, datasetID string) (*Pack, error) {
	candidate, err := store.Active(datasetID)
	if err != nil {
		return nil, err
	}
	return Compile(candidate)
}

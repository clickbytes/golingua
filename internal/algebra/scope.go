package algebra

type checkerScope struct {
	active   map[string]Type
	declared map[string]struct{}
}

func newCheckerScope() *checkerScope {
	return &checkerScope{active: make(map[string]Type), declared: make(map[string]struct{})}
}

func (scope *checkerScope) bind(id string, value Type, path string) error {
	if _, exists := scope.declared[id]; exists {
		return invalid("DUPLICATE_BINDER_ID", path, "binder ID must be globally unique")
	}
	scope.declared[id] = struct{}{}
	scope.active[id] = value
	return nil
}

func (scope *checkerScope) unbind(id string) { delete(scope.active, id) }

func (scope *checkerScope) lookup(id string) (Type, bool) {
	value, ok := scope.active[id]
	return value, ok
}

package tools

// Base provides common functionality for tools
type Base struct {
	name        string
	description string
}

func (t *Base) Name() string {
	return t.name
}

func (t *Base) Description() string {
	return t.description
}

// NewBase creates a new base tool with the given name and description
func NewBase(name, description string) *Base {
	return &Base{
		name:        name,
		description: description,
	}
}

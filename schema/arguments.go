package schema

// Argument ...
type Argument struct {
	Name              string
	DefaultValue      interface{}
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
}


package schema

type InputObject struct {
}

type InputObjectField struct {
	Name         string
	Type         Type
	Description  string
	DefaultValue interface{}
}

package gql_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/language/ast"
)

func Test_StringScalarInput(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		hasErr   bool
	}{
		{
			name: "astInt",
			value: &ast.IntValue{
				Value: `3`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "string",
			value:    "string",
			expected: "string",
			hasErr:   false,
		},
		{
			name: "*string",
			value: func() *string {
				s := "foo"
				return &s
			}(),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "[]byte",
			value:    []byte("bar"),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "int",
			value:    int(42),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "bool",
			value:    true,
			expected: nil,
			hasErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gql.String
			v, err := s.CoerceInputFunc(tt.value)
			if !tt.hasErr && err != nil {
				t.Fatalf("expected NO error, got '%s'", err.Error())
			} else if tt.hasErr && err == nil {
				t.Fatalf("expected to fail, but got NO error")
			} else if !reflect.DeepEqual(v, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}

func Test_IntScalarInput(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		hasErr   bool
	}{
		{
			name: "astString",
			value: &ast.StringValue{
				Value: `"asd"`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astInt",
			value: &ast.IntValue{
				Value: `42`,
			},
			expected: 42,
			hasErr:   false,
		},
		{
			name:     "invalidString",
			value:    "string",
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "validString",
			value:    "42",
			expected: 42,
			hasErr:   false,
		},
		{
			name: "invalidStringPointer",
			value: func() *string {
				s := "foo"
				return &s
			}(),
			expected: nil,
			hasErr:   true,
		},
		{
			name: "validStringPointer",
			value: func() *string {
				s := "42"
				return &s
			}(),
			expected: 42,
			hasErr:   false,
		},
		{
			name:     "[]byte",
			value:    []byte("bar"),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "int",
			value:    int(42),
			expected: 42,
			hasErr:   false,
		},
		{
			name:     "bool",
			value:    true,
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "invalidFloat64",
			value:    float64(3.14),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "validFloat64",
			value:    float64(42),
			expected: 42,
			hasErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gql.Int
			v, err := s.CoerceInputFunc(tt.value)
			if !tt.hasErr && err != nil {
				t.Fatalf("expected NO error, got '%s'", err.Error())
			} else if tt.hasErr && err == nil {
				t.Fatalf("expected to fail, but got NO error")
			} else if !reflect.DeepEqual(v, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}

func Test_IDScalarInput(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		hasErr   bool
	}{
		{
			name: "astString",
			value: &ast.StringValue{
				Value: `"asd"`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astInt",
			value: &ast.IntValue{
				Value: `42`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "nonNumericString",
			value:    `foo`,
			expected: `foo`,
			hasErr:   false,
		},
		{
			name:     "numericString",
			value:    "42",
			expected: "42",
			hasErr:   false,
		},
		{
			name: "nonNumericStringPointer",
			value: func() *string {
				s := "foo"
				return &s
			}(),
			expected: nil,
			hasErr:   true,
		},
		{
			name: "numericStringPointer",
			value: func() *string {
				s := "42"
				return &s
			}(),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "[]byte",
			value:    []byte("bar"),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "bool",
			value:    true,
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "invalidFloat64",
			value:    float64(3.14),
			expected: nil,
			hasErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gql.ID
			v, err := s.CoerceInputFunc(tt.value)
			if !tt.hasErr && err != nil {
				t.Fatalf("expected NO error, got '%s'", err.Error())
			} else if tt.hasErr && err == nil {
				t.Fatalf("expected to fail, but got NO error")
			} else if !reflect.DeepEqual(v, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}

func Test_FloatScalarInput(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		hasErr   bool
	}{
		{
			name: "astString",
			value: &ast.StringValue{
				Value: `"asd"`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astInt",
			value: &ast.IntValue{
				Value: `42`,
			},
			expected: float64(42),
			hasErr:   false,
		},
		{
			name: "astFloat",
			value: &ast.FloatValue{
				Value: `3.14`,
			},
			expected: float64(3.14),
			hasErr:   false,
		},
		{
			name:     "invalidString",
			value:    "foo",
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "validString",
			value:    "42",
			expected: float64(42),
			hasErr:   false,
		},
		{
			name:     "validString",
			value:    "3.14",
			expected: float64(3.14),
			hasErr:   false,
		},
		{
			name: "invalidStringPointer",
			value: func() *string {
				s := "foo"
				return &s
			}(),
			expected: nil,
			hasErr:   true,
		},
		{
			name: "validStringPointer",
			value: func() *string {
				s := "42"
				return &s
			}(),
			expected: float64(42),
			hasErr:   false,
		},
		{
			name: "validStringPointer",
			value: func() *string {
				s := "3.14"
				return &s
			}(),
			expected: float64(3.14),
			hasErr:   false,
		},
		{
			name:     "[]byte",
			value:    []byte("42"),
			expected: float64(42),
			hasErr:   false,
		},
		{
			name:     "int",
			value:    int(42),
			expected: float64(42),
			hasErr:   false,
		},
		{
			name:     "bool",
			value:    true,
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "float64",
			value:    float64(3.14),
			expected: float64(3.14),
			hasErr:   false,
		},
		{
			name:     "float64",
			value:    float64(42),
			expected: float64(42),
			hasErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gql.Float
			v, err := s.CoerceInputFunc(tt.value)
			if !tt.hasErr && err != nil {
				t.Fatalf("expected NO error, got '%s'", err.Error())
			} else if tt.hasErr && err == nil {
				t.Fatalf("expected to fail, but got NO error")
			} else if !reflect.DeepEqual(v, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}

func Test_BooleanScalarInput(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		hasErr   bool
	}{
		{
			name: "astString",
			value: &ast.StringValue{
				Value: `"asd"`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astInt",
			value: &ast.IntValue{
				Value: `42`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astFloat",
			value: &ast.FloatValue{
				Value: `3.14`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astBoolean",
			value: &ast.BooleanValue{
				Value: `true`,
			},
			expected: true,
			hasErr:   false,
		},
		{
			name:     "invalidString",
			value:    "foo",
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "validString",
			value:    "true",
			expected: true,
			hasErr:   false,
		},
		{
			name:     "validString",
			value:    "false",
			expected: false,
			hasErr:   false,
		},
		{
			name:     "[]byte",
			value:    []byte("42"),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "int",
			value:    int(42),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "bool",
			value:    true,
			expected: true,
			hasErr:   false,
		},
		{
			name:     "bool",
			value:    false,
			expected: false,
			hasErr:   false,
		},
		{
			name:     "float64",
			value:    float64(3.14),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "float64",
			value:    float64(42),
			expected: nil,
			hasErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gql.Boolean
			v, err := s.CoerceInputFunc(tt.value)
			if !tt.hasErr && err != nil {
				t.Fatalf("expected NO error, got '%s'", err.Error())
			} else if tt.hasErr && err == nil {
				t.Fatalf("expected to fail, but got NO error")
			} else if !reflect.DeepEqual(v, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}

func Test_DateTimeScalarInput(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		hasErr   bool
	}{
		{
			name: "astString",
			value: &ast.StringValue{
				Value: `"2015-10-21T07:28:05Z"`,
			},
			expected: func() time.Time {
				t, _ := time.Parse(time.RFC3339, "2015-10-21T07:28:05Z")
				return t
			}(),
			hasErr: false,
		},
		{
			name: "astInt",
			value: &ast.IntValue{
				Value: `42`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astFloat",
			value: &ast.FloatValue{
				Value: `3.14`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name: "astBoolean",
			value: &ast.BooleanValue{
				Value: `true`,
			},
			expected: nil,
			hasErr:   true,
		},
		{
			name:  "validString",
			value: "2015-10-21T07:28:05Z",
			expected: func() time.Time {
				t, _ := time.Parse(time.RFC3339, "2015-10-21T07:28:05Z")
				return t
			}(),
			hasErr: false,
		},
		{
			name:     "invalidString",
			value:    "foo",
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "[]byte",
			value:    []byte("42"),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "int",
			value:    int(42),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "bool",
			value:    true,
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "bool",
			value:    false,
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "float64",
			value:    float64(3.14),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "float64",
			value:    float64(42),
			expected: nil,
			hasErr:   true,
		},
		{
			name:     "time.Time",
			value:    time.Unix(1445412485, 0), // 2015-10-21T07:28:05Z00:00
			expected: time.Unix(1445412485, 0), // 2015-10-21T07:28:05Z00:00
			hasErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := gql.DateTime
			v, err := s.CoerceInputFunc(tt.value)
			if !tt.hasErr && err != nil {
				t.Fatalf("expected NO error, got '%s'", err.Error())
			} else if tt.hasErr && err == nil {
				t.Fatalf("expected to fail, but got NO error")
			} else if !reflect.DeepEqual(v, tt.expected) {
				t.Fatalf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}

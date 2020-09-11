package lexer

import (
	"testing"
)

func TestStringValue(t *testing.T) {
	l := &Lexer{
		input: NewInput([]byte(`
		"Hello,\n  World!\n\nYours,\n  GraphQL."
		"""
	Hello,
	  World!
	
	Yours,
	  GraphQL.
  """
`)),
	}
	token1 := l.Read()
	token2 := l.Read()
	if token1.Value != token2.Value {
		t.Fatalf("expected '%s' == '%s'", token1.Value, token2.Value)
	}
}

func TestFloatAsInt(t *testing.T) {
	l := &Lexer{
		input: NewInput([]byte(`123`)),
	}
	token := l.Read()
	if token.Value != "123" || token.Kind != IntValueToken {
		t.Fatalf("expected '%s' == '%s'", token.Value, "123")
	}
}
func TestFloatAsSimpleFloat(t *testing.T) {
	l := &Lexer{
		input: NewInput([]byte(`123.123`)),
	}
	token := l.Read()
	if token.Value != "123.123" || token.Kind != FloatValueToken {
		t.Fatalf("expected '%s' == '%s' & '%v' == '%v'", token.Value, "123.123", token.Kind.String(), FloatValueToken.String())
	}
	t.Logf("got '%s' == '%s' & '%v' == '%v'", token.Value, "123.123", token.Kind.String(), FloatValueToken.String())
}
func TestFloatAsComplexFloat(t *testing.T) {
	l := &Lexer{
		input: NewInput([]byte(`123.123e+20`)),
	}
	token := l.Read()
	if token.Value != "123.123e+20" || token.Kind != FloatValueToken {
		t.Fatalf("expected '%s' == '%s' & '%v' == '%v'", token.Value, "123.123e+20", token.Kind.String(), FloatValueToken.String())
	}
	t.Logf("got '%s' == '%s' & '%v' == '%v'", token.Value, "123.123e+20", token.Kind.String(), FloatValueToken.String())
}

func TestBlockStringValue(t *testing.T) {
	block := ` 
    Hello,
      World!
    
    Yours,
      GraphQL.
	  `
	expect := "Hello,\n  World!\n\nYours,\n  GraphQL."
	res := string(BlockStringValue([]byte(block)))
	if expect != res {
		t.Fatalf("invalid res, \n'%s'\n'%s'", res, expect)
	}
}

func TestLexer(t *testing.T) {
	l := &Lexer{
		input: NewInput([]byte(`
		type Person {
			name(
				"""
				some example arg
				"""
				bar: String
	
				"some other arg"
				foo: Int
			): String
			age: Int
			picture: Url
		}
`)),
	}
	token := l.Read()
	for {
		t.Log(token.Kind, token.Value)
		if token.Kind == EOFToken {
			return
		}
		token = l.Read()
	}
}

var example string = `
type Character { 
	name: String!
	appearsIn: [Episode!]!
}`

var introspectionQuery = `query IntrospectionQuery {
	__schema {
	  queryType {
		name
	  }
	  mutationType {
		name
	  }
	  subscriptionType {
		name
	  }
	  types {
		...FullType
	  }
	  directives {
		name
		description
		locations
		args {
		  ...InputValue
		}
	  }
	}
  }
  
  fragment FullType on __Type {
	kind
	name
	description
	fields(includeDeprecated: true) {
	  name
	  description
	  args {
		...InputValue
	  }
	  type {
		...TypeRef
	  }
	  isDeprecated
	  deprecationReason
	}
	inputFields {
	  ...InputValue
	}
	interfaces {
	  ...TypeRef
	}
	enumValues(includeDeprecated: true) {
	  name
	  description
	  isDeprecated
	  deprecationReason
	}
	possibleTypes {
	  ...TypeRef
	}
  }
  
  fragment InputValue on __InputValue {
	name
	description
	type {
	  ...TypeRef
	}
	defaultValue
  }
  
  fragment TypeRef on __Type {
	kind
	name
	ofType {
	  kind
	  name
	  ofType {
		kind
		name
		ofType {
		  kind
		  name
		  ofType {
			kind
			name
			ofType {
			  kind
			  name
			  ofType {
				kind
				name
				ofType {
				  kind
				  name
				}
			  }
			}
		  }
		}
	  }
	}
  }`

func BenchmarkLexer(b *testing.B) {
	inputBytes := []byte(introspectionQuery)

	lexer := &Lexer{
		input: NewInput(inputBytes),
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.SetBytes(int64(len(inputBytes)))

	for i := 0; i < b.N; i++ {
		lexer.input.Reset()
		var kind TokenKind

		for kind != EOFToken {
			kind = lexer.Read().Kind
		}
	}
}

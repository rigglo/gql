package tokens

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF

	TokenExclamationMark
	TokenDollarSign
	TokenLeftBracket
	TokenRightBracket
	TokenThreeDot
	TokenColon
	TokenEqualSign
	TokenAtSign
	TokenLeftSquareBracket
	TokenRightSquareBracket
	TokenLeftCurlyBracket
	TokenVerticalBar
	TokenRightCurlyBracket

	TokenNewLine

	TokenQuery
	TokenMutation
	TokenSubscription

	TokenField
	TokenFieldArgument
	TokenFieldArgumentValue
	TokenFieldEnd

	TokenSelectionSetStart
	TokenSelectionSetEnd
)

/*

	// ExclamationMark is !
	ExclamationMark string = "!"
	// DollarSign is $
	DollarSign string = "$"
	// LeftBracket is (
	LeftBracket string = "("
	// RightBracket is )
	RightBracket string = ")"
	// ThreeDot is ...
	ThreeDot string = "..."
	// Colon is :
	Colon string = ":"
	// EqualSign is =
	EqualSign string = "="
	// AtSign is @
	AtSign string = "@"
	// LeftSquareBracket is [
	LeftSquareBracket string = "["
	// RightSquareBracket is ]
	RightSquareBracket string = "]"
	// LeftCurlyBracket is {
	LeftCurlyBracket string = "{"
	// VerticalBar is |
	VerticalBar string = "|"
	// RightCurlyBracket is }
	RightCurlyBracket string = "}"

	/*
			LineTerminator
		New Line (U+000A)
		Carriage Return (U+000D)New Line (U+000A)
		Carriage Return (U+000D)New Line (U+000A)


	// NewLine is \n
	NewLine string = "\n"

	/*
		Operations
		query	mutation	subscription


	// Query is the "query" operation
	Query string = "query"
	// Mutation is the "mutation" operation
	Mutation string = "mutation"
	// Subscription is the "subscription" operation
	Subscription string = "subscription"
)
*/

package lexer

const (
	// White space is used to improve legibility of source text and act as separation between tokens, and any amount of white space may appear before or after any token. White space between tokens is not significant to the semantic meaning of a GraphQL Document, however white space characters may appear within a String or Comment token
	runeHorizontalTab = '\u0009'
	runeSpace         = '\u0020'

	// Like white space, line terminators are used to improve the legibility of source text, any amount may appear before or after any other token and have no significance to the semantic meaning of a GraphQL Document. Line terminators are not found within any other token
	runeNewLine        = '\u000A'
	runeCarriageReturn = '\u000D'

	// The “Byte Order Mark” is a special Unicode character which may appear at the beginning of a file containing Unicode which programs may use to determine the fact that the text stream is Unicode, what endianness the text stream is in, and which of several Unicode encodings to interpret
	runeUnicodeBOM = '\uFEFF'

	// Similar to white space and line terminators, commas (,) are used to improve the legibility of source text and separate lexical tokens but are otherwise syntactically and semantically insignificant within GraphQL Documents
	runeComma = ','

	// Punctations
	runeExclamationMark  = '!'
	runeDollar           = '$'
	runeLeftParentheses  = '('
	runeRightParentheses = ')'
	runeLeftBrace        = '{'
	runeRightBrace       = '}'
	runeLeftBracket      = '['
	runeRightBracket     = ']'
	runeColon            = ':'
	runeEqual            = '='
	runeAt               = '@'
	runeVerticalBar      = '|'

	runeDot = '.'

	/* String values are single line "<value>" or multiline
	"""
	<value line 1>
	<value line 2>
	"""*/
	runeQuotation = '"'

	// Hashtag starts a new comment
	runeHashtag = '#'

	// EOF
	runeEOF = 0

	// AND rune
	runeAND = '&'

	// backSlask \ to escape things
	runeBackSlash = '\\'

	// unicode u
	runeU = 'u'

	// negative sign for numeric values
	runeNegativeSign = '-'
)

func isSourceCharacter(r rune) bool {
	return r == '\u0009' ||
		r == '\u000A' ||
		r == '\u000D' ||
		('\u0020' <= r && r <= '\uFFFF')
}

func isLineTerminator(r rune) bool {
	return r == runeNewLine || r == runeCarriageReturn
}

func isWhitespace(r rune) bool {
	return r == runeSpace || r == runeHorizontalTab
}

func isCommentCharacter(r rune) bool {
	return isSourceCharacter(r) && !isLineTerminator(r)
}

func isPunctuator(r rune) bool {
	return r == runeExclamationMark ||
		r == runeDollar ||
		r == runeLeftParentheses ||
		r == runeRightParentheses ||
		r == runeLeftBrace ||
		r == runeRightBrace ||
		r == runeLeftBracket ||
		r == runeRightBracket ||
		r == runeColon ||
		r == runeEqual ||
		r == runeAt ||
		r == runeVerticalBar ||
		r == runeAND
}

func isNameStart(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || r == '_'
}

func isName(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '_'
}

// comment is ignorable as well, but should be checked separately
func canIgnore(r rune) bool {
	return r == runeUnicodeBOM || isWhitespace(r) || isLineTerminator(r) || r == runeComma
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isNonZeroDigit(r rune) bool {
	return '1' <= r && r <= '9'
}

// TODO: check for triple dots

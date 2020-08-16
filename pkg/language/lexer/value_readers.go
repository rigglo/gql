package lexer

import (
	"bytes"
	"log"
)

// String value
func (l *Lexer) readStringValue(t *Token) {
	t.Kind = StringValueToken
	if l.peekEqual(runeQuotation, runeQuotation) {
		l.readStringBlock(t)
	} else {
		l.readSingleLineString(t)
	}
}

func (l *Lexer) readStringBlock(t *Token) {
	l.input.Pos += 3
	t.Start = l.input.Pos
	for {
		if isSourceCharacter(l.input.PeekOne(0)) &&
			!(l.input.PeekOne(0) == runeQuotation && l.peekEqual(runeQuotation, runeQuotation)) &&
			!(l.input.PeekOne(0) == runeBackSlash && l.peekEqual(runeQuotation, runeQuotation, runeQuotation)) {
			l.input.Pos++
		} else if l.input.PeekOne(0) == runeQuotation && l.peekEqual(runeQuotation, runeQuotation) {
			t.End = l.input.Pos
			t.Value = string(BlockStringValue(l.input.raw[t.Start:t.End])) // string(l.input.raw[t.Start:t.End]) //
			l.input.Pos += 3
			return
		} else {

			//log.Printf("%v", l.input.PeekOne(0))
			//log.Printf("got to else in block string: '%s'", string([]rune{l.input.PeekOne(0)}))
			panic("this should not happen, undefined token")
			// TODO: return undefined token
		}
	}
}

func (l *Lexer) readSingleLineString(t *Token) {
	l.input.Pos++
	t.Start = l.input.Pos
	for {
		if isSourceCharacter(l.input.PeekOne(0)) && l.input.PeekOne(0) != runeQuotation && l.input.PeekOne(0) != runeBackSlash && !isLineTerminator(l.input.PeekOne(0)) {
			t.Value += string(l.input.PeekOne(0))
			l.input.Pos++
		} else if l.input.PeekOne(0) == runeBackSlash &&
			l.peekRules(
				func(r rune) bool {
					return r == runeU
				},
				func(r rune) bool {
					return ('0' <= r && r <= '9') || ('A' <= r && r <= 'F') || ('a' <= r && r <= 'f')
				},
				func(r rune) bool {
					return ('0' <= r && r <= '9') || ('A' <= r && r <= 'F') || ('a' <= r && r <= 'f')
				},
				func(r rune) bool {
					return ('0' <= r && r <= '9') || ('A' <= r && r <= 'F') || ('a' <= r && r <= 'f')
				},
				func(r rune) bool {
					return ('0' <= r && r <= '9') || ('A' <= r && r <= 'F') || ('a' <= r && r <= 'f')
				},
			) {
			t.Value += string(l.input.PeekOne(0))
			l.input.Pos += 6
		} else if l.input.PeekOne(0) == runeBackSlash &&
			l.peekRules(
				func(r rune) bool {
					return r == runeQuotation ||
						r == runeBackSlash ||
						r == '/' ||
						r == 'b' ||
						r == 'f' ||
						r == 'n' ||
						r == 'r' ||
						r == 't'
				},
			) {
			switch l.input.PeekOne(1) {
			case '\\':
				t.Value += "\\"
			case '/':
				t.Value += "/"
			case 'b':
				t.Value += "\b"
			case 'f':
				t.Value += "\f"
			case 'n':
				t.Value += "\n"
			case 'r':
				t.Value += "\r"
			case 't':
				t.Value += "\t"
			}
			l.input.Pos += 2
		} else if l.input.PeekOne(0) == runeQuotation {
			t.End = l.input.Pos
			l.input.Pos++
			return
		} else {
			// TODO: return undefined token
		}
	}
}

func BlockStringValue(s []byte) (formatted []byte) {
	formatted = []byte{}
	lines := bytes.FieldsFunc(s, isLineTerminator)
	commonIndent := 0
	for i, line := range lines {
		if i == 0 {
			continue
		}
		indent := len(line) - len(bytes.TrimLeftFunc(line, isWhitespace))
		if indent < len(line) {
			if commonIndent == 0 || indent < commonIndent {
				commonIndent = indent
			}
		}
	}
	log.Println("commonIndent", commonIndent)
	if len(lines) != 0 && len(bytes.TrimLeftFunc(lines[0], isWhitespace)) == 0 {
		if len(lines) > 1 {
			lines = lines[1:]
		} else {
			lines = [][]byte{}
		}
	}
	if commonIndent != 0 {
		for i, line := range lines {
			//log.Printf("line: %v, indent: %v", len(line), commonIndent)
			if len(line) >= commonIndent {
				lines[i] = []byte(line)[commonIndent:]
			}
			log.Printf("line: '%s'", lines[i])
		}
	}
	if len(lines) != 0 && len(bytes.TrimLeftFunc(lines[len(lines)-1], isWhitespace)) == 0 {
		if len(lines) > 1 {
			lines = lines[:len(lines)-1]
		} else {
			lines = [][]byte{}
		}
	}
	for i, l := range lines {
		if i == 0 {
			formatted = append(formatted, l...)
			continue
		}
		formatted = append(formatted, '\n')
		formatted = append(formatted, l...)
	}
	return
}

// Numeric value
func (l *Lexer) readIntValue(t *Token) bool {
	defer func() {
		t.End = l.input.Pos
	}()
	if l.peekEqual(runeNegativeSign) {
		l.input.Pos++
	}
	if l.input.PeekOne(0) == '0' {
		if l.peekRules(func(r rune) bool {
			return isDigit(r) || r == runeDot || isNameStart(r)
		}) {
			// TODO: return undefined token error
			l.input.Pos = t.Start
			return false
		}
		l.input.Pos += 2
	} else if isNonZeroDigit(l.input.PeekOne(0)) {
		l.input.Pos++
		for isDigit(l.input.PeekOne(0)) {
			l.input.Pos++
		}
		if l.peekRules(func(r rune) bool {
			return r == runeDot || isNameStart(r)
		}) {
			// TODO: return undefined token error
			l.input.Pos = t.Start
			return false
		}
	}
	t.Kind = IntValueToken
	t.End = l.input.Pos
	return true
}

// Numeric value
func (l *Lexer) readFloatValue(t *Token) bool {
	defer func() {
		t.End = l.input.Pos
	}()
	if l.peekEqual(runeNegativeSign) {
		l.input.Pos++
	}
	if l.input.PeekOne(0) == '0' {
		if l.peekRules(func(r rune) bool {
			return isDigit(r) || r == runeDot || isNameStart(r)
		}) {
			// TODO: return undefined token error
			l.input.Pos = t.Start
			return false
		}
		l.input.Pos += 2
	} else if isNonZeroDigit(l.input.PeekOne(0)) {
		l.input.Pos++
		for isDigit(l.input.PeekOne(0)) {
			l.input.Pos++
		}
		if l.input.PeekOne(0) == runeDot || isNameStart(l.input.PeekOne(0)) {
			// TODO: return undefined token error
			l.input.Pos = t.Start
			return false
		}
	}
	t.Kind = FloatValueToken
	t.End = l.input.Pos
	return true
}

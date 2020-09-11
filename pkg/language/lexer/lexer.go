package lexer

type Lexer struct {
	input *Input
}

func NewLexer(in *Input) *Lexer {
	return &Lexer{
		input: in,
	}
}

func (l *Lexer) Read() (t Token) {
	defer func() {
		if t.Kind != EOFToken && t.Value == "" {
			t.Value = string(l.input.raw[t.Start:t.End])
		}
	}()

	l.ignore()
	t.Start = l.input.Pos
	if l.isSingleCharacterToken(&t) {
		return
	}
	r := l.input.PeekOne(0)
	switch {
	case runeDot == r:
		l.readDot(&t)
		return
	case runeQuotation == r:
		l.readStringValue(&t)
		return
	case isDigit(r) || runeNegativeSign == r:
		if l.readIntValue(&t) {
			return
		} else if l.readFloatValue(&t) {
			return
		}
		// TODO: undefined token
		return
	}

	l.readName(&t)
	return
}

func (l *Lexer) ignore() {
	if canIgnore(l.input.PeekOne(0)) {
		l.input.Pos++
		l.ignore()
	} else if l.input.PeekOne(0) == runeHashtag {
		for isCommentCharacter(l.input.PeekOne(0)) {
			l.input.Pos++
		}
		l.ignore()
	}
}

func (l *Lexer) peekEqual(bs ...rune) bool {
	for i := 0; i < len(bs); i++ {
		if l.input.PeekOne(i+1) != bs[i] {
			return false
		}
	}
	return true
}

func (l *Lexer) peekRules(bs ...func(rune) bool) bool {
	for i := 0; i < len(bs); i++ {
		if !bs[i](l.input.PeekOne(i)) {
			return false
		}
	}
	return true
}

func (l *Lexer) isSingleCharacterToken(t *Token) bool {
	if l.input.PeekOne(0) == runeEOF {
		t.Kind = EOFToken
	} else if isPunctuator(l.input.PeekOne(0)) {
		t.Kind = PunctuatorToken
		t.Value = string(l.input.raw[l.input.Pos : l.input.Pos+1])
	} else {
		return false
	}
	l.input.Pos++
	t.End = l.input.Pos
	return true
}

func (l *Lexer) readName(t *Token) {
	t.Kind = NameToken
	defer func() {
		t.End = l.input.Pos
	}()
	for {
		if l.input.Pos > t.Start && isName(l.input.PeekOne(0)) {
			l.input.Pos++
		} else if l.input.Pos == t.Start && isNameStart(l.input.PeekOne(0)) {
			l.input.Pos++
		} else {
			return
		}
	}
}

func (l *Lexer) readDot(t *Token) {
	t.Kind = PunctuatorToken
	l.input.Pos += 3
	t.End = l.input.Pos
}

package lexer

import (
	"unicode/utf8"
)

type Input struct {
	raw    []byte
	Pos    int
	Line   int
	Column int
}

func NewInput(bs []byte) *Input {
	return &Input{
		raw: bs,
	}
}

func (i *Input) Reset() {
	i.Pos = 0
	i.Line = 0
	i.Column = 0
}

func (i *Input) Value(t *Token) []byte {
	return i.raw[t.Start:t.End]
}

func (i *Input) PeekOneRune(n int) (rune, int) {
	return utf8.DecodeRune(i.raw[i.Pos+n:])
}

func (i *Input) PeekOne(n int) rune {
	if i.Pos+n >= len(i.raw) {
		return runeEOF
	}
	return rune(i.raw[i.Pos+n])
}

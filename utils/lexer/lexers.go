package lexer

import (
	"strings"

	"github.com/rigglo/gql/utils/lexer/tokens"
)

func LexSearch(l *Lexer) LexFn {
	l.SkipIgnoredTokens()
	if strings.HasPrefix(l.InputToEnd(), tokens.LeftCurlyBracket) {
		l.Pos += len(tokens.LeftCurlyBracket)
		l.Start = l.Pos
		return LexFields
	} else if strings.HasPrefix(l.InputToEnd(), tokens.Query) {
		return LexQuery
	} else if strings.HasPrefix(l.InputToEnd(), tokens.Mutation) {
		return LexMutation
	}
	return l.Errorf("Couldn't find anything in LexSearch")
}

func LexQuery(l *Lexer) LexFn {
	l.Pos += len(tokens.Query)
	l.Emit(tokens.TokenQuery)

	l.SkipIgnoredTokens()

	if strings.HasPrefix(l.InputToEnd(), tokens.LeftCurlyBracket) {
		l.Pos += len(tokens.LeftCurlyBracket)
		l.Emit(tokens.TokenSelectionSetStart)
		return LexFields
	}
	return l.Errorf(tokens.ErrNoSubselection)
}

func LexMutation(l *Lexer) LexFn {
	l.Pos += len(tokens.Mutation)
	l.Emit(tokens.TokenMutation)

	l.SkipIgnoredTokens()

	if strings.HasPrefix(l.InputToEnd(), tokens.LeftCurlyBracket) {
		l.Pos += len(tokens.LeftCurlyBracket)
		l.Start = l.Pos
		return LexFields
	}
	return l.Errorf(tokens.ErrNoSubselection)
}

func LexFields(l *Lexer) LexFn {
	l.SkipIgnoredTokens()
	if l.IsEOF() {
		return l.Errorf(tokens.ErrUnexpectedEOF)
	}

	if strings.HasPrefix(l.InputToEnd(), tokens.RightCurlyBracket) {

		l.Next()
		l.Emit(tokens.TokenSelectionSetEnd)
		if l.Start == len(l.Input) {
			l.Emit(tokens.TokenEOF)
			return nil
		}
		return LexFields
	} else if strings.HasPrefix(l.InputToEnd(), tokens.LeftCurlyBracket) {
		l.Pos += len(tokens.LeftCurlyBracket)
		l.Emit(tokens.TokenSelectionSetStart)
		return LexFields
	}
	return LexField
}

func LexField(l *Lexer) LexFn {
	for {
		if strings.HasPrefix(l.InputToEnd(), tokens.NewLine) {
			l.Emit(tokens.TokenField)
			l.Emit(tokens.TokenFieldEnd)
			return LexFields
		} else if strings.HasPrefix(l.InputToEnd(), tokens.LeftBracket) {
			l.Emit(tokens.TokenField)
			l.Pos += len(tokens.LeftBracket)
			l.Start = l.Pos
			return LexArgument
		} else if strings.HasPrefix(l.InputToEnd(), tokens.Colon) {
			l.Emit(tokens.TokenAlias)
			l.Pos += len(tokens.Colon)
			l.Start = l.Pos
		} else if strings.HasPrefix(l.InputToEnd(), tokens.LeftCurlyBracket) {
			l.Emit(tokens.TokenField)
			l.Pos += len(tokens.LeftCurlyBracket)
			l.Emit(tokens.TokenSelectionSetStart)
			return LexFields
		} else if strings.HasPrefix(l.InputToEnd(), tokens.RightCurlyBracket) {
			//l.Inc()
			l.Pos += len(tokens.RightCurlyBracket)
			l.Emit(tokens.TokenSelectionSetEnd)
			return LexFields
		}

		l.Inc()
	}
}

func LexArgument(l *Lexer) LexFn {
	l.SkipIgnoredTokens()
	for {
		//log.Print("LexArgument:", l.InputToEnd())
		if strings.HasPrefix(l.InputToEnd(), tokens.Colon) {
			l.Emit(tokens.TokenFieldArgument)
			l.Pos += len(tokens.Colon)
			l.Start = l.Pos
			return LexArgumentValue
		}

		l.Inc()

		if l.IsEOF() {
			return l.Errorf(tokens.ErrUnexpectedEOF)
		}
	}
}

func LexArgumentValue(l *Lexer) LexFn {
	l.SkipIgnoredTokens()
	for {
		//log.Print("LexArgumentValue:", l.InputToEnd())
		if strings.HasPrefix(l.InputToEnd(), tokens.RightBracket) {
			l.Emit(tokens.TokenFieldArgumentValue)
			l.Inc()
			if strings.HasPrefix(l.InputToEnd(), tokens.NewLine) {
				l.Emit(tokens.TokenFieldEnd)
				return LexFields
			}
			return LexFields
		} else if strings.HasPrefix(l.InputToEnd(), tokens.Comma) {
			l.Emit(tokens.TokenFieldArgumentValue)
			l.Inc()
			return LexArgument
		} else if false {
			// TODO: check if it's an ArgumentGroup
		}

		l.Inc()

		if l.IsEOF() {
			return l.Errorf(tokens.ErrUnexpectedEOF)
		}
	}
}

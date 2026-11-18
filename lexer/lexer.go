// Package lexer turns Bombo source code into tokens.
//
// Keywords are either single words ("mek", "bombo") or short phrases
// ("sen back", "tek a chance"). Phrases are matched before single words.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	EOF TokenType = iota
	NEWLINE
	IDENT
	NUMBER
	STRING

	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	ASSIGN
	PLUS_ASSIGN
	MINUS_ASSIGN
	EQ
	NEQ
	LT
	GT
	LTE
	GTE
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET
	COMMA
	COLON
	SEMICOLON

	LET      // mek
	PRINT    // seh
	IF       // bombo
	ELIF     // bomboclaat
	ELSE     // bloodclaat
	WHILE    // gwaan
	FOR      // fi
	IN       // inna
	FROM     // from
	TO       // tuh
	FUNC     // wuk
	RETURN   // sen back
	BREAK    // rassclaat
	CONTINUE // gallang
	AND      // an
	OR       // or
	NOT      // nuh
	TRUE     // yeh mon
	FALSE    // nah mon
	NIL      // kakafaat
	TRY      // tek a chance
	CATCH    // if it bruk
	THROW    // bruk
	EXIT     // likkle more
	HELLO    // wah gwaan
	INCR     // big up
	DECR     // bun
	COMMENT  // watch yah
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

type phrase struct {
	words []string
	typ   TokenType
}

// Phrases must come before any single word they start with.
var keywords = []phrase{
	{[]string{"sen", "back"}, RETURN},
	{[]string{"yeh", "mon"}, TRUE},
	{[]string{"nah", "mon"}, FALSE},
	{[]string{"tek", "a", "chance"}, TRY},
	{[]string{"if", "it", "bruk"}, CATCH},
	{[]string{"likkle", "more"}, EXIT},
	{[]string{"wah", "gwaan"}, HELLO},
	{[]string{"big", "up"}, INCR},
	{[]string{"watch", "yah"}, COMMENT},
	{[]string{"mek"}, LET},
	{[]string{"seh"}, PRINT},
	{[]string{"bombo"}, IF},
	{[]string{"bomboclaat"}, ELIF},
	{[]string{"bloodclaat"}, ELSE},
	{[]string{"fi"}, FOR},
	{[]string{"inna"}, IN},
	{[]string{"from"}, FROM},
	{[]string{"tuh"}, TO},
	{[]string{"wuk"}, FUNC},
	{[]string{"rassclaat"}, BREAK},
	{[]string{"gwaan"}, WHILE},
	{[]string{"gallang"}, CONTINUE},
	{[]string{"an"}, AND},
	{[]string{"or"}, OR},
	{[]string{"nuh"}, NOT},
	{[]string{"kakafaat"}, NIL},
	{[]string{"bruk"}, THROW},
	{[]string{"bun"}, DECR},
}

// Error is a lexing error.
type Error struct {
	Line int
	Msg  string
}

func (e *Error) Error() string { return fmt.Sprintf("Line %d: %s", e.Line, e.Msg) }

type lexer struct {
	src   []rune
	pos   int
	line  int
	depth int // newlines inside ( ) and [ ] are ignored
}

// Tokenize converts source into a slice of tokens ending with EOF.
func Tokenize(src string) ([]Token, error) {
	l := &lexer{src: []rune(src), line: 1}
	var toks []Token
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		if tok.Type == NEWLINE {
			if l.depth > 0 || len(toks) == 0 || toks[len(toks)-1].Type == NEWLINE {
				continue
			}
		}
		toks = append(toks, tok)
		if tok.Type == EOF {
			return toks, nil
		}
	}
}

func isLetter(c rune) bool { return unicode.IsLetter(c) || c == '_' }
func isDigit(c rune) bool  { return c >= '0' && c <= '9' }

func (l *lexer) skipLine() {
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
}

func (l *lexer) skipSpaces() {
	for l.pos < len(l.src) && (l.src[l.pos] == ' ' || l.src[l.pos] == '\t') {
		l.pos++
	}
}

func (l *lexer) readWord() string {
	start := l.pos
	for l.pos < len(l.src) && (isLetter(l.src[l.pos]) || isDigit(l.src[l.pos])) {
		l.pos++
	}
	return string(l.src[start:l.pos])
}

func (l *lexer) next() (Token, error) {
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if c == ' ' || c == '\t' || c == '\r' {
			l.pos++
			continue
		}
		if c == '#' {
			l.skipLine()
			continue
		}
		break
	}
	line := l.line
	if l.pos >= len(l.src) {
		return Token{EOF, "", line}, nil
	}
	c := l.src[l.pos]

	switch {
	case c == '\n':
		l.pos++
		l.line++
		return Token{NEWLINE, "\\n", line}, nil
	case isDigit(c):
		start := l.pos
		for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
			l.pos++
		}
		if l.pos+1 < len(l.src) && l.src[l.pos] == '.' && isDigit(l.src[l.pos+1]) {
			l.pos++
			for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
				l.pos++
			}
		}
		return Token{NUMBER, string(l.src[start:l.pos]), line}, nil
	case isLetter(c):
		tok := l.keywordOrIdent(l.readWord(), line)
		if tok.Type == COMMENT {
			l.skipLine()
			return l.next()
		}
		return tok, nil
	case c == '"' || c == '\'':
		return l.str(c)
	}

	if l.pos+1 < len(l.src) {
		two := string(l.src[l.pos : l.pos+2])
		types := map[string]TokenType{"==": EQ, "!=": NEQ, "<=": LTE, ">=": GTE, "+=": PLUS_ASSIGN, "-=": MINUS_ASSIGN}
		if t, ok := types[two]; ok {
			l.pos += 2
			return Token{t, two, line}, nil
		}
	}

	l.pos++
	singles := map[rune]TokenType{
		'+': PLUS, '-': MINUS, '*': STAR, '/': SLASH, '%': PERCENT, '=': ASSIGN,
		'<': LT, '>': GT, '(': LPAREN, ')': RPAREN, '{': LBRACE, '}': RBRACE,
		'[': LBRACKET, ']': RBRACKET, ',': COMMA, ':': COLON, ';': SEMICOLON,
	}
	t, ok := singles[c]
	if !ok {
		return Token{}, &Error{line, fmt.Sprintf("Wah dis? Mi nuh know dis character '%c'", c)}
	}
	switch t {
	case LPAREN, LBRACKET:
		l.depth++
	case RPAREN, RBRACKET:
		if l.depth > 0 {
			l.depth--
		}
	}
	return Token{t, string(c), line}, nil
}

func (l *lexer) keywordOrIdent(word string, line int) Token {
	for _, p := range keywords {
		if p.words[0] != word {
			continue
		}
		save := l.pos
		matched := true
		for _, w := range p.words[1:] {
			l.skipSpaces()
			if l.readWord() != w {
				matched = false
				break
			}
		}
		if matched {
			return Token{p.typ, strings.Join(p.words, " "), line}
		}
		l.pos = save
	}
	return Token{IDENT, word, line}
}

func (l *lexer) str(quote rune) (Token, error) {
	line := l.line
	l.pos++ // opening quote
	var sb strings.Builder
	for {
		if l.pos >= len(l.src) {
			return Token{}, &Error{line, "Yuh never close di string! Put a " + string(quote) + " pon di end"}
		}
		c := l.src[l.pos]
		l.pos++
		if c == quote {
			break
		}
		if c == '\n' {
			l.line++
		}
		if c == '\\' && l.pos < len(l.src) {
			e := l.src[l.pos]
			l.pos++
			switch e {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '\'':
				sb.WriteRune('\'')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(e)
			}
			continue
		}
		sb.WriteRune(c)
	}
	return Token{STRING, sb.String(), line}, nil
}

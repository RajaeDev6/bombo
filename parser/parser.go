// Package parser builds a Bombo syntax tree from tokens.
package parser

import (
	"fmt"
	"strconv"

	"bombo/ast"
	"bombo/lexer"
)

// Error is a syntax error.
type Error struct {
	Line int
	Msg  string
}

func (e *Error) Error() string { return fmt.Sprintf("Line %d: %s", e.Line, e.Msg) }

type parser struct {
	toks []lexer.Token
	pos  int
}

// Parse parses a whole Bombo program.
func Parse(src string) (prog *ast.Program, err error) {
	toks, lerr := lexer.Tokenize(src)
	if lerr != nil {
		return nil, lerr
	}
	p := &parser{toks: toks}
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*Error); ok {
				prog, err = nil, pe
				return
			}
			panic(r)
		}
	}()
	prog = &ast.Program{}
	p.skipNewlines()
	for !p.at(lexer.EOF) {
		prog.Stmts = append(prog.Stmts, p.statement())
		p.skipNewlines()
	}
	return prog, nil
}

func (p *parser) peek() lexer.Token { return p.toks[p.pos] }

func (p *parser) peekAt(n int) lexer.Token {
	if p.pos+n < len(p.toks) {
		return p.toks[p.pos+n]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parser) at(t lexer.TokenType) bool { return p.peek().Type == t }

func (p *parser) advance() lexer.Token {
	t := p.toks[p.pos]
	if t.Type != lexer.EOF {
		p.pos++
	}
	return t
}

func (p *parser) fail(format string, args ...any) {
	panic(&Error{Line: p.peek().Line, Msg: fmt.Sprintf(format, args...)})
}

func (p *parser) expect(t lexer.TokenType, what string) lexer.Token {
	if !p.at(t) {
		p.fail("Mi did expect %s but mi get %s", what, describe(p.peek()))
	}
	return p.advance()
}

func describe(t lexer.Token) string {
	switch t.Type {
	case lexer.EOF:
		return "di end a di file"
	case lexer.NEWLINE:
		return "a new line"
	case lexer.STRING:
		return strconv.Quote(t.Literal)
	}
	return "'" + t.Literal + "'"
}

func (p *parser) skipNewlines() {
	for p.at(lexer.NEWLINE) || p.at(lexer.SEMICOLON) {
		p.advance()
	}
}

func (p *parser) atStmtEnd() bool {
	switch p.peek().Type {
	case lexer.NEWLINE, lexer.SEMICOLON, lexer.RBRACE, lexer.EOF:
		return true
	}
	return false
}

func (p *parser) endStmt() {
	switch p.peek().Type {
	case lexer.NEWLINE, lexer.SEMICOLON:
		p.advance()
	case lexer.RBRACE, lexer.EOF:
	default:
		p.fail("Di line fi done here, but mi see %s. Wah yuh a try do?", describe(p.peek()))
	}
}

func (p *parser) statement() ast.Stmt {
	tok := p.peek()
	pos := ast.At(tok.Line)
	switch tok.Type {
	case lexer.LET:
		p.advance()
		name := p.expect(lexer.IDENT, "a name after 'mek'").Literal
		p.expect(lexer.ASSIGN, "'=' after 'mek "+name+"'")
		val := p.expression()
		p.endStmt()
		return &ast.LetStmt{Pos: pos, Name: name, Value: val}

	case lexer.PRINT:
		p.advance()
		var args []ast.Expr
		if !p.atStmtEnd() {
			args = append(args, p.expression())
			for p.at(lexer.COMMA) {
				p.advance()
				args = append(args, p.expression())
			}
		}
		p.endStmt()
		return &ast.PrintStmt{Pos: pos, Args: args}

	case lexer.IF:
		return p.ifStmt()

	case lexer.ELIF, lexer.ELSE:
		p.fail("'%s' cyaan stan up by itself, it need a 'bombo' first", tok.Literal)

	case lexer.WHILE:
		p.advance()
		cond := p.expression()
		return &ast.WhileStmt{Pos: pos, Cond: cond, Body: p.block()}

	case lexer.FOR:
		return p.forStmt()

	case lexer.FUNC:
		if p.peekAt(1).Type == lexer.IDENT {
			p.advance()
			name := p.advance().Literal
			params := p.params()
			return &ast.FuncStmt{Pos: pos, Name: name, Params: params, Body: p.block()}
		}

	case lexer.RETURN:
		p.advance()
		var val ast.Expr
		if !p.atStmtEnd() {
			val = p.expression()
		}
		p.endStmt()
		return &ast.ReturnStmt{Pos: pos, Value: val}

	case lexer.BREAK:
		p.advance()
		p.endStmt()
		return &ast.BreakStmt{Pos: pos}

	case lexer.CONTINUE:
		p.advance()
		p.endStmt()
		return &ast.ContinueStmt{Pos: pos}

	case lexer.EXIT:
		p.advance()
		p.endStmt()
		return &ast.ExitStmt{Pos: pos}

	case lexer.HELLO:
		p.advance()
		p.endStmt()
		return &ast.HelloStmt{Pos: pos}

	case lexer.TRY:
		p.advance()
		body := p.block()
		p.skipNewlines()
		p.expect(lexer.CATCH, "'if it bruk' after di 'tek a chance' block")
		name := ""
		if p.at(lexer.IDENT) {
			name = p.advance().Literal
		}
		return &ast.TryStmt{Pos: pos, Body: body, ErrName: name, Catch: p.block()}

	case lexer.THROW:
		p.advance()
		val := p.expression()
		p.endStmt()
		return &ast.ThrowStmt{Pos: pos, Value: val}

	case lexer.INCR, lexer.DECR:
		p.advance()
		target := p.postfix()
		p.checkTarget(target)
		p.endStmt()
		delta := 1.0
		if tok.Type == lexer.DECR {
			delta = -1
		}
		return &ast.IncStmt{Pos: pos, Target: target, Delta: delta}
	}

	// expression or assignment
	x := p.expression()
	switch p.peek().Type {
	case lexer.ASSIGN, lexer.PLUS_ASSIGN, lexer.MINUS_ASSIGN:
		op := p.advance().Literal
		p.checkTarget(x)
		val := p.expression()
		p.endStmt()
		return &ast.AssignStmt{Pos: pos, Target: x, Op: op, Value: val}
	}
	p.endStmt()
	return &ast.ExprStmt{Pos: pos, X: x}
}

func (p *parser) checkTarget(x ast.Expr) {
	switch x.(type) {
	case *ast.Ident, *ast.Index:
		return
	}
	panic(&Error{Line: x.Line(), Msg: "Yuh cyaan put a value inna dat. Only variables an list/map slots"})
}

func (p *parser) ifStmt() ast.Stmt {
	tok := p.advance() // bombo / bomboclaat
	cond := p.expression()
	stmt := &ast.IfStmt{Pos: ast.At(tok.Line), Cond: cond, Then: p.block()}

	save := p.pos
	p.skipNewlines()
	switch p.peek().Type {
	case lexer.ELIF:
		stmt.Else = p.ifStmt()
	case lexer.ELSE:
		p.advance()
		if p.at(lexer.IF) {
			stmt.Else = p.ifStmt()
		} else {
			stmt.Else = p.block()
		}
	default:
		p.pos = save
	}
	return stmt
}

func (p *parser) forStmt() ast.Stmt {
	tok := p.advance()
	pos := ast.At(tok.Line)
	name := p.expect(lexer.IDENT, "a variable name after 'fi'").Literal
	switch {
	case p.at(lexer.IN):
		p.advance()
		iter := p.expression()
		return &ast.ForInStmt{Pos: pos, Var: name, Iter: iter, Body: p.block()}
	case p.at(lexer.FROM):
		p.advance()
		from := p.expression()
		p.expect(lexer.TO, "'tuh' (like: fi i from 1 tuh 10)")
		to := p.expression()
		return &ast.ForRangeStmt{Pos: pos, Var: name, From: from, To: to, Body: p.block()}
	}
	p.fail("After 'fi %s' mi expect 'inna' or 'from', but mi get %s", name, describe(p.peek()))
	return nil
}

func (p *parser) params() []string {
	p.expect(lexer.LPAREN, "'(' fi start di function params")
	var params []string
	for !p.at(lexer.RPAREN) {
		params = append(params, p.expect(lexer.IDENT, "a param name").Literal)
		if !p.at(lexer.COMMA) {
			break
		}
		p.advance()
	}
	p.expect(lexer.RPAREN, "')' fi close di function params")
	return params
}

func (p *parser) block() *ast.Block {
	open := p.expect(lexer.LBRACE, "'{' fi start di block")
	b := &ast.Block{Pos: ast.At(open.Line)}
	p.skipNewlines()
	for !p.at(lexer.RBRACE) {
		if p.at(lexer.EOF) {
			panic(&Error{Line: open.Line, Msg: "Yuh open a block wid '{' but never close it wid '}'"})
		}
		b.Stmts = append(b.Stmts, p.statement())
		p.skipNewlines()
	}
	p.advance()
	return b
}

func (p *parser) expression() ast.Expr { return p.or() }

func (p *parser) or() ast.Expr {
	x := p.and()
	for p.at(lexer.OR) {
		t := p.advance()
		x = &ast.Logical{Pos: ast.At(t.Line), Op: "or", L: x, R: p.and()}
	}
	return x
}

func (p *parser) and() ast.Expr {
	x := p.equality()
	for p.at(lexer.AND) {
		t := p.advance()
		x = &ast.Logical{Pos: ast.At(t.Line), Op: "an", L: x, R: p.equality()}
	}
	return x
}

func (p *parser) binaryLevel(next func() ast.Expr, types ...lexer.TokenType) ast.Expr {
	x := next()
	for {
		matched := false
		for _, t := range types {
			if p.at(t) {
				tok := p.advance()
				x = &ast.Binary{Pos: ast.At(tok.Line), Op: tok.Literal, L: x, R: next()}
				matched = true
				break
			}
		}
		if !matched {
			return x
		}
	}
}

func (p *parser) equality() ast.Expr {
	return p.binaryLevel(p.comparison, lexer.EQ, lexer.NEQ)
}

func (p *parser) comparison() ast.Expr {
	return p.binaryLevel(p.term, lexer.LT, lexer.GT, lexer.LTE, lexer.GTE)
}

func (p *parser) term() ast.Expr {
	return p.binaryLevel(p.factor, lexer.PLUS, lexer.MINUS)
}

func (p *parser) factor() ast.Expr {
	return p.binaryLevel(p.unary, lexer.STAR, lexer.SLASH, lexer.PERCENT)
}

func (p *parser) unary() ast.Expr {
	switch p.peek().Type {
	case lexer.NOT:
		t := p.advance()
		return &ast.Unary{Pos: ast.At(t.Line), Op: "!", X: p.unary()}
	case lexer.MINUS:
		t := p.advance()
		return &ast.Unary{Pos: ast.At(t.Line), Op: "-", X: p.unary()}
	}
	return p.postfix()
}

func (p *parser) postfix() ast.Expr {
	x := p.primary()
	for {
		switch p.peek().Type {
		case lexer.LPAREN:
			t := p.advance()
			var args []ast.Expr
			for !p.at(lexer.RPAREN) {
				args = append(args, p.expression())
				if !p.at(lexer.COMMA) {
					break
				}
				p.advance()
			}
			p.expect(lexer.RPAREN, "')' fi close di function call")
			x = &ast.Call{Pos: ast.At(t.Line), Fn: x, Args: args}
		case lexer.LBRACKET:
			t := p.advance()
			idx := p.expression()
			p.expect(lexer.RBRACKET, "']'")
			x = &ast.Index{Pos: ast.At(t.Line), X: x, Idx: idx}
		default:
			return x
		}
	}
}

func (p *parser) primary() ast.Expr {
	tok := p.peek()
	pos := ast.At(tok.Line)
	switch tok.Type {
	case lexer.NUMBER:
		p.advance()
		v, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			p.fail("Dat a nuh real numba: %s", tok.Literal)
		}
		return &ast.NumberLit{Pos: pos, Value: v}
	case lexer.STRING:
		p.advance()
		return &ast.StringLit{Pos: pos, Value: tok.Literal}
	case lexer.TRUE, lexer.FALSE:
		p.advance()
		return &ast.BoolLit{Pos: pos, Value: tok.Type == lexer.TRUE}
	case lexer.NIL:
		p.advance()
		return &ast.NilLit{Pos: pos}
	case lexer.IDENT:
		p.advance()
		return &ast.Ident{Pos: pos, Name: tok.Literal}
	case lexer.LPAREN:
		p.advance()
		x := p.expression()
		p.expect(lexer.RPAREN, "')'")
		return x
	case lexer.LBRACKET:
		p.advance()
		list := &ast.ListLit{Pos: pos}
		for !p.at(lexer.RBRACKET) {
			list.Elems = append(list.Elems, p.expression())
			if !p.at(lexer.COMMA) {
				break
			}
			p.advance()
		}
		p.expect(lexer.RBRACKET, "']' fi close di list")
		return list
	case lexer.LBRACE:
		p.advance()
		m := &ast.MapLit{Pos: pos}
		p.skipNewlines()
		for !p.at(lexer.RBRACE) {
			var key ast.Expr
			if p.at(lexer.IDENT) && p.peekAt(1).Type == lexer.COLON {
				k := p.advance()
				key = &ast.StringLit{Pos: ast.At(k.Line), Value: k.Literal}
			} else {
				key = p.expression()
			}
			p.expect(lexer.COLON, "':' between di key an di value")
			m.Keys = append(m.Keys, key)
			m.Vals = append(m.Vals, p.expression())
			p.skipNewlines()
			if !p.at(lexer.COMMA) {
				break
			}
			p.advance()
			p.skipNewlines()
		}
		p.skipNewlines()
		p.expect(lexer.RBRACE, "'}' fi close di map")
		return m
	case lexer.FUNC:
		p.advance()
		params := p.params()
		return &ast.FuncLit{Pos: pos, Params: params, Body: p.block()}
	}
	p.fail("Mi nuh understand %s yah so", describe(tok))
	return nil
}

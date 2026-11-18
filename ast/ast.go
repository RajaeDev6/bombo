// Package ast defines the syntax tree for Bombo programs.
package ast

type Node interface{ Line() int }

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}

type Pos struct{ Ln int }

func (p Pos) Line() int { return p.Ln }

func At(line int) Pos { return Pos{Ln: line} }

type Program struct{ Stmts []Stmt }

type Block struct {
	Pos
	Stmts []Stmt
}

type LetStmt struct { // mek x = 1
	Pos
	Name  string
	Value Expr
}

type AssignStmt struct { // x = 1, x += 1, xs[0] = 1
	Pos
	Target Expr
	Op     string // "=", "+=", "-="
	Value  Expr
}

type PrintStmt struct { // seh a, b
	Pos
	Args []Expr
}

type IfStmt struct { // bombo ... bomboclaat ... bloodclaat ...
	Pos
	Cond Expr
	Then *Block
	Else Stmt // *Block, *IfStmt or nil
}

type WhileStmt struct { // gwaan cond { }
	Pos
	Cond Expr
	Body *Block
}

type ForInStmt struct { // fi x inna xs { }
	Pos
	Var  string
	Iter Expr
	Body *Block
}

type ForRangeStmt struct { // fi i from 1 tuh 10 { }
	Pos
	Var      string
	From, To Expr
	Body     *Block
}

type FuncStmt struct { // wuk name(a, b) { }
	Pos
	Name   string
	Params []string
	Body   *Block
}

type ReturnStmt struct { // sen back x
	Pos
	Value Expr
}

type BreakStmt struct{ Pos }    // rassclaat
type ContinueStmt struct{ Pos } // gallang
type ExitStmt struct{ Pos }     // likkle more
type HelloStmt struct{ Pos }    // wah gwaan

type TryStmt struct { // tek a chance { } if it bruk e { }
	Pos
	Body    *Block
	ErrName string
	Catch   *Block
}

type ThrowStmt struct { // bruk "message"
	Pos
	Value Expr
}

type IncStmt struct { // big up x / bun x
	Pos
	Target Expr
	Delta  float64
}

type ImportStmt struct { // tek "game", tek "helpers.bombo" as h
	Pos
	Path string
	Name string // the variable the module is stored in
}

type ExprStmt struct {
	Pos
	X Expr
}

func (*Block) stmt()        {}
func (*LetStmt) stmt()      {}
func (*AssignStmt) stmt()   {}
func (*PrintStmt) stmt()    {}
func (*IfStmt) stmt()       {}
func (*WhileStmt) stmt()    {}
func (*ForInStmt) stmt()    {}
func (*ForRangeStmt) stmt() {}
func (*FuncStmt) stmt()     {}
func (*ReturnStmt) stmt()   {}
func (*BreakStmt) stmt()    {}
func (*ContinueStmt) stmt() {}
func (*ExitStmt) stmt()     {}
func (*HelloStmt) stmt()    {}
func (*TryStmt) stmt()      {}
func (*ThrowStmt) stmt()    {}
func (*IncStmt) stmt()      {}
func (*ImportStmt) stmt()   {}
func (*ExprStmt) stmt()     {}

type NumberLit struct {
	Pos
	Value float64
}

type StringLit struct {
	Pos
	Value string
}

type BoolLit struct {
	Pos
	Value bool
}

type NilLit struct{ Pos }

type Ident struct {
	Pos
	Name string
}

type ListLit struct {
	Pos
	Elems []Expr
}

type MapLit struct {
	Pos
	Keys, Vals []Expr
}

type Unary struct {
	Pos
	Op string // "-" or "!"
	X  Expr
}

type Binary struct {
	Pos
	Op   string
	L, R Expr
}

type Logical struct {
	Pos
	Op   string // "an" or "or"
	L, R Expr
}

type Call struct {
	Pos
	Fn   Expr
	Args []Expr
}

type Index struct {
	Pos
	X, Idx Expr
}

type FuncLit struct {
	Pos
	Params []string
	Body   *Block
}

func (*NumberLit) expr() {}
func (*StringLit) expr() {}
func (*BoolLit) expr()   {}
func (*NilLit) expr()    {}
func (*Ident) expr()     {}
func (*ListLit) expr()   {}
func (*MapLit) expr()    {}
func (*Unary) expr()     {}
func (*Binary) expr()    {}
func (*Logical) expr()   {}
func (*Call) expr()      {}
func (*Index) expr()     {}
func (*FuncLit) expr()   {}

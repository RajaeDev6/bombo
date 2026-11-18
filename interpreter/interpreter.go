// Package interpreter runs Bombo syntax trees.
package interpreter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/RajaeDev6/bombo/ast"
)

const maxDepth = 10000

// BomboError is a runtime error, or a value thrown with `bruk`.
type BomboError struct {
	Line  int
	Msg   string
	Value Value // what `if it bruk e` receives
}

func (e *BomboError) Error() string { return fmt.Sprintf("Line %d: %s", e.Line, e.Msg) }

// ErrExit is returned when the program says `likkle more`.
var ErrExit = errors.New("likkle more")

// control-flow signals travel up the Go call stack as errors
type breakSignal struct{ line int }
type continueSignal struct{ line int }
type returnSignal struct {
	line int
	val  Value
}
type exitSignal struct{}

func (breakSignal) Error() string    { return "rassclaat" }
func (continueSignal) Error() string { return "gallang" }
func (returnSignal) Error() string   { return "sen back" }
func (exitSignal) Error() string     { return "likkle more" }

type Interpreter struct {
	Globals *Env
	Out     io.Writer
	In      *bufio.Reader
	// Dir is the folder that `tek "file.bombo"` paths are relative to.
	Dir string

	builtins *Env
	modules  map[string]Value
	loading  map[string]bool
	depth    int
}

func New(out io.Writer, in *bufio.Reader) *Interpreter {
	it := &Interpreter{
		Out:      out,
		In:       in,
		Dir:      ".",
		builtins: NewEnv(nil),
		modules:  map[string]Value{},
		loading:  map[string]bool{},
	}
	it.Globals = NewEnv(it.builtins)
	it.registerBuiltins()
	return it
}

// Errorf makes a Bombo runtime error for the given line. Libraries written
// in Go use it so their errors look like the rest of Bombo's.
func Errorf(line int, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return &BomboError{Line: line, Msg: msg, Value: msg}
}

func (it *Interpreter) fail(line int, format string, args ...any) error {
	return Errorf(line, format, args...)
}

// Run executes a whole program.
func (it *Interpreter) Run(prog *ast.Program) error {
	_, err := it.RunREPL(prog)
	return err
}

// RunREPL executes a program and returns the value of a trailing expression
// statement, so the REPL can echo it.
func (it *Interpreter) RunREPL(prog *ast.Program) (Value, error) {
	var last Value
	for _, s := range prog.Stmts {
		last = nil
		var err error
		if es, ok := s.(*ast.ExprStmt); ok {
			last, err = it.eval(es.X, it.Globals)
		} else {
			err = it.exec(s, it.Globals)
		}
		if err != nil {
			return nil, it.topLevel(err)
		}
	}
	return last, nil
}

func (it *Interpreter) topLevel(err error) error {
	switch e := err.(type) {
	case exitSignal:
		return ErrExit
	case breakSignal:
		return it.fail(e.line, "'rassclaat' (break) only work inside a loop")
	case continueSignal:
		return it.fail(e.line, "'gallang' (continue) only work inside a loop")
	case returnSignal:
		return it.fail(e.line, "'sen back' only work inside a function")
	}
	return err
}

func (it *Interpreter) execBlock(b *ast.Block, env *Env) error {
	for _, s := range b.Stmts {
		if err := it.exec(s, env); err != nil {
			return err
		}
	}
	return nil
}

// loopSignal decides what a loop does with an error from its body.
// It returns (stop, err).
func loopSignal(err error) (bool, error) {
	switch err.(type) {
	case nil, continueSignal:
		return false, nil
	case breakSignal:
		return true, nil
	}
	return true, err
}

func (it *Interpreter) exec(s ast.Stmt, env *Env) error {
	switch s := s.(type) {
	case *ast.Block:
		return it.execBlock(s, NewEnv(env))

	case *ast.LetStmt:
		v, err := it.eval(s.Value, env)
		if err != nil {
			return err
		}
		if f, ok := v.(*Function); ok && f.Name == "" {
			f.Name = s.Name
		}
		env.Define(s.Name, v)

	case *ast.AssignStmt:
		v, err := it.eval(s.Value, env)
		if err != nil {
			return err
		}
		if s.Op != "=" {
			cur, err := it.eval(s.Target, env)
			if err != nil {
				return err
			}
			if v, err = it.binary(s.Op[:1], cur, v, s.Line()); err != nil {
				return err
			}
		}
		return it.assign(s.Target, v, env, s.Line())

	case *ast.IncStmt:
		cur, err := it.eval(s.Target, env)
		if err != nil {
			return err
		}
		n, ok := cur.(float64)
		if !ok {
			word := "big up"
			if s.Delta < 0 {
				word = "bun"
			}
			return it.fail(s.Line(), "Yuh cyaan '%s' a %s, only a numba", word, TypeName(cur))
		}
		return it.assign(s.Target, n+s.Delta, env, s.Line())

	case *ast.PrintStmt:
		parts := make([]string, len(s.Args))
		for i, a := range s.Args {
			v, err := it.eval(a, env)
			if err != nil {
				return err
			}
			parts[i] = ToString(v)
		}
		fmt.Fprintln(it.Out, strings.Join(parts, " "))

	case *ast.IfStmt:
		c, err := it.eval(s.Cond, env)
		if err != nil {
			return err
		}
		if truthy(c) {
			return it.execBlock(s.Then, NewEnv(env))
		} else if s.Else != nil {
			return it.exec(s.Else, env)
		}

	case *ast.WhileStmt:
		for {
			c, err := it.eval(s.Cond, env)
			if err != nil {
				return err
			}
			if !truthy(c) {
				return nil
			}
			if stop, err := loopSignal(it.execBlock(s.Body, NewEnv(env))); stop {
				return err
			}
		}

	case *ast.ForRangeStmt:
		fromV, err := it.eval(s.From, env)
		if err != nil {
			return err
		}
		toV, err := it.eval(s.To, env)
		if err != nil {
			return err
		}
		from, ok1 := fromV.(float64)
		to, ok2 := toV.(float64)
		if !ok1 || !ok2 {
			return it.fail(s.Line(), "'fi %s from ... tuh ...' need numba dem, not %s an %s", s.Var, TypeName(fromV), TypeName(toV))
		}
		step := 1.0
		if from > to {
			step = -1
		}
		for i := from; (step > 0 && i <= to) || (step < 0 && i >= to); i += step {
			loopEnv := NewEnv(env)
			loopEnv.Define(s.Var, i)
			if stop, err := loopSignal(it.execBlock(s.Body, loopEnv)); stop {
				return err
			}
		}

	case *ast.ForInStmt:
		iter, err := it.eval(s.Iter, env)
		if err != nil {
			return err
		}
		var items []Value
		switch v := iter.(type) {
		case *List:
			items = append(items, v.Elems...)
		case string:
			for _, r := range v {
				items = append(items, string(r))
			}
		case *Map:
			for _, k := range v.Keys {
				items = append(items, k)
			}
		default:
			return it.fail(s.Line(), "Yuh cyaan loop inna a %s. Try 'fi %s from 1 tuh n' instead", TypeName(iter), s.Var)
		}
		for _, item := range items {
			loopEnv := NewEnv(env)
			loopEnv.Define(s.Var, item)
			if stop, err := loopSignal(it.execBlock(s.Body, loopEnv)); stop {
				return err
			}
		}

	case *ast.FuncStmt:
		env.Define(s.Name, &Function{Name: s.Name, Params: s.Params, Body: s.Body, Closure: env})

	case *ast.ReturnStmt:
		var v Value
		if s.Value != nil {
			var err error
			if v, err = it.eval(s.Value, env); err != nil {
				return err
			}
		}
		return returnSignal{line: s.Line(), val: v}

	case *ast.BreakStmt:
		return breakSignal{line: s.Line()}
	case *ast.ContinueStmt:
		return continueSignal{line: s.Line()}
	case *ast.ExitStmt:
		return exitSignal{}
	case *ast.HelloStmt:
	case *ast.TryStmt:
		err := it.execBlock(s.Body, NewEnv(env))
		be, ok := err.(*BomboError)
		if !ok {
			return err // nil, or a control-flow signal
		}
		catchEnv := NewEnv(env)
		if s.ErrName != "" {
			catchEnv.Define(s.ErrName, be.Value)
		}
		return it.execBlock(s.Catch, catchEnv)

	case *ast.ThrowStmt:
		v, err := it.eval(s.Value, env)
		if err != nil {
			return err
		}
		return &BomboError{Line: s.Line(), Msg: ToString(v), Value: v}

	case *ast.ExprStmt:
		_, err := it.eval(s.X, env)
		return err

	case *ast.ImportStmt:
		m, err := it.importModule(s.Path, s.Line())
		if err != nil {
			return err
		}
		env.Define(s.Name, m)

	default:
		return it.fail(s.Line(), "internal: unknown statement %T", s)
	}
	return nil
}

func (it *Interpreter) assign(target ast.Expr, v Value, env *Env, line int) error {
	switch t := target.(type) {
	case *ast.Ident:
		if !env.Assign(t.Name, v) {
			return it.fail(line, "Mi cyaan find '%s'. Use 'mek %s = ...' first", t.Name, t.Name)
		}
		return nil
	case *ast.Index:
		obj, err := it.eval(t.X, env)
		if err != nil {
			return err
		}
		idx, err := it.eval(t.Idx, env)
		if err != nil {
			return err
		}
		switch o := obj.(type) {
		case *List:
			i, err := it.listIndex(len(o.Elems), idx, line)
			if err != nil {
				return err
			}
			o.Elems[i] = v
			return nil
		case *Map:
			k, err := it.mapKey(idx, line)
			if err != nil {
				return err
			}
			o.Set(k, v)
			return nil
		}
		return it.fail(line, "Yuh cyaan put tings inside a %s", TypeName(obj))
	}
	return it.fail(line, "Yuh cyaan assign to dat")
}

func (it *Interpreter) eval(e ast.Expr, env *Env) (Value, error) {
	switch e := e.(type) {
	case *ast.NumberLit:
		return e.Value, nil
	case *ast.StringLit:
		return e.Value, nil
	case *ast.BoolLit:
		return e.Value, nil
	case *ast.NilLit:
		return nil, nil

	case *ast.Ident:
		v, ok := env.Get(e.Name)
		if !ok {
			return nil, it.fail(e.Line(), "Mi cyaan find '%s'. Yuh fi 'mek' it first", e.Name)
		}
		return v, nil

	case *ast.ListLit:
		l := &List{Elems: make([]Value, 0, len(e.Elems))}
		for _, x := range e.Elems {
			v, err := it.eval(x, env)
			if err != nil {
				return nil, err
			}
			l.Elems = append(l.Elems, v)
		}
		return l, nil

	case *ast.MapLit:
		m := NewMap()
		for i := range e.Keys {
			kv, err := it.eval(e.Keys[i], env)
			if err != nil {
				return nil, err
			}
			k, err := it.mapKey(kv, e.Line())
			if err != nil {
				return nil, err
			}
			v, err := it.eval(e.Vals[i], env)
			if err != nil {
				return nil, err
			}
			m.Set(k, v)
		}
		return m, nil

	case *ast.Unary:
		v, err := it.eval(e.X, env)
		if err != nil {
			return nil, err
		}
		if e.Op == "!" {
			return !truthy(v), nil
		}
		n, ok := v.(float64)
		if !ok {
			return nil, it.fail(e.Line(), "Yuh cyaan mek a %s negative", TypeName(v))
		}
		return -n, nil

	case *ast.Logical:
		l, err := it.eval(e.L, env)
		if err != nil {
			return nil, err
		}
		if e.Op == "or" && truthy(l) || e.Op == "an" && !truthy(l) {
			return l, nil
		}
		return it.eval(e.R, env)

	case *ast.Binary:
		l, err := it.eval(e.L, env)
		if err != nil {
			return nil, err
		}
		r, err := it.eval(e.R, env)
		if err != nil {
			return nil, err
		}
		return it.binary(e.Op, l, r, e.Line())

	case *ast.Call:
		fn, err := it.eval(e.Fn, env)
		if err != nil {
			return nil, err
		}
		args := make([]Value, len(e.Args))
		for i, a := range e.Args {
			if args[i], err = it.eval(a, env); err != nil {
				return nil, err
			}
		}
		return it.Call(fn, args, e.Line())

	case *ast.Index:
		obj, err := it.eval(e.X, env)
		if err != nil {
			return nil, err
		}
		idx, err := it.eval(e.Idx, env)
		if err != nil {
			return nil, err
		}
		return it.index(obj, idx, e.Line())

	case *ast.FuncLit:
		return &Function{Params: e.Params, Body: e.Body, Closure: env}, nil
	}
	return nil, it.fail(e.Line(), "internal: unknown expression %T", e)
}

// Call invokes a Bombo function (user-defined or built-in).
func (it *Interpreter) Call(fn Value, args []Value, line int) (Value, error) {
	switch f := fn.(type) {
	case *Builtin:
		return f.Fn(it, args, line)
	case *Function:
		name := f.Name
		if name == "" {
			name = "function"
		}
		if len(args) != len(f.Params) {
			return nil, it.fail(line, "Di function '%s' want %d ting but yuh gi it %d", name, len(f.Params), len(args))
		}
		it.depth++
		defer func() { it.depth-- }()
		if it.depth > maxDepth {
			return nil, it.fail(line, "Stack overflow! Function '%s' a call itself too much, bredda", name)
		}
		env := NewEnv(f.Closure)
		for i, p := range f.Params {
			env.Define(p, args[i])
		}
		err := it.execBlock(f.Body, env)
		switch sig := err.(type) {
		case nil:
			return nil, nil
		case returnSignal:
			return sig.val, nil
		case breakSignal:
			return nil, it.fail(sig.line, "'rassclaat' (break) only work inside a loop")
		case continueSignal:
			return nil, it.fail(sig.line, "'gallang' (continue) only work inside a loop")
		}
		return nil, err
	}
	return nil, it.fail(line, "Dat a %s, not a function. Yuh cyaan call it", TypeName(fn))
}

func (it *Interpreter) binary(op string, l, r Value, line int) (Value, error) {
	switch op {
	case "==":
		return equal(l, r), nil
	case "!=":
		return !equal(l, r), nil
	}

	ln, lok := l.(float64)
	rn, rok := r.(float64)
	if lok && rok {
		switch op {
		case "+":
			return ln + rn, nil
		case "-":
			return ln - rn, nil
		case "*":
			return ln * rn, nil
		case "/":
			if rn == 0 {
				return nil, it.fail(line, "Yuh cyaan divide by zero, bredda! Yuh mad?")
			}
			return ln / rn, nil
		case "%":
			if rn == 0 {
				return nil, it.fail(line, "Yuh cyaan '%%' by zero, bredda")
			}
			return math.Mod(ln, rn), nil
		case "<":
			return ln < rn, nil
		case ">":
			return ln > rn, nil
		case "<=":
			return ln <= rn, nil
		case ">=":
			return ln >= rn, nil
		}
	}

	ls, lsok := l.(string)
	rs, rsok := r.(string)
	switch op {
	case "+":
		if lsok || rsok {
			return ToString(l) + ToString(r), nil
		}
		if ll, ok := l.(*List); ok {
			if rl, ok := r.(*List); ok {
				out := &List{Elems: append(append([]Value{}, ll.Elems...), rl.Elems...)}
				return out, nil
			}
		}
	case "*":
		if lsok && rok && rn >= 0 {
			return strings.Repeat(ls, int(rn)), nil
		}
	case "<", ">", "<=", ">=":
		if lsok && rsok {
			c := strings.Compare(ls, rs)
			switch op {
			case "<":
				return c < 0, nil
			case ">":
				return c > 0, nil
			case "<=":
				return c <= 0, nil
			default:
				return c >= 0, nil
			}
		}
	}
	return nil, it.fail(line, "Yuh cyaan do '%s' wid a %s an a %s, bredda", op, TypeName(l), TypeName(r))
}

func (it *Interpreter) listIndex(n int, idx Value, line int) (int, error) {
	f, ok := idx.(float64)
	if !ok || f != math.Trunc(f) {
		return 0, it.fail(line, "List index fi be a whole numba, not %s", Repr(idx))
	}
	i := int(f)
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		return 0, it.fail(line, "Index %s outta bounds, it only have %d ting inna it", formatNum(f), n)
	}
	return i, nil
}

func (it *Interpreter) mapKey(v Value, line int) (string, error) {
	switch v.(type) {
	case string, float64, bool:
		return ToString(v), nil
	}
	return "", it.fail(line, "A %s cyaan be a map key", TypeName(v))
}

func (it *Interpreter) index(obj, idx Value, line int) (Value, error) {
	switch o := obj.(type) {
	case *List:
		i, err := it.listIndex(len(o.Elems), idx, line)
		if err != nil {
			return nil, err
		}
		return o.Elems[i], nil
	case string:
		runes := []rune(o)
		i, err := it.listIndex(len(runes), idx, line)
		if err != nil {
			return nil, err
		}
		return string(runes[i]), nil
	case *Map:
		k, err := it.mapKey(idx, line)
		if err != nil {
			return nil, err
		}
		return o.Vals[k], nil // missing key gives kakafaat
	case Object:
		name, ok := idx.(string)
		if !ok {
			return nil, it.fail(line, "Look inside a %s wid a name, like x.width", o.TypeName())
		}
		return o.Field(name), nil
	}
	return nil, it.fail(line, "Yuh cyaan look inside a %s wid [ ]", TypeName(obj))
}

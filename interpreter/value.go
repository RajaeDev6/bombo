package interpreter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/RajaeDev6/bombo/ast"
)

// Value is any Bombo value: nil (kakafaat), bool, float64, string,
// *List, *Map, *Function or *Builtin.
type Value any

type List struct{ Elems []Value }

type Map struct {
	Keys []string
	Vals map[string]Value
}

func NewMap() *Map { return &Map{Vals: map[string]Value{}} }

func (m *Map) Set(k string, v Value) {
	if _, ok := m.Vals[k]; !ok {
		m.Keys = append(m.Keys, k)
	}
	m.Vals[k] = v
}

func (m *Map) Delete(k string) {
	if _, ok := m.Vals[k]; !ok {
		return
	}
	delete(m.Vals, k)
	for i, key := range m.Keys {
		if key == k {
			m.Keys = append(m.Keys[:i], m.Keys[i+1:]...)
			break
		}
	}
}

type Function struct {
	Name    string
	Params  []string
	Body    *ast.Block
	Closure *Env
}

type builtinFn func(it *Interpreter, args []Value, line int) (Value, error)

type Builtin struct {
	Name string
	Fn   builtinFn
}

type Env struct {
	vars   map[string]Value
	parent *Env
}

func NewEnv(parent *Env) *Env { return &Env{vars: map[string]Value{}, parent: parent} }

func (e *Env) Get(name string) (Value, bool) {
	for env := e; env != nil; env = env.parent {
		if v, ok := env.vars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func (e *Env) Define(name string, v Value) { e.vars[name] = v }

func (e *Env) Assign(name string, v Value) bool {
	for env := e; env != nil; env = env.parent {
		if _, ok := env.vars[name]; ok {
			env.vars[name] = v
			return true
		}
	}
	return false
}

func formatNum(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// ToString is how `seh` shows a value.
func ToString(v Value) string {
	switch v := v.(type) {
	case nil:
		return "kakafaat"
	case bool:
		if v {
			return "yeh mon"
		}
		return "nah mon"
	case float64:
		return formatNum(v)
	case string:
		return v
	case *List:
		parts := make([]string, len(v.Elems))
		for i, e := range v.Elems {
			parts[i] = Repr(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *Map:
		parts := make([]string, len(v.Keys))
		for i, k := range v.Keys {
			parts[i] = strconv.Quote(k) + ": " + Repr(v.Vals[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *Function:
		if v.Name == "" {
			return "<function>"
		}
		return "<function " + v.Name + ">"
	case *Builtin:
		return "<built-in function " + v.Name + ">"
	}
	return fmt.Sprintf("%v", v)
}

// Repr is like ToString but quotes strings (used inside lists, maps and the REPL).
func Repr(v Value) string {
	if s, ok := v.(string); ok {
		return strconv.Quote(s)
	}
	return ToString(v)
}

func TypeName(v Value) string {
	switch v.(type) {
	case nil:
		return "kakafaat"
	case bool:
		return "truth"
	case float64:
		return "numba"
	case string:
		return "wud"
	case *List:
		return "list"
	case *Map:
		return "map"
	case *Function, *Builtin:
		return "function"
	}
	return "unknown"
}

func truthy(v Value) bool {
	switch v := v.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	}
	return true
}

func equal(a, b Value) bool {
	switch a := a.(type) {
	case nil:
		return b == nil
	case bool:
		bb, ok := b.(bool)
		return ok && a == bb
	case float64:
		bb, ok := b.(float64)
		return ok && a == bb
	case string:
		bb, ok := b.(string)
		return ok && a == bb
	case *List:
		bb, ok := b.(*List)
		if !ok || len(a.Elems) != len(bb.Elems) {
			return false
		}
		for i := range a.Elems {
			if !equal(a.Elems[i], bb.Elems[i]) {
				return false
			}
		}
		return true
	case *Map:
		bb, ok := b.(*Map)
		if !ok || len(a.Keys) != len(bb.Keys) {
			return false
		}
		for k, v := range a.Vals {
			if bv, ok := bb.Vals[k]; !ok || !equal(v, bv) {
				return false
			}
		}
		return true
	}
	return a == b
}

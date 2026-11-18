package interpreter

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (it *Interpreter) registerBuiltins() {
	add := func(fn builtinFn, names ...string) {
		b := &Builtin{Name: names[0], Fn: fn}
		for _, n := range names {
			it.Globals.Define(n, b)
		}
	}
	add(bLen, "howmuch", "len")
	add(bPush, "pon", "push")
	add(bPop, "tekoff", "pop")
	add(bInput, "ax", "input")
	add(bType, "wahkind", "type")
	add(bNum, "numba", "num")
	add(bStr, "wud", "str")
	add(bRandom, "luck", "random")
	add(bKeys, "keys")
	add(bHas, "has")
	add(bRemove, "dash", "remove")
	add(bSort, "sort")
	add(bSplit, "split")
	add(bJoin, "join")
	add(bUpper, "upper")
	add(bLower, "lower")
	add(bSleep, "cool", "sleep")
	add(bTime, "time")
	add(mathFn("floor", math.Floor), "floor")
	add(mathFn("ceil", math.Ceil), "ceil")
	add(mathFn("round", math.Round), "round")
	add(mathFn("abs", math.Abs), "abs")
	add(mathFn("sqrt", math.Sqrt), "sqrt")
}

func (it *Interpreter) arity(name string, args []Value, n, line int) error {
	if len(args) != n {
		return it.fail(line, "'%s' want %d ting but yuh gi it %d", name, n, len(args))
	}
	return nil
}

func (it *Interpreter) wrongType(name, want string, got Value, line int) error {
	return it.fail(line, "'%s' want a %s, but yuh gi it a %s", name, want, TypeName(got))
}

func bLen(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("howmuch", args, 1, line); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case string:
		return float64(len([]rune(v))), nil
	case *List:
		return float64(len(v.Elems)), nil
	case *Map:
		return float64(len(v.Keys)), nil
	}
	return nil, it.wrongType("howmuch", "wud, list or map", args[0], line)
}

func bPush(it *Interpreter, args []Value, line int) (Value, error) {
	if len(args) < 2 {
		return nil, it.fail(line, "'pon' want a list an at least one ting fi put pon it")
	}
	l, ok := args[0].(*List)
	if !ok {
		return nil, it.wrongType("pon", "list", args[0], line)
	}
	l.Elems = append(l.Elems, args[1:]...)
	return l, nil
}

func bPop(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("tekoff", args, 1, line); err != nil {
		return nil, err
	}
	l, ok := args[0].(*List)
	if !ok {
		return nil, it.wrongType("tekoff", "list", args[0], line)
	}
	if len(l.Elems) == 0 {
		return nil, it.fail(line, "Cyaan 'tekoff' from an empty list, nutten nuh deh deh")
	}
	v := l.Elems[len(l.Elems)-1]
	l.Elems = l.Elems[:len(l.Elems)-1]
	return v, nil
}

func bInput(it *Interpreter, args []Value, line int) (Value, error) {
	if len(args) > 1 {
		return nil, it.fail(line, "'ax' want one question at most")
	}
	if len(args) == 1 {
		fmt.Fprint(it.Out, ToString(args[0]))
	}
	s, err := it.In.ReadString('\n')
	if err != nil && (err != io.EOF || s == "") {
		return nil, nil
	}
	return strings.TrimRight(s, "\r\n"), nil
}

func bType(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("wahkind", args, 1, line); err != nil {
		return nil, err
	}
	return TypeName(args[0]), nil
}

func bNum(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("numba", args, 1, line); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case float64:
		return v, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, nil // not a number → kakafaat
		}
		return f, nil
	}
	return nil, nil
}

func bStr(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("wud", args, 1, line); err != nil {
		return nil, err
	}
	return ToString(args[0]), nil
}

func bRandom(it *Interpreter, args []Value, line int) (Value, error) {
	if len(args) == 0 {
		return rand.Float64(), nil
	}
	if err := it.arity("luck", args, 2, line); err != nil {
		return nil, err
	}
	lo, ok1 := args[0].(float64)
	hi, ok2 := args[1].(float64)
	if !ok1 || !ok2 || hi < lo {
		return nil, it.fail(line, "'luck' want two numba dem, low an high (like luck(1, 10))")
	}
	a, b := int(lo), int(hi)
	return float64(a + rand.IntN(b-a+1)), nil
}

func bKeys(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("keys", args, 1, line); err != nil {
		return nil, err
	}
	m, ok := args[0].(*Map)
	if !ok {
		return nil, it.wrongType("keys", "map", args[0], line)
	}
	l := &List{}
	for _, k := range m.Keys {
		l.Elems = append(l.Elems, k)
	}
	return l, nil
}

func bHas(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("has", args, 2, line); err != nil {
		return nil, err
	}
	switch c := args[0].(type) {
	case *Map:
		k, err := it.mapKey(args[1], line)
		if err != nil {
			return nil, err
		}
		_, ok := c.Vals[k]
		return ok, nil
	case *List:
		for _, e := range c.Elems {
			if equal(e, args[1]) {
				return true, nil
			}
		}
		return false, nil
	case string:
		return strings.Contains(c, ToString(args[1])), nil
	}
	return nil, it.wrongType("has", "map, list or wud", args[0], line)
}

func bRemove(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("dash", args, 2, line); err != nil {
		return nil, err
	}
	switch c := args[0].(type) {
	case *Map:
		k, err := it.mapKey(args[1], line)
		if err != nil {
			return nil, err
		}
		v := c.Vals[k]
		c.Delete(k)
		return v, nil
	case *List:
		i, err := it.listIndex(len(c.Elems), args[1], line)
		if err != nil {
			return nil, err
		}
		v := c.Elems[i]
		c.Elems = append(c.Elems[:i], c.Elems[i+1:]...)
		return v, nil
	}
	return nil, it.wrongType("dash", "map or list", args[0], line)
}

func bSort(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("sort", args, 1, line); err != nil {
		return nil, err
	}
	l, ok := args[0].(*List)
	if !ok {
		return nil, it.wrongType("sort", "list", args[0], line)
	}
	out := &List{Elems: append([]Value{}, l.Elems...)}
	var sortErr error
	sort.SliceStable(out.Elems, func(i, j int) bool {
		r, err := it.binary("<", out.Elems[i], out.Elems[j], line)
		if err != nil {
			sortErr = err
			return false
		}
		return r.(bool)
	})
	if sortErr != nil {
		return nil, sortErr
	}
	return out, nil
}

func bSplit(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("split", args, 2, line); err != nil {
		return nil, err
	}
	s, ok1 := args[0].(string)
	sep, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, it.fail(line, "'split' want two wud dem: split(\"a,b\", \",\")")
	}
	l := &List{}
	for _, p := range strings.Split(s, sep) {
		l.Elems = append(l.Elems, p)
	}
	return l, nil
}

func bJoin(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("join", args, 2, line); err != nil {
		return nil, err
	}
	l, ok := args[0].(*List)
	if !ok {
		return nil, it.wrongType("join", "list", args[0], line)
	}
	parts := make([]string, len(l.Elems))
	for i, e := range l.Elems {
		parts[i] = ToString(e)
	}
	return strings.Join(parts, ToString(args[1])), nil
}

func bUpper(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("upper", args, 1, line); err != nil {
		return nil, err
	}
	return strings.ToUpper(ToString(args[0])), nil
}

func bLower(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("lower", args, 1, line); err != nil {
		return nil, err
	}
	return strings.ToLower(ToString(args[0])), nil
}

func bSleep(it *Interpreter, args []Value, line int) (Value, error) {
	if err := it.arity("cool", args, 1, line); err != nil {
		return nil, err
	}
	ms, ok := args[0].(float64)
	if !ok {
		return nil, it.wrongType("cool", "numba (milliseconds)", args[0], line)
	}
	time.Sleep(time.Duration(ms * float64(time.Millisecond)))
	return nil, nil
}

func bTime(it *Interpreter, args []Value, line int) (Value, error) {
	return float64(time.Now().UnixNano()) / 1e9, nil
}

func mathFn(name string, f func(float64) float64) builtinFn {
	return func(it *Interpreter, args []Value, line int) (Value, error) {
		if err := it.arity(name, args, 1, line); err != nil {
			return nil, err
		}
		n, ok := args[0].(float64)
		if !ok {
			return nil, it.wrongType(name, "numba", args[0], line)
		}
		return f(n), nil
	}
}

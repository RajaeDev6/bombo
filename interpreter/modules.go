package interpreter

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/RajaeDev6/bombo/ast"
	"github.com/RajaeDev6/bombo/parser"
)

// ModuleLoader builds a library module written in Go. It returns the map
// that `tek "name"` stores, usually full of *Builtin functions.
type ModuleLoader func(it *Interpreter) (*Map, error)

var libraries = map[string]ModuleLoader{}

// RegisterModule makes a Go library available to Bombo code as
// `tek "name"`. Call it from the library package's init function.
func RegisterModule(name string, load ModuleLoader) {
	libraries[name] = load
}

// importModule runs `tek path`. A registered library name wins; anything
// else is a .bombo file relative to the importing file.
func (it *Interpreter) importModule(path string, line int) (Value, error) {
	if load, ok := libraries[path]; ok {
		key := "lib:" + path
		if m, ok := it.modules[key]; ok {
			return m, nil
		}
		m, err := load(it)
		if err != nil {
			return nil, err
		}
		it.modules[key] = m
		return m, nil
	}
	return it.importFile(path, line)
}

func (it *Interpreter) importFile(path string, line int) (Value, error) {
	file := path
	if !strings.HasSuffix(file, ".bombo") {
		file += ".bombo"
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(it.Dir, file)
	}
	abs, err := filepath.Abs(file)
	if err != nil {
		abs = file
	}
	if m, ok := it.modules[abs]; ok {
		return m, nil
	}
	if it.loading[abs] {
		return nil, it.fail(line, "'%s' a 'tek' itself in a circle. Two files cyaan tek each other", path)
	}

	src, err := os.ReadFile(abs)
	if err != nil {
		if strings.ContainsAny(path, "/.") {
			return nil, it.fail(line, "Mi cyaan find di file '%s'", file)
		}
		return nil, it.fail(line, "Mi cyaan find a module name '%s'. Built-in modules: %s. Or mek a file name '%s.bombo'", path, strings.Join(Modules(), ", "), path)
	}
	prog, err := parser.Parse(string(src))
	if err != nil {
		return nil, it.fail(line, "Inna '%s': %s", path, err)
	}

	it.loading[abs] = true
	oldDir := it.Dir
	it.Dir = filepath.Dir(abs)
	env := NewEnv(it.builtins)
	err = it.runIn(prog, env)
	it.Dir = oldDir
	delete(it.loading, abs)
	if err != nil {
		if be, ok := err.(*BomboError); ok {
			return nil, &BomboError{Line: line, Msg: "Inna '" + path + "' line " + itoa(be.Line) + ": " + be.Msg, Value: be.Value}
		}
		return nil, err
	}

	names := make([]string, 0, len(env.vars))
	for n := range env.vars {
		names = append(names, n)
	}
	sort.Strings(names)
	m := NewMap()
	for _, n := range names {
		m.Set(n, env.vars[n])
	}
	it.modules[abs] = m
	return m, nil
}

// Modules lists the Go libraries that can be loaded with `tek`.
func Modules() []string {
	names := make([]string, 0, len(libraries))
	for n := range libraries {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// runIn runs a whole program in env, the way the top level of a file runs.
func (it *Interpreter) runIn(prog *ast.Program, env *Env) error {
	for _, s := range prog.Stmts {
		if err := it.exec(s, env); err != nil {
			return it.topLevel(err)
		}
	}
	return nil
}

func itoa(n int) string { return formatNum(float64(n)) }

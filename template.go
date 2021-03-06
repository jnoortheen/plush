package plush

import (
	"sync"

	"github.com/gobuffalo/plush/ast"

	"github.com/gobuffalo/plush/parser"

	"github.com/pkg/errors"
)

// Template represents an input and helpers to be used
// to evaluate and render the input.
type Template struct {
	Input   string
	Helpers HelperMap
	program *ast.Program
	moot    *sync.Mutex
}

// NewTemplate from the input string. Adds all of the
// global helper functions from "Helpers", this function does not
// cache the template.
func NewTemplate(input string) (*Template, error) {
	hm, err := NewHelperMap()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	t := &Template{
		Input:   input,
		Helpers: hm,
		moot:    &sync.Mutex{},
	}
	err = t.Parse()
	if err != nil {
		return t, errors.WithStack(err)
	}
	return t, nil
}

// Parse the template this can be called many times
// as a successful result is cached and is used on subsequent
// uses.
func (t *Template) Parse() error {
	if t.program != nil {
		return nil
	}
	program, err := parser.Parse(t.Input)
	if err != nil {
		return errors.WithStack(err)
	}
	t.program = program
	return nil
}

// Exec the template using the content and return the results
func (t *Template) Exec(ctx *Context) (string, error) {
	t.moot.Lock()
	defer t.moot.Unlock()
	err := t.Parse()
	if err != nil {
		return "", err
	}

	ctx = ctx.New()
	for k, v := range t.Helpers.Helpers() {
		ctx.Set(k, v)
	}
	ev := compiler{
		ctx:     ctx,
		program: t.program,
	}

	s, err := ev.compile()
	return s, err
}

// Clone a template. This is useful for defining helpers on per "instance" of the template.
func (t *Template) Clone() *Template {
	hm, _ := NewHelperMap()
	hm.AddMany(t.Helpers.Helpers())
	t2 := &Template{
		Helpers: hm,
		Input:   t.Input,
		program: t.program,
	}
	return t2
}

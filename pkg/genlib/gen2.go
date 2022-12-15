// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"strings"
	"text/template"
)

// Gen2 is resolved at construction to a slice of emit functions
type Gen2 struct {
	emitFuncs []emitF
}

// NewGen2
//
// From current benchmarks this generator is at least as CPU performant as Generator, while providing templating.
// Is worse in terms of RAM used and memory allocations (but working on template parse optimization may help).
//
// If you're getting a nil pointer dereference like error, something like:
// template: generator:2:2: executing "generator" at <generate "foobar">: error calling generate: runtime error: invalid memory address or nil pointer dereference
// is probably due to the "foobar" not being a valid field. In this case the generate function will try accessing fieldMap at an invalid location.
func NewGen2(tpl []byte, cfg Config, fields Fields) (*Gen2, error) {

	// extracts objects keys
	// FIXME: this logic works for field.* but what about field.*.*.* (like in gcp package)?
	objectKeys := make(map[string]struct{})
	objectKeysFields := make(map[string]Field)
	for _, field := range fields {
		if strings.HasSuffix(field.Name, ".*") {
			objectKeys[field.Name] = struct{}{}
			objectKeysFields[field.Name] = field
		}
	}

	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]emitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, objectKeys, nil); err != nil {
			return nil, err
		}
	}

	// Preprocess the object keys, generating appropriate emit functions
	// TODO: is this necessary? Works without and is not clear to me what is the benefit
	for k := range objectKeysFields {
		field := objectKeysFields[k]
		if err := bindField(cfg, field, fieldMap, objectKeys, nil); err != nil {
			return nil, err
		}
	}

	t := template.New("generator")
	t = t.Option("missingkey=error")

	f := func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		templateFns := template.FuncMap{}

		templateFns["generate"] = func(field string) string {
			// FIXME: this writes directly to buf, while template expects a return value
			_ = fieldMap[field](state, dupes, buf)
			// NOTE: mandatory empty string return to avoid cluttering the template, otherwise
			// you get a "<no value>" for each field written to buf
			return ""
		}

		// NOTE: Funcs needs to be run before parsing. As we want to bind state and dups this needs
		// to happen here, parsing the template for each emit.
		// This is not needed and if we can get rid of state (which I've not seen used in value
		// generation functions) Funcs and Parse may happen only once; I think this may improve
		// performances.
		parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
		if err != nil {
			return err
		}
		err = parsedTpl.Execute(buf, nil)
		if err != nil {
			return err
		}

		return nil
	}

	// NOTE: as this generator has only 1 emitF is useless to store them in a slice and iterate
	// on them n Emit. The code from `f` could be directly written in Emit.
	return &Gen2{emitFuncs: []emitF{f}}, nil
}

func (e Gen2) Emit(state *GenState, buf *bytes.Buffer) error {
	dupes := make(map[string]struct{})

	for _, f := range e.emitFuncs {
		if err := f(state, dupes, buf); err != nil {
			return err
		}
	}
	return nil
}

// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"text/template"
)

// Gen2 is resolved at construction to a slice of emit functions
type Gen2 struct {
	emitFuncs []emitF
}

func NewGen2(tpl []byte, cfg Config, fields Fields) (*Gen2, error) {

	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]emitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, nil, nil); err != nil {
			return nil, err
		}
	}

	t := template.New("generator")

	f := func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		templateFns := template.FuncMap{}

		templateFns["generate"] = func(field string) any {
			return fieldMap[field](state, dupes, buf)
		}

		// for name, field := range fieldMap {
		// 	fnname := strings.Replace(name, ".", "_", -1) // . are not allowed in Fn names
		// 	fnname = strings.Replace(fnname, "@", "", -1)

		// 	templateFns[fnname] = func() any {
		// 		return field(state, dupes, buf)
		// 	}
		// }

		parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
		if err != nil {
			panic(err)
		}
		err = parsedTpl.Execute(buf, nil)
		if err != nil {
			return err
		}

		return nil
	}

	return &Gen2{emitFuncs: []emitF{f}}, nil
}

func (gen Gen2) Emit(state *GenState, buf *bytes.Buffer) error {

	buf.WriteByte('{')

	if err := gen.emit(state, buf); err != nil {
		return err
	}

	buf.WriteByte('}')

	state.counter += 1

	return nil
}

func (gen Gen2) emit(state *GenState, buf *bytes.Buffer) error {

	dupes := make(map[string]struct{})

	lastComma := -1
	for _, f := range gen.emitFuncs {
		pos := buf.Len()
		if err := f(state, dupes, buf); err != nil {
			return err
		}

		// If we emitted something, write the comma, otherwise skip.
		if buf.Len() > pos {
			buf.WriteByte(',')
			lastComma = buf.Len()
		}
	}

	// Strip dangling comma
	if lastComma == buf.Len() {
		buf.Truncate(buf.Len() - 1)
	}

	return nil
}

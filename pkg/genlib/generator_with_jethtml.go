// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"github.com/CloudyKit/jet/v6"
	"reflect"
)

// GeneratorWithJetHTML
type GeneratorWithJetHTML struct {
	jetHTML *jet.Template
	state   *GenState
}

func NewGeneratorWithJetHTML(tpl []byte, cfg Config, fields Fields) (*GeneratorWithJetHTML, error) {
	if len(tpl) == 0 {
		tpl = []byte("")
	}
	loader := jet.NewInMemLoader()
	loader.Set("template", string(tpl))
	view := jet.NewSet(loader)
	t, err := view.GetTemplate("template")
	if err != nil {
		return nil, err
	}

	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]EmitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, nil, nil, true); err != nil {
			return nil, err
		}
	}

	state := NewGenState()

	view.AddGlobalFunc("generate", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("generate", 1, 1)
		arg := a.Get(0)
		field := arg.String()

		bindF, ok := fieldMap[field]
		if !ok {
			return reflect.ValueOf(nil)
		}

		value, err := bindF(state, nil)
		if err != nil {
			return reflect.ValueOf(nil)
		}
		return reflect.ValueOf(value)
	})

	if err != nil {
		return nil, err
	}

	return &GeneratorWithJetHTML{jetHTML: t, state: state}, nil
}

func (GeneratorWithJetHTML) Close() error {
	return nil
}

func (gen GeneratorWithJetHTML) Emit(state *GenState, buf *bytes.Buffer) error {
	state = gen.state
	if err := gen.emit(state, buf); err != nil {
		return err
	}

	state.counter += 1

	return nil
}

func (gen GeneratorWithJetHTML) emit(state *GenState, buf *bytes.Buffer) error {
	vars := make(jet.VarMap)
	err := gen.jetHTML.Execute(buf, vars, nil)
	if err != nil {
		return err
	}

	return nil
}

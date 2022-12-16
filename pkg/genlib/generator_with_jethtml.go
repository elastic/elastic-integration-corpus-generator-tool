// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"github.com/CloudyKit/jet"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
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
	tmpDir, err := os.MkdirTemp("", "jethtml-*")
	if err != nil {
		return nil, err
	}

	randomFilename := path.Base(tmpDir)
	fullPath := path.Join(tmpDir, randomFilename)
	err = ioutil.WriteFile(fullPath, tpl, 0660)
	if err != nil {
		return nil, err
	}

	var view = jet.NewHTMLSet(tmpDir)
	t, err := view.GetTemplate(randomFilename)
	if err != nil {
		return nil, err
	}

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
		if err := bindField(cfg, field, fieldMap, objectKeys); err != nil {
			return nil, err
		}
	}

	// Preprocess the object keys, generating appropriate emit functions
	// TODO: is this necessary? Works without and is not clear to me what is the benefit
	for k := range objectKeysFields {
		field := objectKeysFields[k]
		if err := bindField(cfg, field, fieldMap, objectKeys); err != nil {
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

		value, err := bindF(state)
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

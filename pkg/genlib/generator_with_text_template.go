// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"strings"
	"text/template"
)

// GeneratorWithTemplate
type GeneratorWithTemplate struct {
	tpl   *template.Template
	state *GenState
}

func NewGeneratorWithTemplate(tpl []byte, cfg Config, fields Fields) (*GeneratorWithTemplate, error) {
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

	t := template.New("generator")
	t = t.Option("missingkey=error")

	state := NewGenState()

	templateFns := template.FuncMap{}
	templateFns["generate"] = func(field string) interface{} {
		bindF, ok := fieldMap[field]
		if !ok {
			return nil
		}

		value, err := bindF(state)
		if err != nil {
			return nil
		}

		return value
	}
	parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &GeneratorWithTemplate{tpl: parsedTpl, state: state}, nil
}

func (gen GeneratorWithTemplate) Emit(state *GenState, buf *bytes.Buffer) error {
	state = gen.state
	if err := gen.emit(state, buf); err != nil {
		return err
	}

	state.counter += 1

	return nil
}

func (gen GeneratorWithTemplate) emit(state *GenState, buf *bytes.Buffer) error {
	err := gen.tpl.Execute(buf, nil)
	if err != nil {
		return err
	}

	return nil
}

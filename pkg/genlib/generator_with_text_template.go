// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"text/template"
	"time"
)

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	tpl   *template.Template
	state *GenState
}

func NewGeneratorWithTextTemplate(tpl []byte, cfg Config, fields Fields) (*GeneratorWithTextTemplate, error) {
	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]EmitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, nil, nil, true); err != nil {
			return nil, err
		}
	}

	t := template.New("generator")
	t = t.Option("missingkey=error")

	state := NewGenState()

	templateFns := sprig.HermeticTxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) interface{} {
		bindF, ok := fieldMap[field]
		if !ok {
			panic(fmt.Errorf("missing field: '%s' (is it present in fields.yml?)", field))
		}

		value, err := bindF(state, nil)
		if err != nil {
			return ""
		}
		return value
	}

	parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &GeneratorWithTextTemplate{tpl: parsedTpl, state: state}, nil
}

func (GeneratorWithTextTemplate) Close() error {
	return nil
}

func (gen GeneratorWithTextTemplate) Emit(state *GenState, buf *bytes.Buffer) error {
	state = gen.state
	if err := gen.emit(state, buf); err != nil {
		return err
	}

	state.counter += 1

	return nil
}

func (gen GeneratorWithTextTemplate) emit(state *GenState, buf *bytes.Buffer) error {
	err := gen.tpl.Execute(buf, nil)
	if err != nil {
		return err
	}

	return nil
}

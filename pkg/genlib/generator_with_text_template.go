// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"io"
	"text/template"
	"time"
)

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	tpl       *template.Template
	state     *GenState
	totEvents uint64
}

func NewGeneratorWithTextTemplate(tpl []byte, cfg Config, fields Fields, totSize uint64) (*GeneratorWithTextTemplate, error) {
	// Preprocess the fields, generating appropriate bound function
	state := NewGenState()
	fieldMap := make(map[string]any)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, true); err != nil {
			return nil, err
		}

		state.prevCacheForDup[field.Name] = make(map[any]struct{})
		state.prevCacheCardinality[field.Name] = make([]any, 0)
	}

	// Generate a single event to calculate the total number of events based on its size
	t := template.New("estimate_tot_events")
	t = t.Option("missingkey=error")

	templateFns := sprig.HermeticTxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) any {
		state := NewGenState()
		state.prevCacheForDup[field] = make(map[any]struct{})
		state.prevCacheCardinality[field] = make([]any, 0)
		bindF, ok := fieldMap[field].(EmitF)
		if !ok {
			return ""
		}

		return bindF(state)
	}

	parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString("")
	err = parsedTpl.Execute(buf, nil)
	if err != nil {
		return nil, err
	}

	var totEvents uint64
	singleEventSize := uint64(buf.Len())
	if singleEventSize == 0 {
		totEvents = 1
	} else {
		totEvents = totSize / singleEventSize
		if totEvents < 1 {
			totEvents = 1
		}
	}

	t = template.New("generator")
	t = t.Option("missingkey=error")

	templateFns = sprig.HermeticTxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) any {
		bindF, ok := fieldMap[field].(EmitF)
		if !ok {
			return ""
		}

		return bindF(state)
	}

	parsedTpl, err = t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &GeneratorWithTextTemplate{tpl: parsedTpl, totEvents: totEvents, state: state}, nil
}

func (gen GeneratorWithTextTemplate) Close() error {
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
	if state.counter < gen.totEvents {
		err := gen.tpl.Execute(buf, nil)
		if err != nil {
			return err
		}
	} else {
		return io.EOF
	}

	return nil
}

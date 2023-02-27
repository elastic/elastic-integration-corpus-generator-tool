// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"errors"
	"github.com/Masterminds/sprig/v3"
	"io"
	"text/template"
	"time"
)

var generateOnFieldNotInFieldsYaml = errors.New("generate called on a field not present in fields yaml definition")

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	tpl       *template.Template
	state     *GenState
	errChan   chan error
	totEvents uint64
}

func calculateTotEventsWithTextTemplate(totSize uint64, fieldMap map[string]any, errChan chan error, tpl []byte) (uint64, error) {
	if totSize == 0 {
		return 0, nil
	}

	// Generate a single event to calculate the total number of events based on its size
	t := template.New("estimate_tot_events")
	t = t.Option("missingkey=error")

	templateFns := sprig.TxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) any {
		state := NewGenState()
		state.prevCacheForDup[field] = make(map[any]struct{})
		state.prevCacheCardinality[field] = make([]any, 0)
		bindF, ok := fieldMap[field].(EmitF)
		if !ok {
			close(errChan)
			return nil
		}

		return bindF(state)
	}

generateErr:
	for {
		select {
		case <-errChan:
			return 0, generateOnFieldNotInFieldsYaml
		default:
			break generateErr
		}
	}

	parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBufferString("")
	err = parsedTpl.Execute(buf, nil)
	if err != nil {
		return 0, err
	}

	singleEventSize := uint64(buf.Len())
	if singleEventSize == 0 {
		return 1, nil
	}

	totEvents := totSize / singleEventSize
	if totEvents < 1 {
		totEvents = 1
	}

	return totEvents, nil
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

	errChan := make(chan error)

	totEvents, err := calculateTotEventsWithTextTemplate(totSize, fieldMap, errChan, tpl)
	if err != nil {
		return nil, err
	}

	t := template.New("generator")
	t = t.Option("missingkey=error")

	templateFns := sprig.TxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) any {
		bindF, ok := fieldMap[field].(EmitF)
		if !ok {
			close(errChan)
			return nil
		}

		return bindF(state)
	}

	parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &GeneratorWithTextTemplate{tpl: parsedTpl, totEvents: totEvents, state: state, errChan: errChan}, nil
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
	if gen.totEvents == 0 || state.counter < gen.totEvents {
		select {
		case <-gen.errChan:
			return generateOnFieldNotInFieldsYaml
		default:
			err := gen.tpl.Execute(buf, nil)
			if err != nil {
				return err
			}
		}
	} else {
		return io.EOF
	}

	return nil
}

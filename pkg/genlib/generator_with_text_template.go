// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"runtime"
	"text/template"
	"time"
)

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	closed  chan struct{}
	bindMap map[string]chan interface{}
	tpl     *template.Template
	state   *GenState
}

func NewGeneratorWithTextTemplate(tpl []byte, cfg Config, fields Fields) (*GeneratorWithTextTemplate, error) {
	state := NewGenState()

	// Preprocess the fields, generating appropriate emit channels
	closedChan := make(chan struct{})
	fieldMap := make(map[string]EmitF)
	bindMap := make(map[string]chan interface{})
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap); err != nil {
			return nil, err
		}

		chanSize := runtime.GOMAXPROCS(0) / 2
		if chanSize < 1 {
			chanSize = 1
		}

		bindChan := make(chan interface{}, chanSize)
		bindMap[field.Name] = bindChan
		go func(bindChan chan interface{}, closedChan chan struct{}, bindF EmitF) {
			for {
				select {
				case <-closedChan:
					return
				default:
					value, err := bindF(state, nil)
					if err != nil {
						bindChan <- ""
						continue
					}
					bindChan <- value
				}
			}
		}(bindChan, closedChan, fieldMap[field.Name])
	}

	t := template.New("generator")
	t = t.Option("missingkey=error")

	templateFns := sprig.HermeticTxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) interface{} {
		bindChan, ok := bindMap[field]
		if !ok {
			return ""
		}

		return <-bindChan
	}

	parsedTpl, err := t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &GeneratorWithTextTemplate{tpl: parsedTpl, state: state, bindMap: bindMap, closed: closedChan}, nil
}

func (gen GeneratorWithTextTemplate) Close() error {
	close(gen.closed)

	return nil
}

func (gen GeneratorWithTextTemplate) Emit(state *GenState, buf *bytes.Buffer) error {
	state = gen.state
	if err := gen.emit(buf); err != nil {
		return err
	}

	return nil
}

func (gen GeneratorWithTextTemplate) emit(buf *bytes.Buffer) error {
	err := gen.tpl.Execute(buf, nil)
	if err != nil {
		return err
	}

	return nil
}

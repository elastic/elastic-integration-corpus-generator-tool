// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"runtime"
	"github.com/Masterminds/sprig/v3"
	"text/template"
	"time"
)

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	closed  chan struct{}
	bindMap map[string]chan any
	tpl     *template.Template
}

func NewGeneratorWithTextTemplate(tpl []byte, cfg Config, fields Fields) (*GeneratorWithTextTemplate, error) {
	// Preprocess the fields, generating appropriate emit channels
	chanSize := runtime.GOMAXPROCS(0) / 2
	if chanSize < 1 {
		chanSize = 1
	}

	closedChan := make(chan struct{})
	fieldMap := make(map[string]any)
	bindMap := make(map[string]chan any)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, true); err != nil {
			return nil, err
		}


		bindChan := make(chan any)
		bindMap[field.Name] = bindChan
		go func(bindChan chan any, closedChan chan struct{}, bindF EmitF) {
			state := newGenState()

			for {
				// now := time.Now()
				select {
				case <-closedChan:
					return
				default:
					// time.Sleep(time.Millisecond)
					value := bindF(state)
					// fmt.Printf("pre chan: %v\n", now.Sub(time.Now()))
					bindChan <- value
					// fmt.Printf("pos chan: %v\n", now.Sub(time.Now()))
				}
			}
		}(bindChan, closedChan, fieldMap[field.Name].(EmitF))
	}

	t := template.New("generator")
	t = t.Option("missingkey=error")

	templateFns := sprig.HermeticTxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) any {
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

	return &GeneratorWithTextTemplate{tpl: parsedTpl, bindMap: bindMap, closed: closedChan}, nil
}

func (gen GeneratorWithTextTemplate) Close() error {
	close(gen.closed)

	return nil
}

func (gen GeneratorWithTextTemplate) Emit(buf *bytes.Buffer) error {
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

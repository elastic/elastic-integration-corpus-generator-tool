// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"io"
	"runtime"
	"text/template"
	"time"
)

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	counter   uint64
	totEvents uint64
	bindMap   map[string]chan any
	tpl       *template.Template
}

func NewGeneratorWithTextTemplate(tpl []byte, cfg Config, fields Fields, totSize uint64) (*GeneratorWithTextTemplate, error) {
	// Preprocess the fields, generating appropriate bound function
	fieldMap := make(map[string]any)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, true); err != nil {
			return nil, err
		}
	}

	// Generate a single event to calculate the total number of events based on its size
	t := template.New("estimate_tot_events")
	t = t.Option("missingkey=error")

	templateFns := sprig.HermeticTxtFuncMap()

	templateFns["timeDuration"] = func(duration int64) time.Duration {
		return time.Duration(duration)
	}

	templateFns["generate"] = func(field string) any {
		state := newGenState()
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

	// Generate appropriate emit channels for each bound function
	chanSize := runtime.GOMAXPROCS(0) / 2
	if chanSize < 1 {
		chanSize = 1
	}

	bindMap := make(map[string]chan any)
	for _, field := range fields {
		bindChan := make(chan any)
		bindMap[field.Name] = bindChan
		go func(bindChan chan any, totEvents uint64, bindF EmitF) {
			state := newGenState()

			for i := uint64(0); i < totEvents; i++ {
				value := bindF(state)
				bindChan <- value
			}
		}(bindChan, totEvents, fieldMap[field.Name].(EmitF))
	}

	t = template.New("generator")
	t = t.Option("missingkey=error")

	templateFns = sprig.HermeticTxtFuncMap()

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

	parsedTpl, err = t.Funcs(templateFns).Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	return &GeneratorWithTextTemplate{tpl: parsedTpl, bindMap: bindMap, totEvents: totEvents}, nil
}

func (gen GeneratorWithTextTemplate) Close() error {
	return nil
}

func (gen GeneratorWithTextTemplate) Emit(buf *bytes.Buffer) error {
	if err := gen.emit(buf); err != nil {
		return err
	}

	return nil
}

func (gen GeneratorWithTextTemplate) emit(buf *bytes.Buffer) error {
	if gen.counter < gen.totEvents {
		err := gen.tpl.Execute(buf, nil)
		if err != nil {
			return err
		}
	} else {
		return io.EOF
	}

	gen.counter += 1

	return nil
}

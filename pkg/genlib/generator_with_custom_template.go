// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"regexp"
)

type emitter struct {
	fieldName string
	fieldType string
	emitFunc  emitFNotReturn
	prefix    []byte
}

// customTemplate holds the compiled, immutable, shareable part of a custom
// template generator. Safe for concurrent use by multiple Generator instances.
type customTemplate struct {
	emitters         []emitter
	trailingTemplate []byte
	totEvents        uint64
	fieldNames       []string
}

// GeneratorWithCustomTemplate pairs a shared customTemplate with per-instance mutable state.
type GeneratorWithCustomTemplate struct {
	tpl   *customTemplate
	state *genState
}

func newGeneratorWithCustomTemplate(template []byte) func(Config, Fields, uint64, options) (Generator, error) {
	return func(cfg Config, flds Fields, totEvents uint64, opts options) (Generator, error) {
		compiled, err := NewCustomTemplate(cfg, flds, totEvents, template)
		if err != nil {
			return nil, err
		}
		var o2 options
		compiled(&o2)
		return o2.make(Config{}, nil, 0, opts)
	}
}

func parseCustomTemplate(template []byte) ([]string, map[string][]byte, []byte) {
	if len(template) == 0 {
		return nil, nil, nil
	}

	tokenizer := regexp.MustCompile(`([^{]*)({{\.[^}]+}})*`)
	allIndexes := tokenizer.FindAllSubmatchIndex(template, -1)

	orderedFields := make([]string, 0, len(allIndexes))
	templateFieldsMap := make(map[string][]byte, len(allIndexes))

	var fieldPrefixBuffer []byte
	var fieldPrefixPreviousN int
	var trimTrailingTemplateN int

	for i, loc := range allIndexes {
		var fieldName []byte
		var fieldPrefix []byte

		if loc[4] > -1 && loc[5] > -1 {
			fieldName = template[loc[4]+3 : loc[5]-2]
		}

		if loc[2] > -1 && loc[3] > -1 {
			fieldPrefix = template[loc[2]:loc[3]]
		}

		if len(fieldName) == 0 {
			if template[fieldPrefixPreviousN] == byte(123) {
				fieldPrefixBuffer = append(fieldPrefixBuffer, byte(123))
			} else {
				if i == len(allIndexes)-1 {
					fieldPrefixBuffer = template[trimTrailingTemplateN:]
				} else {
					fieldPrefixBuffer = append(fieldPrefixBuffer, fieldPrefix...)
					fieldPrefixBufferIdx := bytes.Index(template[trimTrailingTemplateN:], fieldPrefixBuffer)
					if fieldPrefixBufferIdx > 0 {
						trimTrailingTemplateN += fieldPrefixBufferIdx
					}

				}
			}
		} else {
			fieldPrefixBuffer = append(fieldPrefixBuffer, fieldPrefix...)
			trimTrailingTemplateN = loc[5]
			templateFieldsMap[string(fieldName)] = fieldPrefixBuffer
			orderedFields = append(orderedFields, string(fieldName))
			fieldPrefixBuffer = nil
		}

		fieldPrefixPreviousN = loc[2]
	}

	return orderedFields, templateFieldsMap, fieldPrefixBuffer

}

// NewCustomTemplate compiles a custom template from cfg and fields. The result
// is returned as an Option that can be passed to NewGenerator to cheaply create
// multiple Generator instances sharing the compiled template.
//
// If templateBytes is nil, the template is auto-generated from fields (random
// noun keys for object/flattened/nested fields). opts may include WithRandSeed
// (for deterministic noun generation), WithStartTime, and WithTimeSpeed.
//
// Intended usage:
//
//	// Compile once per stream (expensive)
//	opt, err := genlib.NewCustomTemplate(cfg, flds, totEvents, nil)
//
//	// Per drone (cheap — only allocates genState)
//	g, err := genlib.NewGenerator(nil, nil, 0, opt, genlib.WithRandSeed(seed))
func NewCustomTemplate(cfg Config, fields Fields, totEvents uint64, templateBytes []byte, opts ...Option) (Option, error) {
	o := applyOptions(opts)

	if templateBytes == nil {
		tmpState := newGenState(o.randSeed, o.startTime, o.timeSpeed)
		var objectKeysFields Fields
		templateBytes, objectKeysFields = generateCustomTemplateFromField(cfg, fields, tmpState)
		fields = append(fields, objectKeysFields...)
	}

	orderedFields, templateFieldsMap, trailingTemplate := parseCustomTemplate(templateBytes)

	fieldMap := make(map[string]any)
	fieldTypes := make(map[string]string)
	fieldNames := make([]string, 0, len(fields))
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, false); err != nil {
			return nil, err
		}
		fieldTypes[field.Name] = field.Type
		fieldNames = append(fieldNames, field.Name)
	}

	emitters := make([]emitter, 0, len(fieldMap))
	for _, fieldName := range orderedFields {
		emitters = append(emitters, emitter{
			fieldName: fieldName,
			emitFunc:  fieldMap[fieldName].(emitFNotReturn),
			fieldType: fieldTypes[fieldName],
			prefix:    templateFieldsMap[fieldName],
		})
	}

	tpl := &customTemplate{
		emitters:         emitters,
		trailingTemplate: trailingTemplate,
		totEvents:        totEvents,
		fieldNames:       fieldNames,
	}

	return func(o *options) {
		o.make = func(cfg Config, flds Fields, totEvents uint64, opts options) (Generator, error) {
			if len(flds) != 0 || totEvents != 0 || !reflect.ValueOf(cfg).IsZero() {
				return nil, errors.New("cfg, flds and totEvents must be nil/zero when using a pre-compiled template option; pass them to NewCustomTemplate instead")
			}
			state := newGenState(opts.randSeed, opts.startTime, opts.timeSpeed)
			for _, fieldName := range tpl.fieldNames {
				state.prevCacheForDup[fieldName] = make(map[any]struct{})
				state.prevCacheCardinality[fieldName] = make([]any, 0)
			}
			state.totEvents = tpl.totEvents
			return &GeneratorWithCustomTemplate{tpl: tpl, state: state}, nil
		}
	}, nil
}

func (gen *GeneratorWithCustomTemplate) Close() error {
	return nil
}

func (gen *GeneratorWithCustomTemplate) Emit(buf *bytes.Buffer) error {
	if err := gen.emit(buf); err != nil {
		return err
	}

	gen.state.counter += 1

	return nil
}

func (gen *GeneratorWithCustomTemplate) emit(buf *bytes.Buffer) error {
	if gen.tpl.totEvents == 0 || gen.state.counter < gen.tpl.totEvents {
		for _, e := range gen.tpl.emitters {
			buf.Write(e.prefix)
			if err := e.emitFunc(gen.state, buf); err != nil {
				return err
			}
		}

		buf.Write(gen.tpl.trailingTemplate)
	} else {
		return io.EOF
	}

	return nil
}

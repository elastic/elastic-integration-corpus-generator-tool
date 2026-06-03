// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var generateOnFieldNotInFieldsYaml = errors.New("generate called on a field not present in fields yaml definition")

// GeneratorWithTextTemplate
type GeneratorWithTextTemplate struct {
	tpl       *template.Template
	state     *genState
	errChan   chan error
	totEvents uint64
}

// awsAZs list all possible AZs for a specific AWS region
// NOTE: this list is not comprehensive
// missing regions: af-south-1, ap-south-2, ap-southeast-3, ap-southeast-4, eu-central-2, eu-south-1, eu-south-2, me-central-1
var awsAZs map[string][]string = map[string][]string{
	"ap-east-1":      {"ap-east-1a", "ap-east-1b", "ap-east-1c"},
	"ap-northeast-1": {"ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"},
	"ap-northeast-2": {"ap-northeast-2a", "ap-northeast-2b", "ap-northeast-2c", "ap-northeast-2d"},
	"ap-northeast-3": {"ap-northeast-3a", "ap-northeast-3b", "ap-northeast-3c"},
	"ap-south-1":     {"ap-south-1a", "ap-south-1b", "ap-south-1c"},
	"ap-southeast-1": {"ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"},
	"ap-southeast-2": {"ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"},
	"ca-central-1":   {"ca-central-1a", "ca-central-1b", "ca-central-1d"},
	"eu-central-1":   {"eu-central-1a", "eu-central-1b", "eu-central-1c"},
	"eu-north-1":     {"eu-north-1a", "eu-north-1b", "eu-north-1c"},
	"eu-west-1":      {"eu-west-1a", "eu-west-1b", "eu-west-1c"},
	"eu-west-2":      {"eu-west-2a", "eu-west-2b", "eu-west-2c"},
	"eu-west-3":      {"eu-west-3a", "eu-west-3b", "eu-west-3c"},
	"me-south-1":     {"me-south-1a", "me-south-1b", "me-south-1c"},
	"sa-east-1":      {"sa-east-1a", "sa-east-1b", "sa-east-1c"},
	"us-east-1":      {"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f"},
	"us-east-2":      {"us-east-2a", "us-east-2b", "us-east-2c"},
	"us-west-1":      {"us-west-1a", "us-west-1b"},
	"us-west-2":      {"us-west-2a", "us-west-2b", "us-west-2c", "us-west-2d"},
}

func newGeneratorWithTextTemplate(template []byte) func(Config, Fields, uint64, options) (Generator, error) {
	return func(cfg Config, flds Fields, totEvents uint64, opts options) (Generator, error) {
		compiled, err := NewTextTemplate(cfg, flds, totEvents, template)
		if err != nil {
			return nil, err
		}
		var o2 options
		compiled(&o2)
		return o2.make(Config{}, nil, 0, opts)
	}
}

// buildTextFuncMap is the single place where all state-capturing template
// functions are defined. state may be (*genState)(nil) at compile time
// (Parse only checks function signatures, never calls them); any attempt to
// Execute the unbound template will panic immediately, making the bug obvious.
func buildTextFuncMap(state *genState, fieldMap map[string]any, errChan chan error) template.FuncMap {
	return template.FuncMap{
		"generate": func(field string) any {
			bindF, ok := fieldMap[field].(emitF)
			if !ok {
				close(errChan)
				return nil
			}
			return bindF(state)
		},
		"awsAZFromRegion": func(region string) string {
			azs, ok := awsAZs[region]
			if !ok {
				return "NoAZ"
			}
			return azs[state.rand.Intn(len(azs))]
		},
	}
}

// NewTextTemplate compiles a Go text template from cfg and fields. The result
// is returned as an Option that can be passed to NewGenerator to cheaply create
// multiple Generator instances sharing the compiled template.
//
// If templateBytes is nil, the template is auto-generated from fields.
//
// Intended usage:
//
//	// Compile once per stream (expensive)
//	opt, err := genlib.NewTextTemplate(cfg, flds, totEvents, nil)
//
//	// Per drone (cheap — only allocates genState)
//	g, err := genlib.NewGenerator(nil, nil, 0, opt, genlib.WithRandSeed(seed))
func NewTextTemplate(cfg Config, flds Fields, totEvents uint64, templateBytes []byte, opts ...Option) (Option, error) {
	o := applyOptions(opts)

	if templateBytes == nil {
		tmpState := newGenState(o.randSeed, o.startTime, o.timeSpeed)
		var objectKeysFields Fields
		templateBytes, objectKeysFields = generateTextTemplateFromField(cfg, flds, tmpState)
		flds = append(flds, objectKeysFields...)
	}

	fieldMap := make(map[string]any)
	fieldNames := make([]string, 0, len(flds))
	for _, field := range flds {
		if err := bindField(cfg, field, fieldMap, true); err != nil {
			return nil, err
		}
		fieldNames = append(fieldNames, field.Name)
	}

	// Parse with poisoned nil-state closures. Parse only validates function
	// signatures, never calls them; nil-state panics loudly if the template is
	// ever executed without per-drone rebinding via Clone + Funcs.
	allFns := sprig.TxtFuncMap()
	for k, v := range buildTextFuncMap((*genState)(nil), fieldMap, make(chan error)) {
		allFns[k] = v
	}
	t := template.New("generator").Option("missingkey=error")
	parsedTpl, err := t.Funcs(allFns).Parse(string(templateBytes))
	if err != nil {
		return nil, err
	}

	return func(o *options) {
		o.make = func(cfg Config, flds Fields, argTotEvents uint64, opts options) (Generator, error) {
			if len(flds) != 0 || argTotEvents != 0 || !reflect.ValueOf(cfg).IsZero() {
				return nil, errors.New("cfg, flds and totEvents must be nil/zero when using a pre-compiled template option; pass them to NewTextTemplate instead")
			}
			state := newGenState(opts.randSeed, opts.startTime, opts.timeSpeed)
			for _, fieldName := range fieldNames {
				state.prevCacheForDup[fieldName] = make(map[any]struct{})
				state.prevCacheCardinality[fieldName] = make([]any, 0)
			}
			state.totEvents = totEvents

			errChan := make(chan error)
			cloned, err := parsedTpl.Clone()
			if err != nil {
				return nil, err
			}
			cloned.Funcs(buildTextFuncMap(state, fieldMap, errChan))

			return &GeneratorWithTextTemplate{tpl: cloned, totEvents: totEvents, state: state, errChan: errChan}, nil
		}
	}, nil
}

func (gen *GeneratorWithTextTemplate) Close() error {
	return nil
}

func (gen *GeneratorWithTextTemplate) Emit(buf *bytes.Buffer) error {
	if err := gen.emit(buf); err != nil {
		return err
	}

	gen.state.counter += 1
	return nil
}

func (gen *GeneratorWithTextTemplate) emit(buf *bytes.Buffer) error {
	if gen.totEvents == 0 || gen.state.counter < gen.totEvents {
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

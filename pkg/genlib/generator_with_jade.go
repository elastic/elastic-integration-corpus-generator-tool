// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

var heroTemplate = []byte(`
<%!
	import "genlib"

	func generate(field string) interface{} {
		return generateFromHero(fieldName, fieldMap, state)
    }
%>
<%: func Template(fieldMap map[string]genlib.EmitF, state *genlib.GenState, buffer *bytes.Buffer) %>
`)

var mainGo = []byte(`
package main

import (
	"template"
    "genlib"
)

func main() {
		// No way to have it simple: we must serialise config and fields and load them back here 

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
		
        buffer := new(bytes.Buffer)
        template.Template(fieldMap, state, buffer)
        fmt.Fprint(os.Stdout, buffer.Bytes())
}`)

// GeneratorWithHero
type GeneratorWithHero struct {
	state  *GenState
	tmpDir string
}

func NewGeneratorWithHero(tpl []byte, cfg Config, fields Fields) (*GeneratorWithHero, error) {
	if len(tpl) == 0 {
		tpl = []byte("")
	}
	tmpDir, err := os.MkdirTemp("", "hero-*")
	if err != nil {
		return nil, err
	}

	templateDir := path.Join(tmpDir, "template")
	err = os.Mkdir(templateDir, 0770)
	if err != nil {
		return nil, err
	}

	randomFilename := path.Base(tmpDir) + ".html"
	fullPath := path.Join(templateDir, randomFilename)
	err = ioutil.WriteFile(fullPath, []byte(fmt.Sprintf(`%s<%%= %s %%>`, heroTemplate, tpl)), 0660)
	if err != nil {
		return nil, err
	}

	genlibDir := path.Join(tmpDir, "genlib")
	err = os.Mkdir(genlibDir, 0770)
	if err != nil {
		return nil, err
	}

	input, err := ioutil.ReadFile("./generator_interface.go")
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(genlibDir, "generator_interface.go"), input, 0660)
	if err != nil {
		return nil, err
	}

	fullPathMainGo := path.Join(tmpDir, "main.go")
	err = ioutil.WriteFile(fullPathMainGo, mainGo, 0660)
	if err != nil {
		return nil, err
	}

	state := NewGenState()

	return &GeneratorWithHero{state: state, tmpDir: tmpDir}, nil
}

func (gen GeneratorWithHero) Emit(state *GenState, buf *bytes.Buffer) error {
	state = gen.state
	if err := gen.emit(state, buf); err != nil {
		return err
	}

	state.counter += 1

	return nil
}

func (gen GeneratorWithHero) render(buf *bytes.Buffer) {
	buf.WriteString("")
}

func (gen GeneratorWithHero) emit(state *GenState, buf *bytes.Buffer) error {
	gen.render(buf)

	return nil
}

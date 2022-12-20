// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var trailingTemplate []byte

// GeneratorWithTemplate is resolved at construction to a slice of emit functions
type GeneratorWithTemplate struct {
	emitFuncs []emitF
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

func extractObjectKeys(templateFieldMap map[string][]byte, fields Fields) (map[string]struct{}, map[string]Field) {
	objectKeys := make(map[string]struct{})
	objectKeysFields := make(map[string]Field)
	for fieldName := range templateFieldMap {
		fieldNameTokens := strings.Split(fieldName, ".")
		possibleRootFieldName := strings.Join(fieldNameTokens[0:len(fieldNameTokens)-1], ".")
		for _, existingField := range fields {
			if strings.HasSuffix(existingField.Name, possibleRootFieldName) {
				if existingField.Name == possibleRootFieldName || fmt.Sprintf("%s.*", existingField.Name) == possibleRootFieldName {
					objectKeys[fieldName] = struct{}{}
					existingField.Name = fieldName
					objectKeysFields[fieldName] = existingField
				}
			}
		}
	}

	return objectKeys, objectKeysFields
}

func NewGeneratorWithTemplate(template []byte, cfg Config, fields Fields) (*GeneratorWithTemplate, error) {
	// Parse the template and extract relevant information
	orderedFields, templateFieldsMap, fieldPrefixBuffer := parseCustomTemplate(template)
	trailingTemplate = fieldPrefixBuffer

	// Extract object keys for `field.*`-like support
	objectKeys, objectKeysFields := extractObjectKeys(templateFieldsMap, fields)

	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]emitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, objectKeys, templateFieldsMap); err != nil {
			return nil, err
		}
	}

	// Preprocess the object keys, generating appropriate emit functions
	for k := range objectKeysFields {
		field := objectKeysFields[k]
		if err := bindField(cfg, field, fieldMap, objectKeys, templateFieldsMap); err != nil {
			return nil, err
		}
	}

	// Roll into slice of emit functions
	emitFuncs := make([]emitF, 0, len(fieldMap))
	for _, fieldName := range orderedFields {
		f := fieldMap[fieldName]
		emitFuncs = append(emitFuncs, f)
	}

	return &GeneratorWithTemplate{emitFuncs: emitFuncs}, nil
}

func (gen GeneratorWithTemplate) Emit(state *GenState, buf *bytes.Buffer) error {
	if err := gen.emit(state, buf); err != nil {
		return err
	}

	state.counter += 1

	return nil
}

func (gen GeneratorWithTemplate) emit(state *GenState, buf *bytes.Buffer) error {
	dupes := make(map[string]struct{})

	for _, f := range gen.emitFuncs {
		if err := f(state, dupes, buf); err != nil {
			return err
		}
	}

	buf.Write(trailingTemplate)
	return nil
}

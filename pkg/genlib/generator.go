// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/lithammer/shortuuid/v3"
	"math/rand"
	"strings"
)

func fieldValueWrapByType(field Field) string {
	if len(field.Value) > 0 {
		return ""
	}

	switch field.Type {
	case FieldTypeDate, FieldTypeIP:
		return "\""
	case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
		return ""
	case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong:
		return ""
	case FieldTypeConstantKeyword:
		return "\""
	case FieldTypeKeyword:
		return "\""
	case FieldTypeBool:
		return ""
	case FieldTypeObject, FieldTypeNested, FieldTypeFlattened:
		if len(field.ObjectType) > 0 {
			field.Type = field.ObjectType
		} else {
			field.Type = FieldTypeKeyword
		}
		return fieldValueWrapByType(field)
	case FieldTypeGeoPoint:
		return "\""
	default:
		return "\""
	}
}

func generateTemplateFromField(cfg Config, fields Fields) []byte {
	if len(fields) == 0 {
		return nil
	}

	dupes := make(map[string]struct{})
	templateBuffer := bytes.NewBufferString("{")
	for i, field := range fields {
		fieldWrap := fieldValueWrapByType(field)
		if fieldCfg, ok := cfg.GetField(field.Name); ok {
			if fieldCfg.Value != nil {
				fieldWrap = ""
			}
		}

		fieldTrailer := []byte(",")
		if i == len(fields)-1 {
			fieldTrailer = []byte("}")
		}

		if strings.HasSuffix(field.Name, ".*") || field.Type == FieldTypeObject || field.Type == FieldTypeNested || field.Type == FieldTypeFlattened {
			// This is a special case.  We are randomly generating keys on the fly
			// Will set the json field name as "field.Name.N"
			N := 5
			for ii := 0; ii < N; ii++ {
				// Fire or skip
				if rand.Int()%2 == 0 {
					continue
				}

				if string(fieldTrailer) == "}" && ii < N-1 {
					fieldTrailer = []byte(",")
				}

				var try int
				const maxTries = 10
				rNoun := randomdata.Noun()
				_, ok := dupes[rNoun]
				for ; ok && try < maxTries; try++ {
					rNoun = randomdata.Noun()
					_, ok = dupes[rNoun]
				}

				// If all else fails, use a shortuuid.
				// Try to avoid this as it is alloc expensive
				if try >= maxTries {
					rNoun = shortuuid.New()
				}

				dupes[rNoun] = struct{}{}
				fieldNameRoot := replacer.Replace(field.Name)
				fieldTemplate := fmt.Sprintf(`"%s.%s": %s{{.%s.%s}}%s%s`, fieldNameRoot, rNoun, fieldWrap, fieldNameRoot, rNoun, fieldWrap, fieldTrailer)
				templateBuffer.WriteString(fieldTemplate)
			}
		} else {
			fieldTemplate := fmt.Sprintf(`"%s": %s{{.%s}}%s%s`, field.Name, fieldWrap, field.Name, fieldWrap, fieldTrailer)
			templateBuffer.WriteString(fieldTemplate)
		}
	}

	return templateBuffer.Bytes()
}

func NewGenerator(cfg Config, fields Fields) (Generator, error) {
	template := generateTemplateFromField(cfg, fields)

	return NewGeneratorWithTemplate(template, cfg, fields)

}

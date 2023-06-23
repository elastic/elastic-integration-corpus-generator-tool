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
	"time"
)

var customRand *rand.Rand

const (
	textTemplateEngine = iota
	customTemplateEngine
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

func generateCustomTemplateFromField(cfg Config, fields Fields) ([]byte, []Field) {
	return generateTemplateFromField(cfg, fields, customTemplateEngine)
}

func generateTextTemplateFromField(cfg Config, fields Fields) ([]byte, []Field) {
	return generateTemplateFromField(cfg, fields, textTemplateEngine)
}

func generateTemplateFromField(cfg Config, fields Fields, templateEngine int) ([]byte, []Field) {
	if len(fields) == 0 {
		return nil, nil
	}

	dupes := make(map[string]struct{})
	objectKeysField := make([]Field, 0, len(fields))

	templatePrefix := "{ "
	templateBuffer := bytes.NewBufferString(templatePrefix)
	for i, field := range fields {
		fieldWrap := fieldValueWrapByType(field)
		if fieldCfg, ok := cfg.GetField(field.Name); ok {
			if fieldCfg.Value != nil {
				fieldWrap = ""
			}
		}

		fieldTrailer := []byte(",")
		if i == len(fields)-1 {
			fieldTrailer = []byte(" }")
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
				var fieldTemplate string

				fieldNameRoot := replacer.Replace(field.Name)
				fieldVariableName := fieldNormalizerRegex.ReplaceAllString(fmt.Sprintf("%s%s", fieldNameRoot, rNoun), "")
				fieldVariableName += "Var"
				if field.Type == FieldTypeDate {
					if templateEngine == textTemplateEngine {
						fieldTemplate = fmt.Sprintf(`{{ $%s := generate "%s.%s" }}"%s.%s": %s{{$%s.Format "2006-01-02T15:04:05.999999Z07:00"}}%s%s`, fieldVariableName, fieldNameRoot, rNoun, fieldNameRoot, rNoun, fieldWrap, fieldVariableName, fieldWrap, fieldTrailer)
					} else if templateEngine == customTemplateEngine {
						fieldTemplate = fmt.Sprintf(`"%s.%s": %s{{.%s.%s}}%s%s`, fieldNameRoot, rNoun, fieldWrap, fieldNameRoot, rNoun, fieldWrap, fieldTrailer)
					}
				} else {
					if templateEngine == textTemplateEngine {
						fieldTemplate = fmt.Sprintf(`"%s.%s": %s{{generate "%s.%s"}}%s%s`, fieldNameRoot, rNoun, fieldWrap, fieldNameRoot, rNoun, fieldWrap, fieldTrailer)
					} else if templateEngine == customTemplateEngine {
						fieldTemplate = fmt.Sprintf(`"%s.%s": %s{{.%s.%s}}%s%s`, fieldNameRoot, rNoun, fieldWrap, fieldNameRoot, rNoun, fieldWrap, fieldTrailer)
					}
				}

				originalFieldName := field.Name
				field.Name = fieldNameRoot + "." + rNoun
				objectKeysField = append(objectKeysField, field)
				field.Name = originalFieldName

				templateBuffer.WriteString(fieldTemplate)
			}
		} else {
			var fieldTemplate string
			fieldVariableName := fieldNormalizerRegex.ReplaceAllString(field.Name, "")
			fieldVariableName += "Var"
			if field.Type == FieldTypeDate {
				if templateEngine == textTemplateEngine {
					fieldTemplate = fmt.Sprintf(`{{ $%s := generate "%s" }}"%s": %s{{$%s.Format "2006-01-02T15:04:05.999999Z07:00"}}%s%s`, fieldVariableName, field.Name, field.Name, fieldWrap, fieldVariableName, fieldWrap, fieldTrailer)
				} else if templateEngine == customTemplateEngine {
					fieldTemplate = fmt.Sprintf(`"%s": %s{{.%s}}%s%s`, field.Name, fieldWrap, field.Name, fieldWrap, fieldTrailer)
				}
			} else {
				if templateEngine == textTemplateEngine {
					fieldTemplate = fmt.Sprintf(`"%s": %s{{generate "%s"}}%s%s`, field.Name, fieldWrap, field.Name, fieldWrap, fieldTrailer)
				} else if templateEngine == customTemplateEngine {
					fieldTemplate = fmt.Sprintf(`"%s": %s{{.%s}}%s%s`, field.Name, fieldWrap, field.Name, fieldWrap, fieldTrailer)
				}
			}

			templateBuffer.WriteString(fieldTemplate)
		}
	}

	return templateBuffer.Bytes(), objectKeysField
}

func NewGenerator(cfg Config, flds Fields, totEvents uint64) (Generator, error) {
	template, objectKeysField := generateCustomTemplateFromField(cfg, flds)
	flds = append(flds, objectKeysField...)

	return NewGeneratorWithCustomTemplate(template, cfg, flds, totEvents)
}

// InitGeneratorTimeNow sets base timeNow for `date` field
func InitGeneratorTimeNow(timeNow time.Time) {
	// set timeNowToBind to --now flag (already parsed or now)
	timeNowToBind = timeNow
}

// InitGeneratorRandSeed sets rand seed
func InitGeneratorRandSeed(randSeed int64) {
	// set rand and randomdata seed to --seed flag (custom or 1)
	customRand = rand.New(rand.NewSource(randSeed))
	randomdata.CustomRand(customRand)
}

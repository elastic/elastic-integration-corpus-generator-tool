// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/lithammer/shortuuid/v3"
)

var templateFieldMap map[string][]byte
var trailingTemplate []byte

// GeneratorWithTemplate is resolved at construction to a slice of emit functions
type GeneratorWithTemplate struct {
	emitFuncs []emitF
}

func generateTemplateFromField(cfg Config, fields Fields) []byte {
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
			for i := 0; i < N; i++ {
				if string(fieldTrailer) == "}" && i < N-1 {
					fieldTrailer = []byte(",")
				}

				fieldNameRoot := replacer.Replace(field.Name)
				fieldTemplate := fmt.Sprintf(`"%s.%d": %s{{.%s.%d}}%s%s`, fieldNameRoot, i, fieldWrap, fieldNameRoot, i, fieldWrap, fieldTrailer)
				templateBuffer.WriteString(fieldTemplate)
			}
		} else {
			fieldTemplate := fmt.Sprintf(`"%s": %s{{.%s}}%s%s`, field.Name, fieldWrap, field.Name, fieldWrap, fieldTrailer)
			templateBuffer.WriteString(fieldTemplate)
		}
	}

	return templateBuffer.Bytes()
}

func NewGeneratorWithTemplate(template []byte, cfg Config, fields Fields) (*GeneratorWithTemplate, error) {
	if len(template) == 0 {
		template = generateTemplateFromField(cfg, fields)
	}

	tokenizer := regexp.MustCompile(`([^{]*)({{\.[^}]+}})*`)
	allIndexes := tokenizer.FindAllSubmatchIndex(template, -1)

	var fieldPrefixBuffer []byte
	var fieldPrefixPreviousN int

	orderedField := make([]string, 0, len(allIndexes))
	templateFieldMap = make(map[string][]byte, len(allIndexes))

	for _, loc := range allIndexes {
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
				fieldPrefixBuffer = append(fieldPrefixBuffer, fieldPrefix...)
			}
		} else {
			fieldPrefixBuffer = append(fieldPrefixBuffer, fieldPrefix...)
			templateFieldMap[string(fieldName)] = fieldPrefixBuffer
			orderedField = append(orderedField, string(fieldName))
			fieldPrefixBuffer = nil
		}

		fieldPrefixPreviousN = loc[2]
	}

	trailingTemplate = fieldPrefixBuffer
	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]emitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap, true); err != nil {
			return nil, err
		}
	}

	// Roll into slice of emit functions
	emitFuncs := make([]emitF, 0, len(fieldMap))
	for _, fieldName := range orderedField {
		f := fieldMap[fieldName]
		emitFuncs = append(emitFuncs, f)
	}

	return &GeneratorWithTemplate{emitFuncs: emitFuncs}, nil
}

func fieldValueWrapByType(field Field) string {
	fieldType := field.Type
	switch fieldType {
	case FieldTypeDate, FieldTypeIP:
		return "\""
	case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
		return ""
	case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong: // TODO: generate > 63 bit values for unsigned_long
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

func bindConstantKeywordWithTemplate(field Field, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		value, ok := state.prevCache[field.Name].(string)
		if !ok {
			value = randomdata.Noun()
			state.prevCache[field.Name] = value
		}
		buf.Write(prefix)
		buf.WriteString(value)
		return nil
	}

	return nil
}

func bindKeywordWithTemplate(fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	if len(fieldCfg.Enum) > 0 {
		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			idx := rand.Intn(len(fieldCfg.Enum))
			value := fieldCfg.Enum[idx]
			buf.Write(prefix)
			buf.WriteString(value)
			return nil
		}
	} else if len(field.Example) > 0 {

		totWords := len(keywordRegex.Split(field.Example, -1))

		var joiner string
		if strings.Contains(field.Example, "\\.") {
			joiner = "\\."
		} else if strings.Contains(field.Example, "-") {
			joiner = "-"
		} else if strings.Contains(field.Example, "_") {
			joiner = "_"
		} else if strings.Contains(field.Example, " ") {
			joiner = " "
		}

		return bindJoinRandWithTemplate(field, totWords, joiner, fieldMap)
	} else {
		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			value := randomdata.Noun()
			buf.Write(prefix)
			buf.WriteString(value)
			return nil
		}
	}
	return nil
}

func bindJoinRandWithTemplate(field Field, N int, joiner string, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)

		for i := 0; i < N-1; i++ {
			buf.WriteString(randomdata.Noun())
			buf.WriteString(joiner)
		}
		buf.WriteString(randomdata.Noun())
		return nil
	}

	return nil
}

func bindStaticWithTemplate(field Field, v interface{}, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	vstr, err := json.Marshal(v)
	if err != nil {
		return err
	}

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)
		buf.Write(vstr)
		return nil
	}

	return nil
}

func bindBoolWithTemplate(field Field, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)
		switch rand.Int() % 2 {
		case 0:
			buf.WriteString("false")
		case 1:
			buf.WriteString("true")
		}
		return nil
	}

	return nil
}

func bindGeoPointWithTemplate(field Field, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)
		err := randGeoPoint(buf)
		return err
	}

	return nil
}

func bindWordNWithTemplate(field Field, n int, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)
		genNounsN(rand.Intn(n), buf)
		return nil
	}

	return nil
}

func bindNearTimeWithTemplate(field Field, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		buf.Write(prefix)
		buf.WriteString(newTime.Format(FieldTypeTimeLayout))
		return nil
	}

	return nil
}

func bindIPWithTemplate(field Field, fieldMap map[string]emitF) error {
	prefix := templateFieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)

		i0 := rand.Intn(255)
		i1 := rand.Intn(255)
		i2 := rand.Intn(255)
		i3 := rand.Intn(255)

		_, err := fmt.Fprintf(buf, "%d.%d.%d.%d", i0, i1, i2, i3)
		return err
	}

	return nil
}

func bindLongWithTemplate(fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

	prefix := templateFieldMap[field.Name]

	if fuzziness <= 0 {
		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			buf.Write(prefix)
			v := make([]byte, 0, 32)
			v = strconv.AppendInt(v, int64(dummyFunc()), 10)
			buf.Write(v)
			return nil
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		dummyInt := dummyFunc()
		if previousDummyInt, ok := state.prevCache[field.Name].(int); ok {
			adjustedRatio := 1. - float64(rand.Intn(fuzziness))/100.
			if rand.Int()%2 == 0 {
				adjustedRatio = 1. + float64(rand.Intn(fuzziness))/100.
			}
			dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
		}
		state.prevCache[field.Name] = dummyInt
		buf.Write(prefix)
		v := make([]byte, 0, 32)
		v = strconv.AppendInt(v, int64(dummyInt), 10)
		buf.Write(v)
		return nil
	}

	return nil
}

func bindDoubleWithTemplate(fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

	prefix := templateFieldMap[field.Name]

	if fuzziness <= 0 {
		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			dummyFloat := float64(dummyFunc()) / rand.Float64()
			buf.Write(prefix)
			_, err := fmt.Fprintf(buf, "%f", dummyFloat)
			return err
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		dummyFloat := float64(dummyFunc()) / rand.Float64()
		if previousDummyFloat, ok := state.prevCache[field.Name].(float64); ok {
			adjustedRatio := 1. - float64(rand.Intn(fuzziness))/100.
			if rand.Int()%2 == 0 {
				adjustedRatio = 1. + float64(rand.Intn(fuzziness))/100.
			}
			dummyFloat = previousDummyFloat * adjustedRatio
		}
		state.prevCache[field.Name] = dummyFloat
		buf.Write(prefix)
		_, err := fmt.Fprintf(buf, "%f", dummyFloat)
		return err
	}

	return nil
}

func bindCardinalityWithTemplate(cfg Config, field Field, fieldMap map[string]emitF) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((1000. / float64(fieldCfg.Cardinality))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByType(cfg, field, fieldMap, true); err != nil {
		return err
	}

	// We will wrap the function we just generated
	boundF := fieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		var va []bytes.Buffer

		if v, ok := state.prevCache[field.Name]; ok {
			va = v.([]bytes.Buffer)
		}

		// Have we rolled over once?  If not, generate a value and cache it.
		if len(va) < cardinality {

			// Do college try dupe detection on value;
			// Allow dupe if no unique value in nTries.
			nTries := 11 // "These go to 11."
			var tmp bytes.Buffer
			for i := 0; i < nTries; i++ {

				tmp.Reset()
				if err := boundF(state, dupes, &tmp); err != nil {
					return err
				}

				if !isDupe(va, tmp.Bytes()) {
					break
				}
			}

			va = append(va, tmp)
			state.prevCache[field.Name] = va
		}

		idx := int(state.counter % uint64(cardinality))

		// Safety check; should be a noop
		if idx >= len(va) {
			idx = len(va) - 1
		}

		choice := va[idx]
		buf.Write(choice.Bytes())
		return nil
	}

	return nil

}

func makeDynamicStubWithTemplate(root string, key int, boundF emitF) emitF {
	fieldName := fmt.Sprintf("{{.%s.%d}}", root, key)
	fieldReplace := []byte(fieldName)
	prefix := templateFieldMap[fieldName]

	return func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		v := state.pool.Get()
		tmp := v.(*bytes.Buffer)
		tmp.Reset()
		tmp.Write(prefix)
		defer state.pool.Put(tmp)

		// Fire the bound function, write into temp buffer
		if err := boundF(state, dupes, tmp); err != nil {
			return err
		}

		if bytes.Contains(tmp.Bytes(), fieldReplace) {
			return fmt.Errorf("Malformed dynamic function payload %s", tmp.String())
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

		// ok, formatted as expected
		// rename the field
		replaced := bytes.ReplaceAll(tmp.Bytes(), []byte(fmt.Sprintf("\"%s.%d\"", root, key)), []byte(fmt.Sprintf("\"%s.%s\"", root, rNoun)))
		buf.Write(replaced)
		return nil
	}
}

func (gen GeneratorWithTemplate) Emit(state *GenState, buf *bytes.Buffer) error {
	buf.Truncate(0)
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

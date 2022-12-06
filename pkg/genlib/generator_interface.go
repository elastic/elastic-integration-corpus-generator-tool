// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/fields"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	Fields      = fields.Fields
	Field       = fields.Field
	Config      = config.Config
	ConfigField = config.ConfigField
)

const (
	FieldTypeBool            = "boolean"
	FieldTypeKeyword         = "keyword"
	FieldTypeConstantKeyword = "constant_keyword"
	FieldTypeDate            = "date"
	FieldTypeIP              = "ip"
	FieldTypeDouble          = "double"
	FieldTypeFloat           = "float"
	FieldTypeHalfFloat       = "half_float"
	FieldTypeScaledFloat     = "scaled_float"
	FieldTypeInteger         = "integer"
	FieldTypeLong            = "long"
	FieldTypeUnsignedLong    = "unsigned_long"
	FieldTypeObject          = "object"
	FieldTypeNested          = "nested"
	FieldTypeFlattened       = "flattened"
	FieldTypeGeoPoint        = "geo_point"

	FieldTypeTimeRange  = 3600 // seconds
	FieldTypeTimeLayout = "2006-01-02T15:04:05.999999Z07:00"
)

var (
	replacer     = strings.NewReplacer(".*", "")
	keywordRegex = regexp.MustCompile("(\\.|-|_|\\s){1,1}")
)

// Typedef of the internal emit function
type emitF func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error

type Generator interface {
	Emit(state *GenState, buf *bytes.Buffer) error
}

type GenState struct {
	// event counter
	counter uint64

	// internal buffer pool to decrease load on GC
	pool sync.Pool

	// previous value cache; necessary for fuzziness, cardinality, etc.
	prevCache map[string]interface{}
}

func NewGenState() *GenState {
	return &GenState{
		prevCache: make(map[string]interface{}),
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func bindField(cfg Config, field Field, fieldMap map[string]emitF, objectKeys map[string]struct{}, templateFieldMap map[string][]byte) error {

	// Check for hardcoded field value
	if len(field.Value) > 0 {
		return bindStatic(templateFieldMap[field.Name], field, field.Value, fieldMap)
	}

	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		return bindStatic(templateFieldMap[field.Name], field, fieldCfg.Value, fieldMap)
	}

	if fieldCfg.Cardinality > 0 {
		return bindCardinality(cfg, field, fieldMap, objectKeys, templateFieldMap)
	}

	return bindByType(cfg, field, fieldMap, objectKeys, templateFieldMap)
}

// Check for dupes O(n)
func isDupe(va []bytes.Buffer, dst []byte) bool {
	var dupe bool
	for _, b := range va {
		if bytes.Equal(dst, b.Bytes()) {
			dupe = true
			break
		}
	}
	return dupe
}

func bindByType(cfg Config, field Field, fieldMap map[string]emitF, objectKeys map[string]struct{}, templateFieldMap map[string][]byte) (err error) {

	fieldCfg, _ := cfg.GetField(field.Name)

	switch field.Type {
	case FieldTypeDate:
		err = bindNearTime(templateFieldMap[field.Name], field, fieldMap)
	case FieldTypeIP:
		err = bindIP(templateFieldMap[field.Name], field, fieldMap)
	case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
		err = bindDouble(templateFieldMap[field.Name], fieldCfg, field, fieldMap)
	case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong: // TODO: generate > 63 bit values for unsigned_long
		err = bindLong(templateFieldMap[field.Name], fieldCfg, field, fieldMap)
	case FieldTypeConstantKeyword:
		err = bindConstantKeyword(templateFieldMap[field.Name], field, fieldMap)
	case FieldTypeKeyword:
		err = bindKeyword(templateFieldMap[field.Name], fieldCfg, field, fieldMap)
	case FieldTypeBool:
		err = bindBool(templateFieldMap[field.Name], field, fieldMap)
	case FieldTypeObject, FieldTypeNested, FieldTypeFlattened:
		err = bindObject(cfg, fieldCfg, field, fieldMap, objectKeys, templateFieldMap)
	case FieldTypeGeoPoint:
		err = bindGeoPoint(templateFieldMap[field.Name], field, fieldMap)
	default:
		err = bindWordN(templateFieldMap[field.Name], field, 25, fieldMap)
	}

	return
}

func makeIntFunc(fieldCfg ConfigField, field Field) func() int {
	maxValue := fieldCfg.Range

	var dummyFunc func() int

	switch {
	case maxValue > 0:
		dummyFunc = func() int { return rand.Intn(maxValue) }
	case len(field.Example) == 0:
		dummyFunc = func() int { return rand.Intn(10) }
	default:
		totDigit := len(field.Example)
		max := int(math.Pow10(totDigit))
		dummyFunc = func() int {
			return rand.Intn(max)
		}
	}

	return dummyFunc
}

func bindObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitF, objectKeys map[string]struct{}, templateFieldMap map[string][]byte) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	if len(fieldCfg.ObjectKeys) > 0 {
		for _, objectsKey := range fieldCfg.ObjectKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap, objectKeys, templateFieldMap); err != nil {
				return err
			}
		}

		return nil
	}

	// This is a special case.  We are randomly generating keys on the fly
	// Will creating a special emit function that binds statically,
	// but only fires randomly.
	N := 5
	return bindDynamicObject(cfg, fieldCfg, field, fieldMap, N, objectKeys, templateFieldMap)
}

func bindDynamicObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitF, N int, objectKeys map[string]struct{}, templateFieldMap map[string][]byte) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]emitF)

	if err := bindField(cfg, field, dynMap, objectKeys, templateFieldMap); err != nil {
		return err
	}
	stub := makeDynamicStub(templateFieldMap[field.Name], dynMap[field.Name])
	fieldMap[field.Name] = stub

	return nil
}

func genNounsN(n int, buf *bytes.Buffer) {

	for i := 0; i < n-1; i++ {
		buf.WriteString(randomdata.Noun())
		buf.WriteByte(' ')
	}

	buf.WriteString(randomdata.Noun())
}

func randGeoPoint(buf *bytes.Buffer) error {
	lat := rand.Intn(181) - 90
	var latD int
	if lat != -90 && lat != 90 {
		latD = rand.Intn(100)
	}
	var longD int
	long := rand.Intn(361) - 180
	if long != -180 && long != 180 {
		longD = rand.Intn(100)
	}
	_, err := fmt.Fprintf(buf, "%d.%d,%d.%d", lat, latD, long, longD)
	return err
}

func bindConstantKeyword(prefix []byte, field Field, fieldMap map[string]emitF) error {
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

func bindKeyword(prefix []byte, fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {
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

		return bindJoinRand(prefix, field, totWords, joiner, fieldMap)
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

func bindJoinRand(prefix []byte, field Field, N int, joiner string, fieldMap map[string]emitF) error {
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

func bindStatic(prefix []byte, field Field, v interface{}, fieldMap map[string]emitF) error {
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

func bindBool(prefix []byte, field Field, fieldMap map[string]emitF) error {
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

func bindGeoPoint(prefix []byte, field Field, fieldMap map[string]emitF) error {
	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)
		err := randGeoPoint(buf)
		return err
	}

	return nil
}

func bindWordN(prefix []byte, field Field, n int, fieldMap map[string]emitF) error {
	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.Write(prefix)
		genNounsN(rand.Intn(n), buf)
		return nil
	}

	return nil
}

func bindNearTime(prefix []byte, field Field, fieldMap map[string]emitF) error {
	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		buf.Write(prefix)
		buf.WriteString(newTime.Format(FieldTypeTimeLayout))
		return nil
	}

	return nil
}

func bindIP(prefix []byte, field Field, fieldMap map[string]emitF) error {
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

func bindLong(prefix []byte, fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

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

func bindDouble(prefix []byte, fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

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

func bindCardinality(cfg Config, field Field, fieldMap map[string]emitF, objectKeys map[string]struct{}, templateFieldMap map[string][]byte) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((1000. / float64(fieldCfg.Cardinality))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByType(cfg, field, fieldMap, objectKeys, templateFieldMap); err != nil {
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

func makeDynamicStub(prefix []byte, boundF emitF) emitF {
	return func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		v := state.pool.Get()
		tmp := v.(*bytes.Buffer)
		tmp.Reset()
		defer state.pool.Put(tmp)

		// Fire the bound function, write into temp buffer
		if err := boundF(state, dupes, tmp); err != nil {
			return err
		}

		// If bound function did not write for some reason; abort
		if tmp.Len() == 0 {
			return nil
		}

		// ok, formatted as expected, swap it out the payload
		buf.Write(prefix)
		buf.Write(tmp.Bytes())
		return nil
	}
}

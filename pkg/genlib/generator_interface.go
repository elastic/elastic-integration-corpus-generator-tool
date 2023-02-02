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
	"testing"
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
	replacer             = strings.NewReplacer(".*", "")
	fieldNormalizerRegex = regexp.MustCompile("[^a-zA-Z0-9]")
	keywordRegex         = regexp.MustCompile("(\\.|-|_|\\s){1,1}")
)

// This is the emit function for the custom template engine where we stream content directly to the output buffer and no need a return value
type emitFNotReturn func(state *GenState, buf *bytes.Buffer) error

// EmitF Typedef of the internal emit function
type EmitF func(state *GenState, buf *bytes.Buffer) (interface{}, error)

type Generator interface {
	Emit(state *GenState, buf *bytes.Buffer) error
	Close() error
}

type GenState struct {
	// event counter
	counter uint64

	// internal buffer pool to decrease load on GC
	pool sync.Pool

	// previous value cache; necessary for fuzziness, cardinality, etc.
	prevCache any
}

func NewGenState() *GenState {
	return &GenState{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func bindField(cfg Config, field Field, fieldMapWithReturn map[string]EmitF, fieldMap map[string]emitFNotReturn, templateFieldMap map[string][]byte, withReturn bool) error {

	// Check for hardcoded field value
	if len(field.Value) > 0 {
		if withReturn {
			return bindStaticWithReturn(field, field.Value, fieldMapWithReturn)
		} else {
			return bindStatic(templateFieldMap[field.Name], field, field.Value, fieldMap)
		}
	}

	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		if withReturn {
			return bindStaticWithReturn(field, fieldCfg.Value, fieldMapWithReturn)
		} else {
			return bindStatic(templateFieldMap[field.Name], field, fieldCfg.Value, fieldMap)
		}
	}

	if fieldCfg.Cardinality.Numerator > 0 {
		if withReturn {
			return bindCardinalityWithReturn(cfg, field, fieldMapWithReturn)
		} else {
			return bindCardinality(templateFieldMap[field.Name], cfg, field, fieldMap, templateFieldMap)
		}
	}

	if withReturn {
		return bindByTypeWithReturn(cfg, field, fieldMapWithReturn)
	} else {
		return bindByType(cfg, field, fieldMap, templateFieldMap)
	}
}

// Check for dupes O(n)
func isDupeByteSlice(va []bytes.Buffer, dst []byte) bool {
	var dupe bool
	for _, b := range va {
		if bytes.Equal(dst, b.Bytes()) {
			dupe = true
			break
		}
	}
	return dupe
}

// Check for dupes O(n)
func isDupeInterface(va []interface{}, dst interface{}) bool {
	var dupe bool
	for _, b := range va {
		if b == dst {
			dupe = true
			break
		}
	}
	return dupe
}

func bindByType(cfg Config, field Field, fieldMap map[string]emitFNotReturn, templateFieldMap map[string][]byte) (err error) {

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
		err = bindObject(cfg, fieldCfg, field, fieldMap, templateFieldMap)
	case FieldTypeGeoPoint:
		err = bindGeoPoint(templateFieldMap[field.Name], field, fieldMap)
	default:
		err = bindWordN(templateFieldMap[field.Name], field, 25, fieldMap)
	}

	return
}

func bindByTypeWithReturn(cfg Config, field Field, fieldMap map[string]EmitF) (err error) {

	fieldCfg, _ := cfg.GetField(field.Name)

	switch field.Type {
	case FieldTypeDate:
		err = bindNearTimeWithReturn(field, fieldMap)
	case FieldTypeIP:
		err = bindIPWithReturn(field, fieldMap)
	case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
		err = bindDoubleWithReturn(fieldCfg, field, fieldMap)
	case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong: // TODO: generate > 63 bit values for unsigned_long
		err = bindLongWithReturn(fieldCfg, field, fieldMap)
	case FieldTypeConstantKeyword:
		err = bindConstantKeywordWithReturn(field, fieldMap)
	case FieldTypeKeyword:
		err = bindKeywordWithReturn(fieldCfg, field, fieldMap)
	case FieldTypeBool:
		err = bindBoolWithReturn(field, fieldMap)
	case FieldTypeObject, FieldTypeNested, FieldTypeFlattened:
		err = bindObjectWithReturn(cfg, fieldCfg, field, fieldMap)
	case FieldTypeGeoPoint:
		err = bindGeoPointWithReturn(field, fieldMap)
	default:
		err = bindWordNWithReturn(field, 25, fieldMap)
	}

	return
}

func makeFloatFunc(fieldCfg ConfigField, field Field) func() float64 {
	minValue := float64(0)
	maxValue := float64(0)

	switch fieldCfg.Range.Min.(type) {
	case float64:
		minValue = fieldCfg.Range.Min.(float64)
	case uint64:
		minValue = float64(fieldCfg.Range.Min.(uint64))
	case int64:
		minValue = float64(fieldCfg.Range.Min.(int64))
	}

	switch fieldCfg.Range.Max.(type) {
	case float64:
		maxValue = fieldCfg.Range.Max.(float64)
	case uint64:
		maxValue = float64(fieldCfg.Range.Max.(uint64))
	case int64:
		maxValue = float64(fieldCfg.Range.Max.(int64))
	}

	var dummyFunc func() float64

	switch {
	case maxValue > 0:
		dummyFunc = func() float64 { return minValue + rand.Float64()*(maxValue-minValue) }
	case len(field.Example) == 0:
		dummyFunc = func() float64 { return rand.Float64() * 10 }
	default:
		totDigit := len(field.Example)
		max := math.Pow10(totDigit)
		dummyFunc = func() float64 {
			return rand.Float64() * max
		}
	}

	return dummyFunc
}

func makeIntFunc(fieldCfg ConfigField, field Field) func() int {
	minValue := 0
	maxValue := 0

	switch fieldCfg.Range.Min.(type) {
	case uint64:
		minValue = int(fieldCfg.Range.Min.(uint64))
	case int64:
		minValue = int(fieldCfg.Range.Min.(int64))
	}

	switch fieldCfg.Range.Max.(type) {
	case uint64:
		maxValue = int(fieldCfg.Range.Max.(uint64))
	case int64:
		maxValue = int(fieldCfg.Range.Max.(int64))
	}

	var dummyFunc func() int

	switch {
	case maxValue > 0:
		dummyFunc = func() int { return rand.Intn(maxValue-minValue) + minValue }
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

func bindObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitFNotReturn, templateFieldMap map[string][]byte) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	if len(fieldCfg.ObjectKeys) > 0 {
		for _, objectsKey := range fieldCfg.ObjectKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, nil, fieldMap, templateFieldMap, false); err != nil {
				return err
			}
		}

		return nil
	}

	return bindDynamicObject(cfg, field, fieldMap, templateFieldMap)
}

func bindDynamicObject(cfg Config, field Field, fieldMap map[string]emitFNotReturn, templateFieldMap map[string][]byte) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]emitFNotReturn)

	if err := bindField(cfg, field, nil, dynMap, templateFieldMap, false); err != nil {
		return err
	}
	stub := makeDynamicStub(dynMap[field.Name])
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

func genNounsNWithReturn(n int) string {
	value := ""
	for i := 0; i < n-1; i++ {
		value += randomdata.Noun() + " "
	}

	value += randomdata.Noun()

	return value
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

func randGeoPointWithReturn() string {
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

	return fmt.Sprintf("%d.%d,%d.%d", lat, latD, long, longD)
}

func bindConstantKeyword(prefix []byte, field Field, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		value, ok := state.prevCache.(string)
		if !ok {
			value = randomdata.Noun()
			state.prevCache = value
		}
		buf.Write(prefix)
		buf.WriteString(value)
		return nil
	}

	return nil
}

func bindKeyword(prefix []byte, fieldCfg ConfigField, field Field, fieldMap map[string]emitFNotReturn) error {
	if len(fieldCfg.Enum) > 0 {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
			idx := rand.Intn(len(fieldCfg.Enum))
			buf.Write(prefix)
			buf.WriteString(fieldCfg.Enum[idx])
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
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
			buf.Write(prefix)
			buf.WriteString(randomdata.Noun())
			return nil
		}
	}
	return nil
}

func bindJoinRand(prefix []byte, field Field, N int, joiner string, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
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

func bindStatic(prefix []byte, field Field, v interface{}, fieldMap map[string]emitFNotReturn) error {
	vstr, err := json.Marshal(v)
	if err != nil {
		return err
	}

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		buf.Write(prefix)
		buf.Write(vstr)
		return nil
	}

	return nil
}

func bindBool(prefix []byte, field Field, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
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

func bindGeoPoint(prefix []byte, field Field, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		buf.Write(prefix)
		return randGeoPoint(buf)
	}

	return nil
}

func bindWordN(prefix []byte, field Field, n int, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		buf.Write(prefix)
		genNounsN(rand.Intn(n), buf)
		return nil
	}

	return nil
}

func bindNearTime(prefix []byte, field Field, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		buf.Write(prefix)
		buf.WriteString(newTime.Format(FieldTypeTimeLayout))
		return nil
	}

	return nil
}

func bindIP(prefix []byte, field Field, fieldMap map[string]emitFNotReturn) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
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

func bindLong(prefix []byte, fieldCfg ConfigField, field Field, fieldMap map[string]emitFNotReturn) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzzinessNumerator := fieldCfg.Fuzziness.Numerator
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
			buf.Write(prefix)
			v := make([]byte, 0, 32)
			v = strconv.AppendInt(v, int64(dummyFunc()), 10)
			buf.Write(v)
			return nil
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		dummyInt := dummyFunc()
		if previousDummyInt, ok := state.prevCache.(int); ok {
			adjustedRatio := float64(rand.Intn(fuzzinessNumerator)) / fuzzinessDenominator
			if rand.Int()%2 == 0 {
				adjustedRatio += 1.
			} else {
				adjustedRatio = 1. - adjustedRatio
			}
			dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
		}
		state.prevCache = dummyInt
		buf.Write(prefix)
		v := make([]byte, 0, 32)
		v = strconv.AppendInt(v, int64(dummyInt), 10)
		buf.Write(v)
		return nil
	}

	return nil
}

func bindDouble(prefix []byte, fieldCfg ConfigField, field Field, fieldMap map[string]emitFNotReturn) error {

	dummyFunc := makeFloatFunc(fieldCfg, field)

	fuzzinessNumerator := fieldCfg.Fuzziness.Numerator
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
			dummyFloat := dummyFunc()
			buf.Write(prefix)
			_, err := fmt.Fprintf(buf, "%f", dummyFloat)
			return err
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		dummyFloat := dummyFunc()
		if previousDummyFloat, ok := state.prevCache.(float64); ok {
			adjustedRatio := float64(rand.Intn(fuzzinessNumerator)) / fuzzinessDenominator
			if rand.Int()%2 == 0 {
				adjustedRatio += 1.
			} else {
				adjustedRatio = 1. - adjustedRatio
			}
			dummyFloat = previousDummyFloat * adjustedRatio
		}
		state.prevCache = dummyFloat
		buf.Write(prefix)
		_, err := fmt.Fprintf(buf, "%f", dummyFloat)
		return err
	}

	return nil
}

func bindCardinality(prefix []byte, cfg Config, field Field, fieldMap map[string]emitFNotReturn, templateFieldMap map[string][]byte) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((float64(fieldCfg.Cardinality.Denominator) / float64(fieldCfg.Cardinality.Numerator))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByType(cfg, field, fieldMap, templateFieldMap); err != nil {
		return err
	}

	// We will wrap the function we just generated
	boundF := fieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) error {
		var va []bytes.Buffer

		if v, ok := state.prevCache.([]bytes.Buffer); ok {
			va = v
		}

		// Have we rolled over once?  If not, generate a value and cache it.
		if len(va) < cardinality {

			// Do college try dupe detection on value;
			// Allow dupe if no unique value in nTries.
			nTries := 11 // "These go to 11."
			var tmp bytes.Buffer
			for i := 0; i < nTries; i++ {

				tmp.Reset()
				if err := boundF(state, &tmp); err != nil {
					return err
				}

				if !isDupeByteSlice(va, tmp.Bytes()) {
					break
				}
			}

			va = append(va, tmp)
			state.prevCache = va
		}

		idx := int(state.counter % uint64(cardinality))
		state.counter += 1

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

func makeDynamicStub(boundF emitFNotReturn) emitFNotReturn {
	return func(state *GenState, buf *bytes.Buffer) error {
		v := state.pool.Get()
		tmp := v.(*bytes.Buffer)
		tmp.Reset()
		defer state.pool.Put(tmp)

		// Fire the bound function, write into temp buffer
		err := boundF(state, tmp)
		if err != nil {
			return err
		}

		// If bound function did not write for some reason; abort
		if tmp.Len() == 0 {
			return nil
		}

		// ok, formatted as expected, swap it out the payload
		buf.Write(tmp.Bytes())
		return nil
	}
}

func makeDynamicStubWithReturn(boundF EmitF) EmitF {
	return func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		v := state.pool.Get()
		tmp := v.(*bytes.Buffer)
		tmp.Reset()
		defer state.pool.Put(tmp)

		// Fire the bound function, write into temp buffer
		return boundF(state, tmp)
	}
}

func bindConstantKeywordWithReturn(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		value, ok := state.prevCache.(string)
		if !ok {
			value = randomdata.Noun()
			state.prevCache = value
		}
		return value, nil
	}

	return nil
}

func bindKeywordWithReturn(fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {
	if len(fieldCfg.Enum) > 0 {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
			idx := rand.Intn(len(fieldCfg.Enum))
			return fieldCfg.Enum[idx], nil
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

		return bindJoinRandWithReturn(field, totWords, joiner, fieldMap)
	} else {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
			return randomdata.Noun(), nil
		}
	}
	return nil
}

func bindJoinRandWithReturn(field Field, N int, joiner string, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		value := ""
		for i := 0; i < N-1; i++ {
			value += randomdata.Noun() + joiner
		}

		value += randomdata.Noun()

		return value, nil
	}

	return nil
}

func bindStaticWithReturn(field Field, v interface{}, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		return v, nil
	}

	return nil
}

func bindBoolWithReturn(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		switch rand.Int() % 2 {
		case 0:
			return false, nil
		case 1:
			return true, nil
		}

		return nil, nil
	}

	return nil
}

func bindGeoPointWithReturn(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		return randGeoPointWithReturn(), nil
	}

	return nil
}

func bindWordNWithReturn(field Field, n int, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		return genNounsNWithReturn(rand.Intn(n)), nil
	}

	return nil
}

func bindNearTimeWithReturn(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		return newTime, nil
	}

	return nil
}

func bindIPWithReturn(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		i0 := rand.Intn(255)
		i1 := rand.Intn(255)
		i2 := rand.Intn(255)
		i3 := rand.Intn(255)

		return fmt.Sprintf("%d.%d.%d.%d", i0, i1, i2, i3), nil
	}

	return nil
}

func bindLongWithReturn(fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzzinessNumerator := fieldCfg.Fuzziness.Numerator
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
			return dummyFunc(), nil
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		dummyInt := dummyFunc()
		if previousDummyInt, ok := state.prevCache.(int); ok {
			adjustedRatio := 1. - float64(rand.Intn(fuzzinessNumerator))/fuzzinessDenominator
			if rand.Int()%2 == 0 {
				adjustedRatio = 1. + float64(rand.Intn(fuzzinessNumerator))/fuzzinessDenominator
			}
			dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
		}
		state.prevCache = dummyInt
		return dummyInt, nil
	}

	return nil
}

func bindDoubleWithReturn(fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {

	dummyFunc := makeFloatFunc(fieldCfg, field)

	fuzzinessNumerator := fieldCfg.Fuzziness.Numerator
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
			return dummyFunc(), nil
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		dummyFloat := dummyFunc()
		if previousDummyFloat, ok := state.prevCache.(float64); ok {
			adjustedRatio := 1. - float64(rand.Intn(fuzzinessNumerator))/fuzzinessDenominator
			if rand.Int()%2 == 0 {
				adjustedRatio = 1. + float64(rand.Intn(fuzzinessNumerator))/fuzzinessDenominator
			}
			dummyFloat = previousDummyFloat * adjustedRatio
		}
		state.prevCache = dummyFloat
		return dummyFloat, nil
	}

	return nil
}

func bindCardinalityWithReturn(cfg Config, field Field, fieldMap map[string]EmitF) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((float64(fieldCfg.Cardinality.Denominator) / float64(fieldCfg.Cardinality.Numerator))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByTypeWithReturn(cfg, field, fieldMap); err != nil {
		return err
	}

	// We will wrap the function we just generated
	boundFWithReturn := fieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState, buf *bytes.Buffer) (interface{}, error) {
		var va []interface{}

		if v, ok := state.prevCache.([]interface{}); ok {
			va = v
		}

		var value interface{}
		// Have we rolled over once?  If not, generate a value and cache it.
		if len(va) < cardinality {

			// Do college try dupe detection on value;
			// Allow dupe if no unique value in nTries.
			nTries := 11 // "These go to 11."
			var tmp bytes.Buffer
			for i := 0; i < nTries; i++ {
				tmp.Reset()
				var err error
				value, err = boundFWithReturn(state, &tmp)
				if err != nil {
					return value, err
				}

				if !isDupeInterface(va, value) {
					break
				}
			}

			va = append(va, value)
			state.prevCache = va
		}

		idx := int(state.counter % uint64(cardinality))
		state.counter += 1

		// Safety check; should be a noop
		if idx >= len(va) {
			idx = len(va) - 1
		}

		choice := va[idx]

		return choice, nil
	}

	return nil
}

func bindObjectWithReturn(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	if len(fieldCfg.ObjectKeys) > 0 {
		for _, objectsKey := range fieldCfg.ObjectKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap, nil, nil, true); err != nil {
				return err
			}
		}

		return nil
	}

	return bindDynamicObjectWithReturn(cfg, field, fieldMap)
}

func bindDynamicObjectWithReturn(cfg Config, field Field, fieldMap map[string]EmitF) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]EmitF)

	if err := bindField(cfg, field, dynMap, nil, nil, true); err != nil {
		return err
	}
	stub := makeDynamicStubWithReturn(dynMap[field.Name])
	fieldMap[field.Name] = stub

	return nil
}

func unmarshalJSONT[T any](t *testing.T, data []byte) map[string]T {
	m := make(map[string]T)
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

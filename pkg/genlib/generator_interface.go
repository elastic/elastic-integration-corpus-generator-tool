// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"encoding/json"
	"errors"
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
type EmitF func(state *GenState) any

type Generator interface {
	Emit(state *GenState, buf *bytes.Buffer) error
	Close() error
}

type GenState struct {
	// event counter
	counter uint64
	// previous value cache; necessary for fuzziness, cardinality, etc.
	prevCache map[string]any
	// previous value cache for dup check; necessary for cardinality
	prevCacheForDup map[string]map[any]struct{}
	// previous cardinality value cache; necessary for cardinality
	prevCacheCardinality map[string][]any
	// internal buffer pool to decrease load on GC
	pool sync.Pool
}

func NewGenState() *GenState {
	return &GenState{
		prevCache:            make(map[string]any),
		prevCacheForDup:      make(map[string]map[any]struct{}),
		prevCacheCardinality: make(map[string][]any, 0),
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func bindField(cfg Config, field Field, fieldMap map[string]any, withReturn bool) error {

	// Check for hardcoded field value
	if len(field.Value) > 0 {
		if withReturn {
			return bindStaticWithReturn(field, field.Value, fieldMap)
		} else {
			return bindStatic(field, field.Value, fieldMap)
		}
	}

	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		if withReturn {
			return bindStaticWithReturn(field, fieldCfg.Value, fieldMap)
		} else {
			return bindStatic(field, fieldCfg.Value, fieldMap)
		}
	}

	if fieldCfg.Cardinality.Numerator > 0 {
		if withReturn {
			return bindCardinalityWithReturn(cfg, field, fieldMap)
		} else {
			return bindCardinality(cfg, field, fieldMap)
		}
	}

	if withReturn {
		return bindByTypeWithReturn(cfg, field, fieldMap)
	} else {
		return bindByType(cfg, field, fieldMap)
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

func isDupeAny(va map[any]struct{}, dst any) bool {
	_, ok := va[dst]
	return ok
}

// Check for dupes O(n)
func isDupeInterface(va []any, dst any) bool {
	var dupe bool
	for _, b := range va {
		if b == dst {
			dupe = true
			break
		}
	}
	return dupe
}

func bindByType(cfg Config, field Field, fieldMap map[string]any) (err error) {

	fieldCfg, _ := cfg.GetField(field.Name)

	switch field.Type {
	case FieldTypeDate:
		err = bindNearTime(field, fieldMap)
	case FieldTypeIP:
		err = bindIP(field, fieldMap)
	case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
		err = bindDouble(fieldCfg, field, fieldMap)
	case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong: // TODO: generate > 63 bit values for unsigned_long
		err = bindLong(fieldCfg, field, fieldMap)
	case FieldTypeConstantKeyword:
		err = bindConstantKeyword(field, fieldMap)
	case FieldTypeKeyword:
		err = bindKeyword(fieldCfg, field, fieldMap)
	case FieldTypeBool:
		err = bindBool(field, fieldMap)
	case FieldTypeObject, FieldTypeNested, FieldTypeFlattened:
		err = bindObject(cfg, fieldCfg, field, fieldMap)
	case FieldTypeGeoPoint:
		err = bindGeoPoint(field, fieldMap)
	default:
		err = bindWordN(field, 25, fieldMap)
	}

	return
}

func bindByTypeWithReturn(cfg Config, field Field, fieldMap map[string]any) (err error) {

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

func bindObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]any) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	if len(fieldCfg.ObjectKeys) > 0 {
		for _, objectsKey := range fieldCfg.ObjectKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap, false); err != nil {
				return err
			}
		}

		return nil
	}

	return bindDynamicObject(cfg, field, fieldMap)
}

func bindDynamicObject(cfg Config, field Field, fieldMap map[string]any) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]any)

	if err := bindField(cfg, field, dynMap, false); err != nil {
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

	// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
	buf.WriteString(randomdata.Adjective())
	buf.WriteString(randomdata.Noun())
}

func genNounsNWithReturn(n int) string {
	value := ""
	for i := 0; i < n-1; i++ {
		value += randomdata.Noun() + " "
	}

	// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
	value += randomdata.Adjective()
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

func bindConstantKeyword(field Field, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		value, ok := state.prevCache[field.Name].(string)
		if !ok {
			// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
			value = randomdata.Adjective() + randomdata.Noun()
			state.prevCache[field.Name] = value
		}
		buf.WriteString(value)
		return nil
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindKeyword(fieldCfg ConfigField, field Field, fieldMap map[string]any) error {
	if len(fieldCfg.Enum) > 0 {
		var emitFNotReturn emitFNotReturn
		emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
			idx := rand.Intn(len(fieldCfg.Enum))
			buf.WriteString(fieldCfg.Enum[idx])
			return nil
		}

		fieldMap[field.Name] = emitFNotReturn
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

		return bindJoinRand(field, totWords, joiner, fieldMap)
	} else {
		var emitFNotReturn emitFNotReturn
		emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
			// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
			buf.WriteString(randomdata.Adjective() + randomdata.Noun())
			return nil
		}

		fieldMap[field.Name] = emitFNotReturn
	}
	return nil
}

func bindJoinRand(field Field, N int, joiner string, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		for i := 0; i < N-1; i++ {
			buf.WriteString(randomdata.Noun())
			buf.WriteString(joiner)
		}
		// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
		buf.WriteString(randomdata.Adjective())
		buf.WriteString(randomdata.Noun())
		return nil
	}

	fieldMap[field.Name] = emitFNotReturn

	return nil
}

func bindStatic(field Field, v any, fieldMap map[string]any) error {
	vstr, err := json.Marshal(v)
	if err != nil {
		return err
	}

	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		buf.Write(vstr)
		return nil
	}
	fieldMap[field.Name] = emitFNotReturn

	return nil
}

func bindBool(field Field, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		switch rand.Int() % 2 {
		case 0:
			buf.WriteString("false")
		case 1:
			buf.WriteString("true")
		}
		return nil
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindGeoPoint(field Field, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		return randGeoPoint(buf)
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindWordN(field Field, n int, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		genNounsN(rand.Intn(n), buf)
		return nil
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindNearTime(field Field, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		buf.WriteString(newTime.Format(FieldTypeTimeLayout))
		return nil
	}
	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindIP(field Field, fieldMap map[string]any) error {
	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		i0 := rand.Intn(255)
		i1 := rand.Intn(255)
		i2 := rand.Intn(255)
		i3 := rand.Intn(255)

		_, err := fmt.Fprintf(buf, "%d.%d.%d.%d", i0, i1, i2, i3)
		return err
	}

	fieldMap[field.Name] = emitFNotReturn

	return nil
}

func bindLong(fieldCfg ConfigField, field Field, fieldMap map[string]any) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzzinessNumerator := float64(fieldCfg.Fuzziness.Numerator)
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		var emitFNotReturn emitFNotReturn
		emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
			v := make([]byte, 0, 32)
			v = strconv.AppendInt(v, int64(dummyFunc()), 10)
			buf.Write(v)
			return nil
		}

		fieldMap[field.Name] = emitFNotReturn

		return nil
	}

	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		dummyInt := dummyFunc()
		if previousDummyInt, ok := state.prevCache[field.Name].(int); ok {
			adjustedRatio := rand.Float64() * (fuzzinessNumerator / fuzzinessDenominator)
			if rand.Int()%2 == 0 {
				adjustedRatio += 1.
			} else {
				adjustedRatio = 1. - adjustedRatio
			}
			dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
		}
		state.prevCache[field.Name] = dummyInt
		v := make([]byte, 0, 32)
		v = strconv.AppendInt(v, int64(dummyInt), 10)
		buf.Write(v)
		return nil
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindDouble(fieldCfg ConfigField, field Field, fieldMap map[string]any) error {

	dummyFunc := makeFloatFunc(fieldCfg, field)

	fuzzinessNumerator := float64(fieldCfg.Fuzziness.Numerator)
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		var emitFNotReturn emitFNotReturn
		emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
			dummyFloat := dummyFunc()
			_, err := fmt.Fprintf(buf, "%f", dummyFloat)
			return err
		}

		fieldMap[field.Name] = emitFNotReturn
		return nil
	}

	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		dummyFloat := dummyFunc()
		if previousDummyFloat, ok := state.prevCache[field.Name].(float64); ok {
			adjustedRatio := rand.Float64() * (fuzzinessNumerator / fuzzinessDenominator)
			if rand.Int()%2 == 0 {
				adjustedRatio += 1.
			} else {
				adjustedRatio = 1. - adjustedRatio
			}
			dummyFloat = previousDummyFloat * adjustedRatio
		}
		state.prevCache[field.Name] = dummyFloat
		_, err := fmt.Fprintf(buf, "%f", dummyFloat)

		return err
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func bindCardinality(cfg Config, field Field, fieldMap map[string]any) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((float64(fieldCfg.Cardinality.Denominator) / float64(fieldCfg.Cardinality.Numerator))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByType(cfg, field, fieldMap); err != nil {
		return err
	}

	// We will wrap the function we just generated
	boundF, ok := fieldMap[field.Name].(emitFNotReturn)
	if !ok {
		return errors.New("cannot bind cardinality")
	}

	var emitFNotReturn emitFNotReturn
	emitFNotReturn = func(state *GenState, buf *bytes.Buffer) error {
		// Have we rolled over once?  If not, generate a value and cache it.
		if len(state.prevCacheCardinality[field.Name]) < cardinality {

			// Do college try dupe detection on value;
			// Allow dupe if no unique value in nTries.
			nTries := 11 // "These go to 11."
			var tmp bytes.Buffer
			var value []byte
			for i := 0; i < nTries; i++ {

				tmp.Reset()
				if err := boundF(state, &tmp); err != nil {
					return err
				}

				value = tmp.Bytes()
				if !isDupeAny(state.prevCacheForDup[field.Name], string(value)) {
					break
				}
			}

			state.prevCacheForDup[field.Name][string(value)] = struct{}{}
			state.prevCacheCardinality[field.Name] = append(state.prevCacheCardinality[field.Name], value)
		}

		idx := int(state.counter % uint64(cardinality))

		// Safety check; should be a noop
		if idx >= len(state.prevCacheCardinality[field.Name]) {
			idx = len(state.prevCacheCardinality[field.Name]) - 1
		}

		choice := state.prevCacheCardinality[field.Name][idx].([]byte)
		buf.Write(choice)
		return nil
	}

	fieldMap[field.Name] = emitFNotReturn
	return nil
}

func makeDynamicStub(boundF any) emitFNotReturn {
	return func(state *GenState, buf *bytes.Buffer) error {
		v := state.pool.Get()
		tmp := v.(*bytes.Buffer)
		tmp.Reset()
		defer state.pool.Put(tmp)

		// Fire the bound function, write into temp buffer
		err := boundF.(emitFNotReturn)(state, tmp)
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

func makeDynamicStubWithReturn(boundF any) EmitF {
	return func(state *GenState) any {
		return boundF.(EmitF)(state)
	}
}

func bindConstantKeywordWithReturn(field Field, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		value, ok := state.prevCache[field.Name].(string)
		if !ok {
			// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
			value = randomdata.Adjective() + randomdata.Noun()
			state.prevCache[field.Name] = value
		}
		return value
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindKeywordWithReturn(fieldCfg ConfigField, field Field, fieldMap map[string]any) error {
	if len(fieldCfg.Enum) > 0 {
		var emitF EmitF
		emitF = func(state *GenState) any {
			idx := rand.Intn(len(fieldCfg.Enum))
			return fieldCfg.Enum[idx]
		}

		fieldMap[field.Name] = emitF
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
		var emitF EmitF
		emitF = func(state *GenState) any {
			// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
			return randomdata.Adjective() + randomdata.Noun()
		}

		fieldMap[field.Name] = emitF
	}
	return nil
}

func bindJoinRandWithReturn(field Field, N int, joiner string, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		value := ""
		for i := 0; i < N-1; i++ {
			value += randomdata.Noun() + joiner
		}

		// randomdata.Adjective() + randomdata.Noun() -> 364 * 527 (~190k) different values
		value += randomdata.Adjective()
		value += randomdata.Noun()

		return value
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindStaticWithReturn(field Field, v any, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		return v
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindBoolWithReturn(field Field, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		switch rand.Int() % 2 {
		case 0:
			return false
		default:
			return true
		}
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindGeoPointWithReturn(field Field, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		return randGeoPointWithReturn()
	}

	fieldMap[field.Name] = emitF

	return nil
}

func bindWordNWithReturn(field Field, n int, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		return genNounsNWithReturn(rand.Intn(n))
	}
	fieldMap[field.Name] = emitF
	return nil
}

func bindNearTimeWithReturn(field Field, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		return newTime
	}
	fieldMap[field.Name] = emitF
	return nil
}

func bindIPWithReturn(field Field, fieldMap map[string]any) error {
	var emitF EmitF
	emitF = func(state *GenState) any {
		i0 := rand.Intn(255)
		i1 := rand.Intn(255)
		i2 := rand.Intn(255)
		i3 := rand.Intn(255)

		return fmt.Sprintf("%d.%d.%d.%d", i0, i1, i2, i3)
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindLongWithReturn(fieldCfg ConfigField, field Field, fieldMap map[string]any) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzzinessNumerator := float64(fieldCfg.Fuzziness.Numerator)
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		var emitF EmitF
		emitF = func(state *GenState) any {
			return dummyFunc()
		}

		fieldMap[field.Name] = emitF
		return nil
	}

	var emitF EmitF
	emitF = func(state *GenState) any {
		dummyInt := dummyFunc()
		if previousDummyInt, ok := state.prevCache[field.Name].(int); ok {
			adjustedRatio := rand.Float64() * (fuzzinessNumerator / fuzzinessDenominator)
			if rand.Int()%2 == 0 {
				adjustedRatio += 1.
			} else {
				adjustedRatio = 1. - adjustedRatio
			}
			dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
		}
		state.prevCache[field.Name] = dummyInt
		return dummyInt
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindDoubleWithReturn(fieldCfg ConfigField, field Field, fieldMap map[string]any) error {

	dummyFunc := makeFloatFunc(fieldCfg, field)

	fuzzinessNumerator := float64(fieldCfg.Fuzziness.Numerator)
	fuzzinessDenominator := float64(fieldCfg.Fuzziness.Denominator)

	if fuzzinessNumerator <= 0 {
		var emitF EmitF
		emitF = func(state *GenState) any {
			return dummyFunc()
		}

		fieldMap[field.Name] = emitF

		return nil
	}

	var emitF EmitF
	emitF = func(state *GenState) any {
		dummyFloat := dummyFunc()
		if previousDummyFloat, ok := state.prevCache[field.Name].(float64); ok {
			adjustedRatio := rand.Float64() * (fuzzinessNumerator / fuzzinessDenominator)
			if rand.Int()%2 == 0 {
				adjustedRatio += 1.
			} else {
				adjustedRatio = 1. - adjustedRatio
			}
			dummyFloat = previousDummyFloat * adjustedRatio
		}
		state.prevCache[field.Name] = dummyFloat
		return dummyFloat
	}

	fieldMap[field.Name] = emitF

	return nil
}

func bindCardinalityWithReturn(cfg Config, field Field, fieldMap map[string]any) error {

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
	boundFWithReturn := fieldMap[field.Name].(EmitF)
	var emitF EmitF
	emitF = func(state *GenState) any {
		var value any
		// Have we rolled over once?  If not, generate a value and cache it.
		if len(state.prevCacheCardinality[field.Name]) < cardinality {
			// Do college try dupe detection on value;
			// Allow dupe if no unique value in nTries.
			nTries := 11 // "These go to 11."
			for i := 0; i < nTries; i++ {
				value = boundFWithReturn(state)

				if !isDupeAny(state.prevCacheForDup[field.Name], value) {
					break
				}
			}

			state.prevCacheForDup[field.Name][value] = struct{}{}
			state.prevCacheCardinality[field.Name] = append(state.prevCacheCardinality[field.Name], value)
		}

		idx := int(state.counter % uint64(cardinality))

		// Safety check; should be a noop
		if idx >= len(state.prevCacheCardinality[field.Name]) {
			idx = len(state.prevCacheCardinality[field.Name]) - 1
		}

		choice := state.prevCacheCardinality[field.Name][idx]

		return choice
	}

	fieldMap[field.Name] = emitF
	return nil
}

func bindObjectWithReturn(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]any) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	if len(fieldCfg.ObjectKeys) > 0 {
		for _, objectsKey := range fieldCfg.ObjectKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap, true); err != nil {
				return err
			}
		}

		return nil
	}

	return bindDynamicObjectWithReturn(cfg, field, fieldMap)
}

func bindDynamicObjectWithReturn(cfg Config, field Field, fieldMap map[string]any) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]any)

	if err := bindField(cfg, field, dynMap, true); err != nil {
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

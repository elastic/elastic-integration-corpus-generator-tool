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

// Typedef of the emit function
type EmitF func(state *GenState) (interface{}, error)

// GenerateFromHero is the helper for hero
func GenerateFromHero(field string, fieldMap map[string]EmitF, state *GenState) interface{} {
	bindF, ok := fieldMap[field]
	if !ok {
		return nil
	}

	value, err := bindF(state)
	if err != nil {
		return nil
	}

	return value
}

type Generator interface {
	Emit(state *GenState, buf *bytes.Buffer) error
	Close() error
}

type GenState struct {
	// event counter
	counter uint64

	// internal buffer pool to decrease load on GC
	dynamicStubPool sync.Pool

	// previous value cache; necessary for fuzziness, cardinality, etc.
	prevCache map[string]interface{}
}

func NewGenState() *GenState {
	return &GenState{
		prevCache: make(map[string]interface{}),
		dynamicStubPool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func (s *GenState) Inc() {
	s.counter += 1
}

// BindField is the exported symbol of bindField to be used in hero
func BindField(cfg Config, field Field, fieldMap map[string]EmitF, objectKeys map[string]struct{}) error {
	return bindField(cfg, field, fieldMap, objectKeys)
}

func bindField(cfg Config, field Field, fieldMap map[string]EmitF, objectKeys map[string]struct{}) error {
	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		return bindStatic(field, fieldCfg.Value, fieldMap)
	}

	// Check for hardcoded field value
	if len(field.Value) > 0 {
		return bindStatic(field, field.Value, fieldMap)
	}

	if fieldCfg.Cardinality > 0 {
		return bindCardinality(cfg, field, fieldMap, objectKeys)
	}

	return bindByType(cfg, field, fieldMap, objectKeys)
}

// Check for dupes O(n)
func isDupe(va []interface{}, dst interface{}) bool {
	var dupe bool
	for _, b := range va {
		if b == dst {
			dupe = true
			break
		}
	}
	return dupe
}

func bindByType(cfg Config, field Field, fieldMap map[string]EmitF, objectKeys map[string]struct{}) (err error) {

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
		err = bindObject(cfg, fieldCfg, field, fieldMap, objectKeys)
	case FieldTypeGeoPoint:
		err = bindGeoPoint(field, fieldMap)
	default:
		err = bindWordN(field, 25, fieldMap)
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

func bindObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]EmitF, objectKeys map[string]struct{}) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	if len(fieldCfg.ObjectKeys) > 0 {
		for _, objectsKey := range fieldCfg.ObjectKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap, objectKeys); err != nil {
				return err
			}
		}

		return nil
	}

	return bindDynamicObject(cfg, field, fieldMap, objectKeys)
}

func bindDynamicObject(cfg Config, field Field, fieldMap map[string]EmitF, objectKeys map[string]struct{}) error {
	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated EmitFunction for use in the stub.
	dynMap := make(map[string]EmitF)

	if err := bindField(cfg, field, dynMap, objectKeys); err != nil {
		return err
	}
	stub := makeDynamicStub(dynMap[field.Name])
	fieldMap[field.Name] = stub

	return nil
}

func genNounsN(n int) string {
	value := ""
	for i := 0; i < n-1; i++ {
		value += randomdata.Noun()
		value += " "
	}

	value += randomdata.Noun()
	return value
}

func randGeoPoint() string {
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

func bindConstantKeyword(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		value, ok := state.prevCache[field.Name].(string)

		if !ok {
			value = randomdata.Noun()
			state.prevCache[field.Name] = value
		}

		return value, nil
	}

	return nil
}

func bindKeyword(fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {
	if len(fieldCfg.Enum) > 0 {
		fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
			idx := rand.Intn(len(fieldCfg.Enum))
			value := fieldCfg.Enum[idx]

			return value, nil
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

		return bindJoinRand(field, totWords, joiner, fieldMap)
	} else {
		fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
			value := randomdata.Noun()

			return value, nil
		}
	}
	return nil
}

func bindJoinRand(field Field, N int, joiner string, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		value := ""
		for i := 0; i < N-1; i++ {
			value += randomdata.Noun()
			value += joiner
		}

		value += randomdata.Noun()

		return value, nil
	}

	return nil
}

func bindStatic(field Field, v interface{}, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		return v, nil
	}

	return nil
}

func bindBool(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		var value bool
		switch rand.Int() % 2 {
		case 0:
			value = false
		case 1:
			value = true
		}

		return value, nil
	}

	return nil
}

func bindGeoPoint(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		value := randGeoPoint()

		return value, nil
	}

	return nil
}

func bindWordN(field Field, n int, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		value := genNounsN(rand.Intn(n))

		return value, nil
	}

	return nil
}

func bindNearTime(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		return newTime, nil
	}

	return nil
}

func bindIP(field Field, fieldMap map[string]EmitF) error {
	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {

		i0 := rand.Intn(255)
		i1 := rand.Intn(255)
		i2 := rand.Intn(255)
		i3 := rand.Intn(255)

		value := fmt.Sprintf("%d.%d.%d.%d", i0, i1, i2, i3)

		return value, nil
	}

	return nil
}

func bindLong(fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

	if fuzziness <= 0 {
		fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
			value := dummyFunc()

			return value, nil
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		dummyInt := dummyFunc()
		if previousDummyInt, ok := state.prevCache[field.Name].(int); ok {
			adjustedRatio := 1. - float64(rand.Intn(fuzziness))/100.
			if rand.Int()%2 == 0 {
				adjustedRatio = 1. + float64(rand.Intn(fuzziness))/100.
			}
			dummyInt = int(math.Ceil(float64(previousDummyInt) * adjustedRatio))
		}
		state.prevCache[field.Name] = dummyInt

		return dummyInt, nil
	}

	return nil
}

func bindDouble(fieldCfg ConfigField, field Field, fieldMap map[string]EmitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

	if fuzziness <= 0 {
		fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
			dummyFloat := float64(dummyFunc()) / rand.Float64()

			return dummyFloat, nil
		}

		return nil
	}

	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		dummyFloat := float64(dummyFunc()) / rand.Float64()
		if previousDummyFloat, ok := state.prevCache[field.Name].(float64); ok {
			adjustedRatio := 1. - float64(rand.Intn(fuzziness))/100.
			if rand.Int()%2 == 0 {
				adjustedRatio = 1. + float64(rand.Intn(fuzziness))/100.
			}
			dummyFloat = previousDummyFloat * adjustedRatio
		}
		state.prevCache[field.Name] = dummyFloat

		return dummyFloat, nil
	}

	return nil
}

func bindCardinality(cfg Config, field Field, fieldMap map[string]EmitF, objectKeys map[string]struct{}) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((1000. / float64(fieldCfg.Cardinality))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByType(cfg, field, fieldMap, objectKeys); err != nil {
		return err
	}

	// We will wrap the function we just generated
	boundF := fieldMap[field.Name]

	fieldMap[field.Name] = func(state *GenState) (interface{}, error) {
		var va []interface{}

		if v, ok := state.prevCache[field.Name]; ok {
			va = v.([]interface{})
		}

		// Have we rolled over once?  If not, generate a value and cache it.
		if len(va) < cardinality {

			// Do college try dupe detection on value;
			// Allow dupe if no unique value in nTries.
			nTries := 11 // "These go to 11."
			var value interface{}
			var err error
			for i := 0; i < nTries; i++ {
				value, err = boundF(state)
				if err != nil {
					return nil, err
				}

				if !isDupe(va, value) {
					break
				}
			}

			va = append(va, value)
			state.prevCache[field.Name] = va
		}

		idx := int(state.counter % uint64(cardinality))

		// Safety check; should be a noop
		if idx >= len(va) {
			idx = len(va) - 1
		}

		choice := va[idx]

		return choice, nil
	}

	return nil

}

func makeDynamicStub(boundF EmitF) EmitF {
	return func(state *GenState) (interface{}, error) {
		// Fire the bound function, write into temp buffer
		value, err := boundF(state)
		if err != nil {
			return nil, err
		}

		return value, nil
	}
}

func unmarshalJSONT[T any](t *testing.T, data []byte) map[string]T {
	m := make(map[string]T)
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

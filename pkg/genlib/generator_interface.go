// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package genlib

import (
	"bytes"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/fields"
	"github.com/lithammer/shortuuid/v3"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"sync"
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
		if len(templateFieldMap) > 0 {
			return bindStaticWithTemplate(templateFieldMap[field.Name], field, field.Value, fieldMap)
		}
		return bindStatic(field, field.Value, fieldMap)
	}

	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		if len(templateFieldMap) > 0 {
			return bindStaticWithTemplate(templateFieldMap[field.Name], field, fieldCfg.Value, fieldMap)
		}
		return bindStatic(field, fieldCfg.Value, fieldMap)
	}

	if fieldCfg.Cardinality > 0 {
		if len(templateFieldMap) > 0 {
			return bindCardinalityWithTemplate(cfg, field, fieldMap, objectKeys, templateFieldMap)
		}
		return bindCardinality(cfg, field, fieldMap)
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

	if len(templateFieldMap) > 0 {
		switch field.Type {
		case FieldTypeDate:
			err = bindNearTimeWithTemplate(templateFieldMap[field.Name], field, fieldMap)
		case FieldTypeIP:
			err = bindIPWithTemplate(templateFieldMap[field.Name], field, fieldMap)
		case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
			err = bindDoubleWithTemplate(templateFieldMap[field.Name], fieldCfg, field, fieldMap)
		case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong: // TODO: generate > 63 bit values for unsigned_long
			err = bindLongWithTemplate(templateFieldMap[field.Name], fieldCfg, field, fieldMap)
		case FieldTypeConstantKeyword:
			err = bindConstantKeywordWithTemplate(templateFieldMap[field.Name], field, fieldMap)
		case FieldTypeKeyword:
			err = bindKeywordWithTemplate(templateFieldMap[field.Name], fieldCfg, field, fieldMap)
		case FieldTypeBool:
			err = bindBoolWithTemplate(templateFieldMap[field.Name], field, fieldMap)
		case FieldTypeObject, FieldTypeNested, FieldTypeFlattened:
			err = bindObject(cfg, fieldCfg, field, fieldMap, objectKeys, templateFieldMap)
		case FieldTypeGeoPoint:
			err = bindGeoPointWithTemplate(templateFieldMap[field.Name], field, fieldMap)
		default:
			err = bindWordNWithTemplate(templateFieldMap[field.Name], field, 25, fieldMap)
		}
	} else {
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
			err = bindObject(cfg, fieldCfg, field, fieldMap, nil, nil)
		case FieldTypeGeoPoint:
			err = bindGeoPoint(field, fieldMap)
		default:
			err = bindWordN(field, 25, fieldMap)
		}
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

	if len(templateFieldMap) > 0 {
		if err := bindField(cfg, field, dynMap, objectKeys, templateFieldMap); err != nil {
			return err
		}
		stub := makeDynamicStubWithTemplate(templateFieldMap[field.Name], dynMap[field.Name])
		fieldMap[field.Name] = stub

		return nil
	}

	objectRootFieldName := replacer.Replace(field.Name)

	for i := 0; i < N; i++ {
		// Generate a guid for binding, we will replace later
		key := shortuuid.New()
		field.Name = key
		if err := bindField(cfg, field, dynMap, nil, nil); err != nil {
			return err
		}
		stub := makeDynamicStub(objectRootFieldName, key, dynMap[key])
		fieldMap[objectRootFieldName+"."+key] = stub
	}

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

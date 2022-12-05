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

func bindField(cfg Config, field Field, fieldMap map[string]emitF, withTemplate bool) error {

	// Check for hardcoded field value
	if len(field.Value) > 0 {
		if withTemplate {
			return bindStaticWithTemplate(field, field.Value, fieldMap)
		}
		return bindStatic(field, field.Value, fieldMap)
	}

	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		if withTemplate {
			return bindStaticWithTemplate(field, fieldCfg.Value, fieldMap)
		}
		return bindStatic(field, fieldCfg.Value, fieldMap)
	}

	if fieldCfg.Cardinality > 0 {
		if withTemplate {
			return bindCardinalityWithTemplate(cfg, field, fieldMap)
		}
		return bindCardinality(cfg, field, fieldMap)
	}

	return bindByType(cfg, field, fieldMap, withTemplate)
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

func bindByType(cfg Config, field Field, fieldMap map[string]emitF, withTemplate bool) (err error) {

	fieldCfg, _ := cfg.GetField(field.Name)

	if withTemplate {
		switch field.Type {
		case FieldTypeDate:
			err = bindNearTimeWithTemplate(field, fieldMap)
		case FieldTypeIP:
			err = bindIPWithTemplate(field, fieldMap)
		case FieldTypeDouble, FieldTypeFloat, FieldTypeHalfFloat, FieldTypeScaledFloat:
			err = bindDoubleWithTemplate(fieldCfg, field, fieldMap)
		case FieldTypeInteger, FieldTypeLong, FieldTypeUnsignedLong: // TODO: generate > 63 bit values for unsigned_long
			err = bindLongWithTemplate(fieldCfg, field, fieldMap)
		case FieldTypeConstantKeyword:
			err = bindConstantKeywordWithTemplate(field, fieldMap)
		case FieldTypeKeyword:
			err = bindKeywordWithTemplate(fieldCfg, field, fieldMap)
		case FieldTypeBool:
			err = bindBoolWithTemplate(field, fieldMap)
		case FieldTypeObject, FieldTypeNested, FieldTypeFlattened:
			err = bindObject(cfg, fieldCfg, field, fieldMap, withTemplate)
		case FieldTypeGeoPoint:
			err = bindGeoPointWithTemplate(field, fieldMap)
		default:
			err = bindWordNWithTemplate(field, 25, fieldMap)
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
			err = bindObject(cfg, fieldCfg, field, fieldMap, withTemplate)
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

func bindObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitF, withTemplate bool) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = FieldTypeKeyword
	}

	objectRootFieldName := replacer.Replace(field.Name)

	objectsKeys := fieldCfg.ObjectKeys
	if len(objectsKeys) > 0 {
		for _, objectsKey := range objectsKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap, withTemplate); err != nil {
				return err
			}
		}

		return nil
	}

	// This is a special case.  We are randomly generating keys on the fly
	// Will creating a special emit function that binds statically,
	// but only fires randomly.
	N := 5
	return bindDynamicObject(cfg, fieldCfg, field, fieldMap, N, withTemplate)
}

func bindDynamicObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitF, N int, withTemplate bool) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]emitF)

	objectRootFieldName := replacer.Replace(field.Name)

	for i := 0; i < N; i++ {
		// Generate a guid for binding, we will replace later
		key := shortuuid.New()

		if withTemplate {
			field.Name = fmt.Sprintf("%s.%d", field.Name, i)
		} else {
			field.Name = key
		}

		if err := bindField(cfg, field, dynMap, withTemplate); err != nil {
			return err
		}

		var stub emitF
		if withTemplate {
			stub = makeDynamicStubWithTemplate(objectRootFieldName, i, dynMap[field.Name])
			fieldMap[fmt.Sprintf("%s.%d", objectRootFieldName, i)] = stub
		} else {
			stub = makeDynamicStub(objectRootFieldName, key, dynMap[key])
			fieldMap[objectRootFieldName+"."+key] = stub
		}

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

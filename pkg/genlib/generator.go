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
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/fields"
	"github.com/lithammer/shortuuid/v3"
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

// Generate is resolved at construction to a slice of emit functions
type Generator struct {
	emitFuncs []emitF
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

func NewGenerator(cfg Config, fields Fields) (*Generator, error) {

	// Preprocess the fields, generating appropriate emit functions
	fieldMap := make(map[string]emitF)
	for _, field := range fields {
		if err := bindField(cfg, field, fieldMap); err != nil {
			return nil, err
		}
	}

	// Roll into slice of emit functions
	emitFuncs := make([]emitF, 0, len(fieldMap))
	for _, f := range fieldMap {
		emitFuncs = append(emitFuncs, f)
	}

	return &Generator{emitFuncs: emitFuncs}, nil

}

func bindField(cfg Config, field Field, fieldMap map[string]emitF) error {

	// Check for hardcoded field value
	if len(field.Value) > 0 {
		return bindStatic(field, field.Value, fieldMap)
	}

	// Check config override of value
	fieldCfg, _ := cfg.GetField(field.Name)
	if fieldCfg.Value != nil {
		return bindStatic(field, fieldCfg.Value, fieldMap)
	}

	if fieldCfg.Cardinality > 0 {
		return bindCardinality(cfg, field, fieldMap)
	}

	return bindByType(cfg, field, fieldMap)
}

func bindCardinality(cfg Config, field Field, fieldMap map[string]emitF) error {

	fieldCfg, _ := cfg.GetField(field.Name)
	cardinality := int(math.Ceil((1000. / float64(fieldCfg.Cardinality))))

	if strings.HasSuffix(field.Name, ".*") {
		field.Name = replacer.Replace(field.Name)
	}

	// Go ahead and bind the original field
	if err := bindByType(cfg, field, fieldMap); err != nil {
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
			var tmp bytes.Buffer
			if err := boundF(state, dupes, &tmp); err != nil {
				return err
			}

			// check for dupes O(n)
			isDupe := false
			for _, b := range va {
				if bytes.Equal(tmp.Bytes(), b.Bytes()) {
					isDupe = true
					break
				}
			}

			// ok, we succeeed, append the value to our cache
			if !isDupe {
				state.prevCache[field.Name] = append(va, tmp)
			}

			buf.Write(tmp.Bytes())
			return nil
		}

		choice := va[rand.Intn(len(va))]
		buf.Write(choice.Bytes())
		return nil
	}

	return nil

}

func bindByType(cfg Config, field Field, fieldMap map[string]emitF) (err error) {

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
		err = bindKeyword(field, fieldMap)
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

func bindConstantKeyword(field Field, fieldMap map[string]emitF) error {

	prefix := fmt.Sprintf("\"%s\":\"", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		value, ok := state.prevCache[field.Name].(string)
		if !ok {
			value = randomdata.Noun()
			state.prevCache[field.Name] = value
		}
		buf.WriteString(prefix)
		buf.WriteString(value)
		buf.WriteByte('"')
		return nil
	}

	return nil
}

func bindKeyword(field Field, fieldMap map[string]emitF) error {

	if len(field.Example) > 0 {

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
		prefix := fmt.Sprintf("\"%s\":\"", field.Name)

		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			value := randomdata.Noun()
			buf.WriteString(prefix)
			buf.WriteString(value)
			buf.WriteByte('"')
			return nil
		}
	}
	return nil
}

func bindJoinRand(field Field, N int, joiner string, fieldMap map[string]emitF) error {

	prefix := fmt.Sprintf("\"%s\":\"", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {

		buf.WriteString(prefix)

		for i := 0; i < N-1; i++ {
			buf.WriteString(randomdata.Noun())
			buf.WriteString(joiner)
		}
		buf.WriteString(randomdata.Noun())
		buf.WriteByte('"')
		return nil
	}

	return nil
}

func bindStatic(field Field, v interface{}, fieldMap map[string]emitF) error {

	vstr, err := json.Marshal(v)
	if err != nil {
		return err
	}

	payload := fmt.Sprintf("\"%s\":%s", field.Name, vstr)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.WriteString(payload)
		return nil
	}

	return nil
}

func bindBool(field Field, fieldMap map[string]emitF) error {

	prefix := fmt.Sprintf("\"%s\":", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.WriteString(prefix)
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

func bindGeoPoint(field Field, fieldMap map[string]emitF) error {

	prefix := fmt.Sprintf("\"%s\":\"", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.WriteString(prefix)
		err := randGeoPoint(buf)
		buf.WriteByte('"')
		return err
	}

	return nil
}

func bindWordN(field Field, n int, fieldMap map[string]emitF) error {
	prefix := fmt.Sprintf("\"%s\":\"", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		buf.WriteString(prefix)
		genNounsN(rand.Intn(n), buf)
		buf.WriteByte('"')
		return nil
	}

	return nil
}

func bindNearTime(field Field, fieldMap map[string]emitF) error {
	prefix := fmt.Sprintf("\"%s\":\"", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		offset := time.Duration(rand.Intn(FieldTypeTimeRange)*-1) * time.Second
		newTime := time.Now().Add(offset)

		buf.WriteString(prefix)
		buf.WriteString(newTime.Format(FieldTypeTimeLayout))
		buf.WriteByte('"')
		return nil
	}

	return nil
}

func bindIP(field Field, fieldMap map[string]emitF) error {
	prefix := fmt.Sprintf("\"%s\":", field.Name)

	fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {

		buf.WriteString(prefix)

		i0 := rand.Intn(255)
		i1 := rand.Intn(255)
		i2 := rand.Intn(255)
		i3 := rand.Intn(255)

		_, err := fmt.Fprintf(buf, "\"%d.%d.%d.%d\"", i0, i1, i2, i3)
		return err
	}

	return nil
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

func bindLong(fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

	prefix := fmt.Sprintf("\"%s\":", field.Name)

	if fuzziness <= 0 {
		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			buf.WriteString(prefix)
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
		buf.WriteString(prefix)
		v := make([]byte, 0, 32)
		v = strconv.AppendInt(v, int64(dummyInt), 10)
		buf.Write(v)
		return nil
	}

	return nil
}

func bindDouble(fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {

	dummyFunc := makeIntFunc(fieldCfg, field)

	fuzziness := fieldCfg.Fuzziness

	prefix := fmt.Sprintf("\"%s\":", field.Name)

	if fuzziness <= 0 {
		fieldMap[field.Name] = func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
			dummyFloat := float64(dummyFunc()) / rand.Float64()
			buf.WriteString(prefix)
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
		buf.WriteString(prefix)
		_, err := fmt.Fprintf(buf, "%f", dummyFloat)
		return err
	}

	return nil
}

func bindObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitF) error {
	if len(field.ObjectType) > 0 {
		field.Type = field.ObjectType
	} else {
		field.Type = "keyword"
	}

	objectRootFieldName := replacer.Replace(field.Name)

	objectsKeys := fieldCfg.ObjectKeys
	if len(objectsKeys) > 0 {
		for _, objectsKey := range objectsKeys {
			field.Name = objectRootFieldName + "." + objectsKey

			if err := bindField(cfg, field, fieldMap); err != nil {
				return err
			}
		}

		return nil
	}

	// This is a special case.  We are randomly generating keys on the fly
	// Will creating a special emit function that binds statically,
	// but only fires randomly.
	N := 5
	return bindDynamicObject(cfg, fieldCfg, field, fieldMap, N)
}

func bindDynamicObject(cfg Config, fieldCfg ConfigField, field Field, fieldMap map[string]emitF, N int) error {

	// Temporary fieldMap which we pass to the bind function,
	// then extract the generated emitFunction for use in the stub.
	dynMap := make(map[string]emitF)

	objectRootFieldName := replacer.Replace(field.Name)

	for i := 0; i < N; i++ {
		// Generate a guid for binding, we will replace later
		key := shortuuid.New()
		field.Name = key

		if err := bindField(cfg, field, dynMap); err != nil {
			return err
		}

		stub := makeDynamicStub(objectRootFieldName, key, dynMap[key])
		fieldMap[objectRootFieldName+"."+key] = stub
	}

	return nil
}

func makeDynamicStub(root, key string, boundF emitF) emitF {
	target := fmt.Sprintf("\"%s\":", key)

	return func(state *GenState, dupes map[string]struct{}, buf *bytes.Buffer) error {
		// Fire or skip
		if rand.Int()%2 == 0 {
			return nil
		}

		v := state.pool.Get()
		tmp := v.(*bytes.Buffer)
		tmp.Reset()
		defer state.pool.Put(tmp)

		// Fire the bound function, write into temp buffer
		if err := boundF(state, dupes, tmp); err != nil {
			return err
		}

		// If bound function did not write for some reason; abort
		if tmp.Len() <= len(target) {
			return nil
		}

		if !bytes.HasPrefix(tmp.Bytes(), []byte(target)) {
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

		// ok, formatted as expected, swap it out the payload
		buf.WriteByte('"')
		buf.WriteString(root)
		buf.WriteByte('.')
		buf.WriteString(rNoun)
		buf.WriteString("\":")
		buf.Write(tmp.Bytes()[len(target):])
		return nil
	}
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

func (gen Generator) Emit(state *GenState, buf *bytes.Buffer) error {

	buf.WriteByte('{')

	if err := gen.emit(state, buf); err != nil {
		return err
	}

	buf.WriteByte('}')

	state.counter += 1

	return nil
}

func (gen Generator) emit(state *GenState, buf *bytes.Buffer) error {

	dupes := make(map[string]struct{})

	lastComma := -1
	for _, f := range gen.emitFuncs {
		pos := buf.Len()
		if err := f(state, dupes, buf); err != nil {
			return err
		}

		// If we emitted something, write the comma, otherwise skip.
		if buf.Len() > pos {
			buf.WriteByte(',')
			lastComma = buf.Len()
		}
	}

	// Strip dangling comma
	if lastComma == buf.Len() {
		buf.Truncate(buf.Len() - 1)
	}

	return nil
}

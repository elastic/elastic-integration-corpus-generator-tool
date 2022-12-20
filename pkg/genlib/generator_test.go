package genlib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/fields"
)

/*
const cardinalityCfg = `
- name: event.id
  cardinality: 250
- name: process.pid
  fuzziness: 10
  range: 100
`
*/

func Test_ParseTemplate(t *testing.T) {
	testCases := []struct {
		template                  []byte
		expectedOrderFields       []string
		expectedTemplateFieldsMap map[string][]byte
		expectedTrailingTemplate  []byte
	}{
		{
			template:                  []byte("no field"),
			expectedOrderFields:       []string{},
			expectedTemplateFieldsMap: map[string][]byte{},
			expectedTrailingTemplate:  []byte("no field"),
		},
		{
			template:                  []byte("{{.aField}}"),
			expectedOrderFields:       []string{"aField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": nil},
			expectedTrailingTemplate:  nil,
		},
		{
			template:                  []byte("{{.aField}} {{.anotherField}}"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": nil, "anotherField": []byte(" ")},
			expectedTrailingTemplate:  nil,
		},
		{
			template:                  []byte("with prefix {{.aField}} {{.anotherField}}"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("with prefix"), "anotherField": []byte(" ")},
			expectedTrailingTemplate:  nil,
		},
		{
			template:                  []byte("{{.aField}} {{.anotherField}} with trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": nil, "anotherField": []byte(" ")},
			expectedTrailingTemplate:  []byte(" with trailing"),
		},
		{
			template:                  []byte("with prefix {{.aField}} {{.anotherField}} and trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("with prefix"), "anotherField": []byte(" ")},
			expectedTrailingTemplate:  []byte(" and trailing"),
		},
		{
			template:                  []byte("{{.aField}} with { in the middle {{.anotherField}}"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": nil, "anotherField": []byte(" with { in the middle ")},
			expectedTrailingTemplate:  nil,
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} {{.anotherField}}"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" ")},
			expectedTrailingTemplate:  nil,
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} and { in the middle {{.anotherField}}"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" and { in the middle ")},
			expectedTrailingTemplate:  nil,
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} {{.anotherField}} and trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" ")},
			expectedTrailingTemplate:  []byte(" and trailing"),
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} and { in the middle {{.anotherField}} and trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" and { in the middle ")},
			expectedTrailingTemplate:  []byte(" and trailing"),
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} {{.anotherField}} and { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing"),
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} and { in the middle {{.anotherField}} and { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" and { in the middle ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing"),
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} {{.anotherField}} and { curly brace in trailing with again { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing with again { curly brace in trailing"),
		},
		{
			template:                  []byte("{ with curly brace as prefix {{.aField}} and { in the middle {{.anotherField}} and { curly brace in trailing with again { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{ with curly brace as prefix "), "anotherField": []byte(" and { in the middle ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing with again { curly brace in trailing"),
		},
		{
			template:                  []byte("{{{.aField}} with curly brace as prefix just before a field {{.anotherField}} and trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{"), "anotherField": []byte(" with curly brace as prefix just before a field ")},
			expectedTrailingTemplate:  []byte(" and trailing"),
		},
		{
			template:                  []byte("{{{.aField}} with curly brace as prefix just before a field and { in the middle {{.anotherField}} and trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{"), "anotherField": []byte(" with curly brace as prefix just before a field and { in the middle ")},
			expectedTrailingTemplate:  []byte(" and trailing"),
		},
		{
			template:                  []byte("{{{.aField}} with curly brace as prefix just before a field {{.anotherField}} and { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{"), "anotherField": []byte(" with curly brace as prefix just before a field ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing"),
		},
		{
			template:                  []byte("{{{.aField}} with curly brace as prefix just before a field and { in the middle {{.anotherField}} and { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{"), "anotherField": []byte(" with curly brace as prefix just before a field and { in the middle ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing"),
		},
		{
			template:                  []byte("{{{.aField}} with curly brace as prefix just before a field {{.anotherField}} and { curly brace in trailing with again { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{"), "anotherField": []byte(" with curly brace as prefix just before a field ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing with again { curly brace in trailing"),
		},
		{
			template:                  []byte("{{{.aField}} with curly brace as prefix just before a field and { in the middle {{.anotherField}} and { curly brace in trailing with again { curly brace in trailing"),
			expectedOrderFields:       []string{"aField", "anotherField"},
			expectedTemplateFieldsMap: map[string][]byte{"aField": []byte("{"), "anotherField": []byte(" with curly brace as prefix just before a field and { in the middle ")},
			expectedTrailingTemplate:  []byte(" and { curly brace in trailing with again { curly brace in trailing"),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			orderedFields, templateFieldsMap, trailingTemplate := parseTemplate(testCase.template)
			if len(orderedFields) != len(testCase.expectedOrderFields) {
				t.Errorf("Expected equal orderedFields")
			}

			for i := range orderedFields {
				if orderedFields[i] != testCase.expectedOrderFields[i] {
					t.Errorf("Expected ordered field at position %d is wrong (expected: `%s`, given: `%s`", i, testCase.expectedOrderFields[i], orderedFields[i])
				}
			}

			if len(templateFieldsMap) != len(testCase.expectedTemplateFieldsMap) {
				t.Errorf("Expected equal templateFieldsMap")
			}

			for k := range templateFieldsMap {
				if _, ok := testCase.expectedTemplateFieldsMap[k]; !ok {
					t.Errorf("Missing expected field `%s` in templateFieldsMap", k)
				}
			}

			if string(trailingTemplate) != string(testCase.expectedTrailingTemplate) {
				t.Errorf("Expected trailing template is wrong (expected: `%s`, given: `%s`", testCase.expectedTrailingTemplate, trailingTemplate)
			}
		})
	}
}

func Test_EmptyCase(t *testing.T) {
	template, _ := generateTemplateFromField(Config{}, Fields{})
	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: template,
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			g, state := makeGenerator(t, Config{}, []Field{}, testCase.template)

			var buf bytes.Buffer

			if err := g.Emit(state, &buf); err != nil {
				t.Fatal(err)
			}

			if len(buf.Bytes()) != 0 {
				t.Errorf("Expected empty bytes")
			}
		})
	}
}

func Test_Cardinality(t *testing.T) {

	test_CardinalityT[string](t, FieldTypeKeyword)
	test_CardinalityT[int](t, FieldTypeInteger)
	test_CardinalityT[float64](t, FieldTypeFloat)
	test_CardinalityT[string](t, FieldTypeGeoPoint)
	test_CardinalityT[string](t, FieldTypeIP)
	test_CardinalityT[string](t, FieldTypeDate)
}

func test_CardinalityT[T any](t *testing.T, ty string) {
	template := `{"alpha":"{{.alpha}}"}`
	if ty == "integer" || ty == "float" {
		template = `{"alpha":{{.alpha}}}`
	}
	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(template),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("for type %s, with template: %s", ty, string(testCase.template)), func(t *testing.T) {
			// It's cardinality per mille, so a bit confusing :shrug:
			for cardinality := 1000; cardinality >= 10; cardinality /= 10 {

				// Add the range to get some variety in integers
				tmpl := "- name: alpha\n  cardinality: %d\n  range: 10000"
				yaml := []byte(fmt.Sprintf(tmpl, cardinality))

				cfg, err := config.LoadConfigFromYaml(yaml)
				if err != nil {
					t.Fatal(err)
				}

				g, state := makeGenerator(t, cfg, []Field{fld}, testCase.template)

				vmap := make(map[any]int)

				nSpins := 16384
				for i := 0; i < nSpins; i++ {

					var buf bytes.Buffer
					if err := g.Emit(state, &buf); err != nil {
						t.Fatal(err)
					}

					m := unmarshalJSONT[T](t, buf.Bytes())

					if len(m) != 1 {
						t.Errorf("Expected map size 1, got %d", len(m))
					}

					v, ok := m[fld.Name]

					if !ok {
						t.Errorf("Missing key %v", fld.Name)
					}

					vmap[v] = vmap[v] + 1
				}

				if len(vmap) != 1000/cardinality {
					t.Errorf("Expected cardinality of %d got %d", 1000/cardinality, len(vmap))
				}
			}
		})
	}
}

func Test_FieldBool(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeBool,
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":{{.alpha}}}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			// Enough spins, so we can make sure we get at least one true and at least one false
			var cntTrue int
			nSpins := 1024
			for i := 0; i < nSpins; i++ {
				b := testSingleT[bool](t, fld, nil, testCase.template)

				if b {
					cntTrue += 1
				}
			}

			if cntTrue == 0 {
				t.Errorf("No true values, really?")
			}

			if cntTrue == nSpins {
				t.Errorf("No false values, really?")
			}
		})
	}
}

func Test_FieldConstKeyword(t *testing.T) {

	fld := Field{
		Name:  "alpha",
		Type:  FieldTypeConstantKeyword,
		Value: "constant_keyword",
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":{{.alpha}}}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			b := testSingleT[string](t, fld, nil, testCase.template)
			if b != fld.Value {
				t.Errorf("static value not match")
			}
		})
	}
}

func Test_FieldStaticOverrideString(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: beta")
	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":{{.alpha}}}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			b := testSingleT[string](t, fld, yaml, testCase.template)
			if b != "beta" {
				t.Errorf("static value not match")
			}
		})
	}
}

func Test_FieldStaticOverrideNumeric(t *testing.T) {
	fld := Field{

		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: 33")
	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":{{.alpha}}}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			b := testSingleT[float64](t, fld, yaml, testCase.template)

			if b != 33.0 {
				t.Errorf("static value not match")
			}
		})
	}
}

func Test_FieldStaticOverrideBool(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: true")
	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":{{.alpha}}}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			b := testSingleT[bool](t, fld, yaml, testCase.template)

			if b != true {
				t.Errorf("static value not match")
			}
		})
	}
}

func Test_FieldGeoPoint(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeGeoPoint,
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":"{{.alpha}}"}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			nSpins := 1024
			for i := 0; i < nSpins; i++ {

				b := testSingleT[string](t, fld, nil, testCase.template)

				// Expect geo point in form of lat,long
				// where lat is [-90.0..90.0]
				// and long is  [-180.0..180.0]

				s := strings.Split(b, ",")
				if len(s) != 2 {
					t.Fatal("expected comma separated lat,long")
				}

				lat := s[0]
				long := s[1]

				// no whitespace please
				if len(lat) != len(strings.TrimSpace(lat)) {
					t.Errorf("extra whitespace on latitude %s", lat)
				}

				// no whitespace please
				if len(long) != len(strings.TrimSpace(long)) {
					t.Errorf("extra whitespace on longitude %s", long)
				}

				latF, err := strconv.ParseFloat(lat, 64)
				if err != nil {
					t.Errorf("Fail parse latitude as float")
				}
				longF, err := strconv.ParseFloat(long, 64)
				if err != nil {
					t.Errorf("Fail parse longitude as float")
				}

				if latF < -90.0 || latF > 90.0 {
					t.Errorf("latitude out of range %v", latF)
				}

				if longF < -180.0 || longF > 180.0 {
					t.Errorf("longitutde out of range %v", longF)
				}
			}
		})
	}
}

func Test_FieldDate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeDate,
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":"{{.alpha}}"}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			nSpins := rand.Intn(1024) + 1
			for i := 0; i < nSpins; i++ {
				now := time.Now()

				b := testSingleT[string](t, fld, nil, testCase.template)

				if ts, err := time.Parse(FieldTypeTimeLayout, b); err != nil {
					t.Errorf("Fail parse timestamp %v", err)
				} else {
					// Timestamp should be +- FieldTypeDurationSpan from now within a second of slop
					ts.Add(time.Second * -1)
					ts.Add(time.Second)

					diff := ts.Sub(now)
					if diff < 0 {
						diff = -diff
					}

					if diff >= FieldTypeTimeRange*time.Second {
						t.Errorf("Date generated out of span range %v", diff)
					}
				}
			}
		})
	}
}

func Test_FieldIP(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeIP,
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":"{{.alpha}}"}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			nSpins := rand.Intn(1024) + 1
			for i := 0; i < nSpins; i++ {

				b := testSingleT[string](t, fld, nil, testCase.template)

				if ip := net.ParseIP(b); ip == nil {
					t.Errorf("Fail parse ip %s", b)
				}
			}
		})
	}
}

func Test_FieldFloats(t *testing.T) {
	_testNumeric[float64](t, FieldTypeDouble)
	_testNumeric[float32](t, FieldTypeFloat)
	_testNumeric[float32](t, FieldTypeHalfFloat)
	_testNumeric[float64](t, FieldTypeScaledFloat)

}

func Test_FieldIntegers(t *testing.T) {
	_testNumeric[int](t, FieldTypeInteger)
	_testNumeric[int64](t, FieldTypeLong)
	_testNumeric[uint64](t, FieldTypeUnsignedLong)
}

func _testNumeric[T any](t *testing.T, ty string) {
	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	testCases := []struct {
		template []byte
	}{
		{
			template: nil,
		},
		{
			template: []byte(`{"alpha":{{.alpha}}}`),
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("with template: %s", string(testCase.template)), func(t *testing.T) {
			nSpins := rand.Intn(1024) + 1
			for i := 0; i < nSpins; i++ {
				testSingleT[T](t, fld, nil, testCase.template)
			}
		})
	}
}

func testSingleT[T any](t *testing.T, fld Field, yaml []byte, template []byte) T {
	var err error
	var cfg Config

	if yaml != nil {
		cfg, err = config.LoadConfigFromYaml(yaml)
		if err != nil {
			t.Fatal(err)
		}
	}

	g, state := makeGenerator(t, cfg, []Field{fld}, template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	// Buffer should now contain an event shaped like {"alpha": "constant_keyword"}
	m := unmarshalJSONT[T](t, buf.Bytes())

	if len(m) != 1 {
		t.Errorf("Expected map size 1, got %d", len(m))
	}

	v, ok := m[fld.Name]

	if !ok {
		t.Errorf("Missing key %v", fld.Name)

	}

	return v
}

func unmarshalJSONT[T any](t *testing.T, data []byte) map[string]T {
	m := make(map[string]T)
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func makeGenerator(t *testing.T, cfg Config, fields Fields, template []byte) (Generator, *GenState) {
	var g Generator
	var err error
	if len(template) == 0 {
		g, err = NewGenerator(cfg, fields)
	} else {
		g, err = NewGeneratorWithTemplate(template, cfg, fields)
	}

	if err != nil {
		t.Fatal(err)
	}

	return g, NewGenState()
}

func Benchmark_Generator(b *testing.B) {
	ctx := context.Background()
	flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "endpoint", "process", "8.2.0")

	if err != nil {
		b.Fatal(err)
	}

	g, err := NewGenerator(Config{}, flds)

	if err != nil {
		b.Fatal(err)
	}

	var buf bytes.Buffer

	state := NewGenState()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := g.Emit(state, &buf)
		if err != nil {
			b.Fatal(err)
		}
		buf.Reset()
	}
}

func Benchmark_GeneratorWithTemplate(b *testing.B) {
	ctx := context.Background()
	flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "endpoint", "process", "8.2.0")

	if err != nil {
		b.Fatal(err)
	}

	template, objectKeysField := generateTemplateFromField(Config{}, flds)
	flds = append(flds, objectKeysField...)
	g, err := NewGeneratorWithTemplate(template, Config{}, flds)

	if err != nil {
		b.Fatal(err)
	}

	var buf bytes.Buffer

	state := NewGenState()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := g.Emit(state, &buf)
		if err != nil {
			b.Fatal(err)
		}
		buf.Reset()
	}
}

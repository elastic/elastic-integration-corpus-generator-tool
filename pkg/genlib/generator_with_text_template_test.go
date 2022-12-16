package genlib

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
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

func Test_EmptyCaseWithTemplate(t *testing.T) {
	template := generateTextTemplateFromField(Config{}, []Field{})
	t.Logf("with template: %s", string(template))
	g, state := makeGeneratorWithTemplate(t, Config{}, []Field{}, template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	if len(buf.Bytes()) != 0 {
		t.Errorf("Expected empty bytes")
	}
}

func Test_CardinalityWithTemplate(t *testing.T) {

	test_CardinalityTWithTemplate[string](t, FieldTypeKeyword)
	test_CardinalityTWithTemplate[int](t, FieldTypeInteger)
	test_CardinalityTWithTemplate[float64](t, FieldTypeFloat)
	test_CardinalityTWithTemplate[string](t, FieldTypeGeoPoint)
	test_CardinalityTWithTemplate[string](t, FieldTypeIP)
	test_CardinalityTWithTemplate[string](t, FieldTypeDate)
}

func test_CardinalityTWithTemplate[T any](t *testing.T, ty string) {
	template := []byte(`{"alpha":"{{generate "alpha"}}"}`)
	if ty == "integer" || ty == "float" {
		template = []byte(`{"alpha":{{generate "alpha"}}}`)
	}
	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	t.Logf("for type %s, with template: %s", ty, string(template))
	// It's cardinality per mille, so a bit confusing :shrug:
	for cardinality := 1000; cardinality >= 10; cardinality /= 10 {

		// Add the range to get some variety in integers
		tmpl := "- name: alpha\n  cardinality: %d\n  range: 10000"
		yaml := []byte(fmt.Sprintf(tmpl, cardinality))

		cfg, err := config.LoadConfigFromYaml(yaml)
		if err != nil {
			t.Fatal(err)
		}

		g, state := makeGeneratorWithTemplate(t, cfg, []Field{fld}, template)

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
}

func Test_FieldBoolWithTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeBool,
	}

	template := []byte(`{"alpha":{{generate "alpha"}}}`)
	t.Logf("with template: %s", string(template))
	// Enough spins, so we can make sure we get at least one true and at least one false
	var cntTrue int
	nSpins := 1024
	for i := 0; i < nSpins; i++ {
		b := testSingleTWithTextTemplate[bool](t, fld, nil, template)

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
}

func Test_FieldConstKeywordWithTemplate(t *testing.T) {

	fld := Field{
		Name:  "alpha",
		Type:  FieldTypeConstantKeyword,
		Value: "constant_keyword",
	}

	template := []byte(`{"alpha":"{{generate "alpha"}}"}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[string](t, fld, nil, template)
	if b != fld.Value {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideStringWithTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: beta")
	template := []byte(`{"alpha":"{{generate "alpha"}}"}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[string](t, fld, yaml, template)
	if b != "beta" {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideNumericWithTemplate(t *testing.T) {
	fld := Field{

		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: 33")
	template := []byte(`{"alpha":{{generate "alpha"}}}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[float64](t, fld, yaml, template)

	if b != 33.0 {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideBoolWithTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: true")
	template := []byte(`{"alpha":{{generate "alpha"}}}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[bool](t, fld, yaml, template)

	if b != true {
		t.Errorf("static value not match")
	}
}

func Test_FieldGeoPointWithTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeGeoPoint,
	}

	template := []byte(`{"alpha":"{{generate "alpha"}}"}`)
	t.Logf("with template: %s", string(template))
	nSpins := 1024
	for i := 0; i < nSpins; i++ {

		b := testSingleTWithTextTemplate[string](t, fld, nil, template)

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
}

func Test_FieldDateWithTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeDate,
	}

	template := []byte(`{{$alpha := generate "alpha"}}{"alpha":"{{$alpha.Format "2006-01-02T15:04:05.999999Z07:00"}}"}`)
	t.Logf("with template: %s", string(template))
	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {
		now := time.Now()

		b := testSingleTWithTextTemplate[string](t, fld, nil, template)

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
}

func Test_FieldIPWithTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeIP,
	}

	template := []byte(`{"alpha":"{{generate "alpha"}}"}`)
	t.Logf("with template: %s", string(template))
	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {

		b := testSingleTWithTextTemplate[string](t, fld, nil, template)

		if ip := net.ParseIP(b); ip == nil {
			t.Errorf("Fail parse ip %s", b)
		}
	}
}

func Test_FieldFloatsWithTemplate(t *testing.T) {
	_testNumericWithTextTemplate[float64](t, FieldTypeDouble)
	_testNumericWithTextTemplate[float32](t, FieldTypeFloat)
	_testNumericWithTextTemplate[float32](t, FieldTypeHalfFloat)
	_testNumericWithTextTemplate[float64](t, FieldTypeScaledFloat)

}

func Test_FieldIntegersWithTemplate(t *testing.T) {
	_testNumericWithTextTemplate[int](t, FieldTypeInteger)
	_testNumericWithTextTemplate[int64](t, FieldTypeLong)
	_testNumericWithTextTemplate[uint64](t, FieldTypeUnsignedLong)
}

func _testNumericWithTextTemplate[T any](t *testing.T, ty string) {
	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	template := []byte(`{"alpha":{{generate "alpha"}}}`)
	t.Logf("with template: %s", string(template))
	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {
		testSingleTWithTextTemplate[T](t, fld, nil, template)
	}
}

func testSingleTWithTextTemplate[T any](t *testing.T, fld Field, yaml []byte, template []byte) T {
	var err error
	var cfg Config

	if yaml != nil {
		cfg, err = config.LoadConfigFromYaml(yaml)
		if err != nil {
			t.Fatal(err)
		}
	}

	g, state := makeGeneratorWithTemplate(t, cfg, []Field{fld}, template)

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

func makeGeneratorWithTemplate(t *testing.T, cfg Config, fields Fields, template []byte) (Generator, *GenState) {
	g, err := NewGeneratorWithTemplate(template, cfg, fields)

	if err != nil {
		t.Fatal(err)
	}

	return g, NewGenState()
}

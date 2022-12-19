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

func Test_EmptyCaseWithJetHTML(t *testing.T) {
	template := generateJetTemplateFromField(Config{}, []Field{})
	t.Logf("with template: %s", string(template))
	g, state := makeGeneratorWithJetHTML(t, Config{}, []Field{}, template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	if len(buf.Bytes()) != 0 {
		t.Errorf("Expected empty bytes")
	}
}

func Test_CardinalityWithJetHTML(t *testing.T) {

	test_CardinalityT[string](t, FieldTypeKeyword)
	test_CardinalityT[int](t, FieldTypeInteger)
	test_CardinalityT[float64](t, FieldTypeFloat)
	test_CardinalityT[string](t, FieldTypeGeoPoint)
	test_CardinalityT[string](t, FieldTypeIP)
	test_CardinalityT[string](t, FieldTypeDate)
}

func test_CardinalityT[T any](t *testing.T, ty string) {
	template := []byte(`{"alpha":"{{ "alpha"|generate }}"}`)
	if ty == "integer" || ty == "float" {
		template = []byte(`{"alpha":{{ "alpha"|generate }}}`)
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

		g, state := makeGeneratorWithJetHTML(t, cfg, []Field{fld}, template)

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

func Test_FieldBoolWithJetHTML(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeBool,
	}

	template := []byte(`{"alpha":{{ "alpha"|generate }}}`)
	t.Logf("with template: %s", string(template))
	// Enough spins, so we can make sure we get at least one true and at least one false
	var cntTrue int
	nSpins := 1024
	for i := 0; i < nSpins; i++ {
		b := testSingleTWithJetHTML[bool](t, fld, nil, template)

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

func Test_FieldConstKeywordWithJetHTML(t *testing.T) {

	fld := Field{
		Name:  "alpha",
		Type:  FieldTypeConstantKeyword,
		Value: "constant_keyword",
	}

	template := []byte(`{"alpha":"{{ "alpha"|generate }}"}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithJetHTML[string](t, fld, nil, template)
	if b != fld.Value {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideStringWithJetHTML(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: beta")
	template := []byte(`{"alpha":"{{ "alpha"|generate }}"}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithJetHTML[string](t, fld, yaml, template)
	if b != "beta" {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideNumericWithJetHTML(t *testing.T) {
	fld := Field{

		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: 33")
	template := []byte(`{"alpha":{{ "alpha"|generate }}}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithJetHTML[float64](t, fld, yaml, template)

	if b != 33.0 {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideBoolWithJetHTML(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("- name: alpha\n  value: true")
	template := []byte(`{"alpha":{{ "alpha"|generate }}}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithJetHTML[bool](t, fld, yaml, template)

	if b != true {
		t.Errorf("static value not match")
	}
}

func Test_FieldGeoPointWithJetHTML(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeGeoPoint,
	}

	template := []byte(`{"alpha":"{{ "alpha"|generate }}"}`)
	t.Logf("with template: %s", string(template))
	nSpins := 1024
	for i := 0; i < nSpins; i++ {

		b := testSingleTWithJetHTML[string](t, fld, nil, template)

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

func Test_FieldDateWithJetHTML(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeDate,
	}

	template := []byte(`{{ alpha := generate("alpha") }}{"alpha":"{{ alpha.Format: "2006-01-02T15:04:05.999999Z07:00" }}"}`)
	t.Logf("with template: %s", string(template))
	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {
		now := time.Now()

		b := testSingleTWithJetHTML[string](t, fld, nil, template)

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

func Test_FieldIPWithJetHTML(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeIP,
	}

	template := []byte(`{"alpha":"{{ "alpha"|generate }}"}`)
	t.Logf("with template: %s", string(template))
	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {

		b := testSingleTWithJetHTML[string](t, fld, nil, template)

		if ip := net.ParseIP(b); ip == nil {
			t.Errorf("Fail parse ip %s", b)
		}
	}
}

func Test_FieldFloatsWithJetHTML(t *testing.T) {
	_testNumericWithJetHTML[float64](t, FieldTypeDouble)
	_testNumericWithJetHTML[float32](t, FieldTypeFloat)
	_testNumericWithJetHTML[float32](t, FieldTypeHalfFloat)
	_testNumericWithJetHTML[float64](t, FieldTypeScaledFloat)

}

func Test_FieldIntegersWithJetHTML(t *testing.T) {
	_testNumericWithJetHTML[int](t, FieldTypeInteger)
	_testNumericWithJetHTML[int64](t, FieldTypeLong)
	_testNumericWithJetHTML[uint64](t, FieldTypeUnsignedLong)
}

func _testNumericWithJetHTML[T any](t *testing.T, ty string) {
	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	template := []byte(`{"alpha":{{ "alpha"|generate }}}`)
	t.Logf("with template: %s", string(template))
	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {
		testSingleTWithJetHTML[T](t, fld, nil, template)
	}
}

func testSingleTWithJetHTML[T any](t *testing.T, fld Field, yaml []byte, template []byte) T {
	var err error
	var cfg Config

	if yaml != nil {
		cfg, err = config.LoadConfigFromYaml(yaml)
		if err != nil {
			t.Fatal(err)
		}
	}

	g, state := makeGeneratorWithJetHTML(t, cfg, []Field{fld}, template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	// BufferWithMutex should now contain an event shaped like {"alpha": "constant_keyword"}
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

func makeGeneratorWithJetHTML(t *testing.T, cfg Config, fields Fields, template []byte) (Generator, *GenState) {
	g, err := NewGeneratorWithJetHTML(template, cfg, fields)

	if err != nil {
		t.Fatal(err)
	}

	return g, NewGenState()
}

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

func Test_EmptyCaseWithTextTemplate(t *testing.T) {
	template, _ := generateTextTemplateFromField(Config{}, []Field{})
	t.Logf("with template: %s", string(template))
	g := makeGeneratorWithTextTemplate(t, Config{}, []Field{}, template, 0)

	var buf bytes.Buffer

	if err := g.Emit(&buf); err != nil {
		t.Fatal(err)
	}

	if len(buf.Bytes()) != 0 {
		t.Errorf("Expected empty bytes")
	}
}

func Test_CardinalityWithTextTemplate(t *testing.T) {

	test_CardinalityTWithTextTemplate[string](t, FieldTypeKeyword)
	test_CardinalityTWithTextTemplate[int](t, FieldTypeInteger)
	test_CardinalityTWithTextTemplate[float64](t, FieldTypeFloat)
	test_CardinalityTWithTextTemplate[string](t, FieldTypeGeoPoint)
	test_CardinalityTWithTextTemplate[string](t, FieldTypeIP)
	test_CardinalityTWithTextTemplate[string](t, FieldTypeDate)
}

func test_CardinalityTWithTextTemplate[T any](t *testing.T, ty string) {
	template := []byte(`{"alpha":"{{generate "alpha"}}", "beta":"{{generate "beta"}}"}`)
	if ty == FieldTypeInteger || ty == FieldTypeFloat {
		template = []byte(`{"alpha":{{generate "alpha"}}, "beta":{{generate "beta"}}}`)
	}

	fldAlpha := Field{
		Name: "alpha",
		Type: ty,
	}
	fldBeta := Field{
		Name: "beta",
		Type: ty,
	}

	t.Logf("for type %s, with template: %s", ty, string(template))
	// It's cardinality per mille, so a bit confusing :shrug:
	for cardinality := 1000; cardinality >= 10; cardinality /= 10 {

		currentCardinality := 1000
		currentCardinality /= cardinality

		rangeTrailing := ""
		if ty == FieldTypeFloat {
			rangeTrailing = "."
		}

		rangeMin := rand.Intn(100)
		rangeMax := rand.Intn(10000-rangeMin) + rangeMin

		// Add the range to get some variety in integers
		tmpl := "fields:\n  - name: alpha\n    cardinality: %d\n    range:\n      min: %d%s\n      max: %d%s\n"
		tmpl += "  - name: beta\n    cardinality: %d\n    range:\n      min: %d%s\n      max: %d%s"

		yaml := []byte(fmt.Sprintf(tmpl, currentCardinality, rangeMin, rangeTrailing, rangeMax, rangeTrailing, currentCardinality*2, rangeMin, rangeTrailing, rangeMax, rangeTrailing))

		cfg, err := config.LoadConfigFromYaml(yaml)
		if err != nil {
			t.Fatal(err)
		}

		nSpins := 16384
		g := makeGeneratorWithTextTemplate(t, cfg, []Field{fldAlpha, fldBeta}, template, uint64(nSpins))

		vmapAlpha := make(map[any]int)
		vmapBeta := make(map[any]int)

		for i := 0; i < nSpins; i++ {

			var buf bytes.Buffer
			if err := g.Emit(&buf); err != nil {
				t.Fatal(err)
			}

			m := unmarshalJSONT[T](t, buf.Bytes())

			if len(m) != 2 {
				t.Errorf("Expected map size 1, got %d", len(m))
			}

			v, ok := m[fldAlpha.Name]

			if !ok {
				t.Errorf("Missing key %v", fldAlpha.Name)
			}

			vmapAlpha[v] = vmapAlpha[v] + 1

			v, ok = m[fldBeta.Name]

			if !ok {
				t.Errorf("Missing key %v", fldBeta.Name)
			}

			vmapBeta[v] = vmapBeta[v] + 1
		}

		if len(vmapAlpha) != 1000/cardinality {
			t.Errorf("Expected cardinality of %d got %d", 1000/cardinality, len(vmapAlpha))
		}

		if len(vmapBeta) != 2000/cardinality {
			t.Errorf("Expected cardinality of %d got %d", 2000/cardinality, len(vmapBeta))
		}
	}
}

func Test_FieldBoolWithTextTemplate(t *testing.T) {
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

func Test_FieldConstKeywordWithTextTemplate(t *testing.T) {

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

func Test_FieldStaticOverrideStringWithTextTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("fields:\n  - name: alpha\n    value: beta")
	template := []byte(`{"alpha":"{{generate "alpha"}}"}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[string](t, fld, yaml, template)
	if b != "beta" {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideNumericWithTextTemplate(t *testing.T) {
	fld := Field{

		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("fields:\n  - name: alpha\n    value: 33")
	template := []byte(`{"alpha":{{generate "alpha"}}}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[float64](t, fld, yaml, template)

	if b != 33.0 {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideBoolWithTextTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	yaml := []byte("fields:\n  - name: alpha\n    value: true")
	template := []byte(`{"alpha":{{generate "alpha"}}}`)
	t.Logf("with template: %s", string(template))
	b := testSingleTWithTextTemplate[bool](t, fld, yaml, template)

	if b != true {
		t.Errorf("static value not match")
	}
}

func Test_FieldGeoPointWithTextTemplate(t *testing.T) {
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

func Test_FieldDateWithTextTemplate(t *testing.T) {
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

func Test_FieldDateAndPeriodWithTextTemplate(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeDate,
	}

	template := []byte(`{{$alpha := generate "alpha"}}{"alpha":"{{$alpha.Format "2006-01-02T15:04:05.999999Z07:00"}}"}`)
	configYaml := []byte("fields:\n  - name: alpha\n    period: 10s")
	t.Logf("with template: %s", string(template))

	cfg, err := config.LoadConfigFromYaml(configYaml)
	if err != nil {
		t.Fatal(err)
	}

	g := makeGeneratorWithTextTemplate(t, cfg, []Field{fld}, template, 10)

	var buf bytes.Buffer

	nSpins := 10
	for i := 0; i < nSpins; i++ {
		if err := g.Emit(&buf); err != nil {
			t.Fatal(err)
		}

		m := unmarshalJSONT[string](t, buf.Bytes())
		buf.Reset()

		if len(m) != 1 {
			t.Errorf("Expected map size 1, got %d", len(m))
		}

		v, ok := m[fld.Name]

		if !ok {
			t.Errorf("Missing key %v", fld.Name)
		}

		if ts, err := time.Parse(FieldTypeTimeLayout, v); err != nil {
			t.Errorf("Fail parse timestamp %v", err)
		} else {
			// Timestamp should be +1s for every iteration
			expectedTime := timeNowToBind.Truncate(time.Millisecond).Add(time.Second * time.Duration(i))

			diff := expectedTime.Sub(ts.Truncate(time.Millisecond))
			if diff != 0 {
				t.Errorf("Date generated out of period range %v", diff)
			}
		}
	}
}

func Test_FieldIPWithTextTemplate(t *testing.T) {
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

func Test_FieldFloatsWithTextTemplate(t *testing.T) {
	_testNumericWithTextTemplate[float64](t, FieldTypeDouble)
	_testNumericWithTextTemplate[float32](t, FieldTypeFloat)
	_testNumericWithTextTemplate[float32](t, FieldTypeHalfFloat)
	_testNumericWithTextTemplate[float64](t, FieldTypeScaledFloat)

}

func Test_FieldIntegersWithTextTemplate(t *testing.T) {
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

	g := makeGeneratorWithTextTemplate(t, cfg, []Field{fld}, template, 0)

	var buf bytes.Buffer

	if err := g.Emit(&buf); err != nil {
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

func makeGeneratorWithTextTemplate(t *testing.T, cfg Config, fields Fields, template []byte, totEvents uint64) Generator {
	g, err := NewGeneratorWithTextTemplate(template, cfg, fields, totEvents)

	if err != nil {
		t.Fatal(err)
	}

	return g
}

package genlib

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
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

func Test_EmptyCaseWithHero(t *testing.T) {
	template := generateHeroTemplateFromField(Config{}, []Field{})
	t.Logf("with template: %s", string(template))
	fields := ""

	g, state := makeGeneratorWithHero(t, "", fields, template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	if len(buf.Bytes()) != 0 {
		t.Errorf("Expected empty bytes")
	}
}

func Test_CardinalityWithHero(t *testing.T) {

	test_CardinalityTWithHero[string](t, FieldTypeKeyword)
	test_CardinalityTWithHero[int](t, FieldTypeInteger)
	test_CardinalityTWithHero[float64](t, FieldTypeFloat)
	test_CardinalityTWithHero[string](t, FieldTypeGeoPoint)
	test_CardinalityTWithHero[string](t, FieldTypeIP)
	test_CardinalityTWithHero[string](t, FieldTypeDate)
}

func test_CardinalityTWithHero[T any](t *testing.T, ty string) {
	template := []byte(`<%==s "{\"alpha\": \"" %><%==v generate("alpha") %><%==s "\"}" %>`)
	if ty == "integer" || ty == "float" {
		template = []byte(`<%==s "{\"alpha\": " %><%==v generate("alpha") %><%==s "}" %>`)
	}

	fields := fmt.Sprintf(`
- name: alpha
  type: %s
`, ty)

	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	t.Logf("for type %s, with template: %s", ty, string(template))
	// It's cardinality per mille, so a bit confusing :shrug:
	for cardinality := 1000; cardinality >= 10; cardinality /= 10 {

		// Add the range to get some variety in integers
		cfg := fmt.Sprintf("- name: alpha\n  cardinality: %d\n  range: 10000", cardinality)

		g, state := makeGeneratorWithHero(t, cfg, fields, template)

		vmap := make(map[any]int)

		nSpins := 16384
		for i := 0; i < nSpins; i++ {

			var buf bytes.Buffer
			if err := g.Emit(state, &buf); err != nil {
				t.Fatal(err)
			}

			m := unmarshalJSONT[T](t, buf.Bytes())
			buf.Reset()

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

func Test_FieldBoolWithHero(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeBool,
	}

	template := []byte(`<%==s "{\"alpha\": " %><%==v generate("alpha") %><%==s "}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, "", string(fields), template)

	var buf bytes.Buffer

	// Enough spins, so we can make sure we get at least one true and at least one false
	var cntTrue int
	nSpins := 1024
	for i := 0; i < nSpins; i++ {
		if err := g.Emit(state, &buf); err != nil {
			t.Fatal(err)
		}

		b := testSingleTWithHero[bool](t, fld, &buf)
		buf.Reset()

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

func Test_FieldConstKeywordWithHero(t *testing.T) {

	fld := Field{
		Name:  "alpha",
		Type:  FieldTypeConstantKeyword,
		Value: "constant_keyword",
	}

	template := []byte(`<%==s "{\"alpha\": \"" %><%==v generate("alpha") %><%==s "\"}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, "", string(fields), template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	b := testSingleTWithHero[string](t, fld, &buf)

	if b != fld.Value {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideStringWithHero(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	cfg := "- name: alpha\n  value: beta"
	template := []byte(`<%==s "{\"alpha\": \"" %><%==v generate("alpha") %><%==s "\"}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, cfg, string(fields), template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	b := testSingleTWithHero[string](t, fld, &buf)
	if b != "beta" {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideNumericWithHero(t *testing.T) {
	fld := Field{

		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	cfg := "- name: alpha\n  value: 33"
	template := []byte(`<%==s "{\"alpha\": " %><%==v generate("alpha") %><%==s "}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, cfg, string(fields), template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	b := testSingleTWithHero[float64](t, fld, &buf)

	if b != 33.0 {
		t.Errorf("static value not match")
	}
}

func Test_FieldStaticOverrideBoolWithHero(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeKeyword,
	}

	cfg := "- name: alpha\n  value: true"
	template := []byte(`<%==s "{\"alpha\": " %><%==v generate("alpha") %><%==s "}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, cfg, string(fields), template)

	var buf bytes.Buffer

	if err := g.Emit(state, &buf); err != nil {
		t.Fatal(err)
	}

	b := testSingleTWithHero[bool](t, fld, &buf)

	if b != true {
		t.Errorf("static value not match")
	}
}

func Test_FieldGeoPointWithHero(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeGeoPoint,
	}

	template := []byte(`<%==s "{\"alpha\": \"" %><%==v generate("alpha") %><%==s "\"}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, "", string(fields), template)

	var buf bytes.Buffer

	nSpins := 1024
	for i := 0; i < nSpins; i++ {

		if err := g.Emit(state, &buf); err != nil {
			t.Fatal(err)
		}

		b := testSingleTWithHero[string](t, fld, &buf)
		buf.Reset()

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

func Test_FieldDateWithHero(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeDate,
	}

	template := []byte(`<%==s "{\"alpha\": \"" %><% alpha := generate("alpha") %><% alphaTime := alpha.(time.Time) %><%==v alphaTime.Format("2006-01-02T15:04:05.999999Z07:00") %><%==s "\"}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, "", string(fields), template)

	var buf bytes.Buffer

	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {
		now := time.Now()

		if err := g.Emit(state, &buf); err != nil {
			t.Fatal(err)
		}

		b := testSingleTWithHero[string](t, fld, &buf)
		buf.Reset()

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

func Test_FieldIPWithHero(t *testing.T) {
	fld := Field{
		Name: "alpha",
		Type: FieldTypeIP,
	}

	template := []byte(`<%==s "{\"alpha\": \"" %><%==v generate("alpha") %><%==s "\"}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, "", string(fields), template)

	var buf bytes.Buffer

	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {

		if err := g.Emit(state, &buf); err != nil {
			t.Fatal(err)
		}

		b := testSingleTWithHero[string](t, fld, &buf)
		buf.Reset()

		if ip := net.ParseIP(b); ip == nil {
			t.Errorf("Fail parse ip %s", b)
		}
	}
}

func Test_FieldFloatsWithHero(t *testing.T) {
	_testNumericWithHero[float64](t, FieldTypeDouble)
	_testNumericWithHero[float32](t, FieldTypeFloat)
	_testNumericWithHero[float32](t, FieldTypeHalfFloat)
	_testNumericWithHero[float64](t, FieldTypeScaledFloat)

}

func Test_FieldIntegersWithHero(t *testing.T) {
	_testNumericWithHero[int](t, FieldTypeInteger)
	_testNumericWithHero[int64](t, FieldTypeLong)
	_testNumericWithHero[uint64](t, FieldTypeUnsignedLong)
}

func _testNumericWithHero[T any](t *testing.T, ty string) {
	fld := Field{
		Name: "alpha",
		Type: ty,
	}

	template := []byte(`<%==s "{\"alpha\": " %><%==v generate("alpha") %><%==s "}" %>`)
	t.Logf("with template: %s", string(template))

	fields, err := yaml.Marshal(Fields{fld})
	if err != nil {
		t.Fatal(err)
	}

	g, state := makeGeneratorWithHero(t, "", string(fields), template)

	var buf bytes.Buffer

	nSpins := rand.Intn(1024) + 1
	for i := 0; i < nSpins; i++ {
		if err := g.Emit(state, &buf); err != nil {
			t.Fatal(err)
		}

		testSingleTWithHero[T](t, fld, &buf)
		buf.Reset()
	}
}

func testSingleTWithHero[T any](t *testing.T, fld Field, buf *bytes.Buffer) T {
	// BufferWithMutex should now contain an event shaped like <%==s "{\"alpha\": " %> "constant_keyword<%==s "\"}" %>
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

func makeGeneratorWithHero(t *testing.T, cfg, fields string, template []byte) (Generator, *GenState) {
	tmpFields, err := os.CreateTemp("", "testhero-*")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmpFields.WriteString(fields)
	if err != nil {
		t.Fatal(err)
	}

	configPath := ""
	if len(cfg) > 0 {
		tmpCfg, err := os.CreateTemp("", "testhero-*")
		if err != nil {
			t.Fatal(err)
		}

		_, err = tmpCfg.WriteString(cfg)
		if err != nil {
			t.Fatal(err)
		}

		configPath = tmpCfg.Name()
	}

	g, err := NewGeneratorWithHero(template, configPath, tmpFields.Name())

	if err != nil {
		t.Fatal(err)
	}

	return g, NewGenState()
}

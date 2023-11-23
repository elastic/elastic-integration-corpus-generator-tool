package config

import (
	"github.com/elastic/go-ucfg/yaml"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

const sampleConfigFile = `---
fields:
  - name: field
    value: foobar
`

func TestLoadConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	configFile := "/cfg.yml"

	data := []byte(sampleConfigFile)
	afero.WriteFile(fs, configFile, data, 0666)

	cfg, err := LoadConfig(fs, configFile)
	assert.Nil(t, err)

	f, ok := cfg.GetField("field")
	assert.True(t, ok)
	assert.Equal(t, "field", f.Name)
	assert.Equal(t, "foobar", f.Value.(string))
}

func TestRange_MaxAsFloat64(t *testing.T) {
	testCases := []struct {
		scenario  string
		rangeYaml string
		expected  float64
		hasError  bool
	}{
		{
			scenario:  "max nil",
			rangeYaml: "min: 10",
			expected:  math.MaxFloat64,
			hasError:  true,
		},
		{
			scenario:  "float64",
			rangeYaml: "max: 10.",
			expected:  10,
		},
		{
			scenario:  "uint64",
			rangeYaml: "max: 10",
			expected:  10,
		},
		{
			scenario:  "int64",
			rangeYaml: "max: -10",
			expected:  -10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.rangeYaml))
			if err != nil {
				t.Fatal(err)
			}

			var rangeCfg Range
			err = cfg.Unpack(&rangeCfg)
			if err != nil {
				t.Fatal(err)
			}

			v, err := rangeCfg.MaxAsFloat64()
			if testCase.hasError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !testCase.hasError && err != nil {
				t.Fatal("expected no error but got one")
			}
			if testCase.expected != v {
				t.Fatalf("expected %v, got %v", testCase.expected, v)
			}
		})
	}
}

func TestRange_MaxAsInt64(t *testing.T) {
	testCases := []struct {
		scenario  string
		rangeYaml string
		expected  int64
		hasError  bool
	}{
		{
			scenario:  "max nil",
			rangeYaml: "min: 10",
			expected:  math.MaxInt64,
			hasError:  true,
		},
		{
			scenario:  "float64",
			rangeYaml: "max: 10.",
			expected:  10,
		},
		{
			scenario:  "uint64",
			rangeYaml: "max: 10",
			expected:  10,
		},
		{
			scenario:  "int64",
			rangeYaml: "max: -10",
			expected:  -10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.rangeYaml))
			if err != nil {
				t.Fatal(err)
			}

			var rangeCfg Range
			err = cfg.Unpack(&rangeCfg)
			if err != nil {
				t.Fatal(err)
			}

			v, err := rangeCfg.MaxAsInt64()
			if testCase.hasError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !testCase.hasError && err != nil {
				t.Fatal("expected no error but got one")
			}
			if testCase.expected != v {
				t.Fatalf("expected %v, got %v", testCase.expected, v)
			}
		})
	}
}

func TestRange_MinAsFloat64(t *testing.T) {
	testCases := []struct {
		scenario  string
		rangeYaml string
		expected  float64
		hasError  bool
	}{
		{
			scenario:  "min nil",
			rangeYaml: "max: 10",
			expected:  0,
			hasError:  true,
		},
		{
			scenario:  "float64",
			rangeYaml: "min: 10.",
			expected:  10,
		},
		{
			scenario:  "uint64",
			rangeYaml: "min: 10",
			expected:  10,
		},
		{
			scenario:  "int64",
			rangeYaml: "min: -10",
			expected:  -10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.rangeYaml))
			if err != nil {
				t.Fatal(err)
			}

			var rangeCfg Range
			err = cfg.Unpack(&rangeCfg)
			if err != nil {
				t.Fatal(err)
			}

			v, err := rangeCfg.MinAsFloat64()
			if testCase.hasError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !testCase.hasError && err != nil {
				t.Fatal("expected no error but got one")
			}
			if testCase.expected != v {
				t.Fatalf("expected %v, got %v", testCase.expected, v)
			}
		})
	}
}

func TestRange_MinAsInt64(t *testing.T) {
	testCases := []struct {
		scenario  string
		rangeYaml string
		expected  int64
		hasError  bool
	}{
		{
			scenario:  "min nil",
			rangeYaml: "max: 10",
			expected:  0,
			hasError:  true,
		},
		{
			scenario:  "float64",
			rangeYaml: "min: 10.",
			expected:  10,
		},
		{
			scenario:  "uint64",
			rangeYaml: "min: 10",
			expected:  10,
		},
		{
			scenario:  "int64",
			rangeYaml: "min: -10",
			expected:  -10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.rangeYaml))
			if err != nil {
				t.Fatal(err)
			}

			var rangeCfg Range
			err = cfg.Unpack(&rangeCfg)
			if err != nil {
				t.Fatal(err)
			}

			v, err := rangeCfg.MinAsInt64()
			if testCase.hasError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !testCase.hasError && err != nil {
				t.Fatal("expected no error but got one")
			}
			if testCase.expected != v {
				t.Fatalf("expected %v, got %v", testCase.expected, v)
			}
		})
	}
}

func TestRange_FromAsTime(t *testing.T) {
	from, err := time.Parse("2006-01-02T15:04:05Z", "2023-11-23T08:35:38Z")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		scenario  string
		rangeYaml string
		expected  time.Time
		hasError  bool
	}{
		{
			scenario:  "from nil",
			rangeYaml: "to: 2023-11-23T08:35:38Z",
			expected:  time.Time{},
			hasError:  true,
		},
		{
			scenario:  "from not nil",
			rangeYaml: "from: 2023-11-23T08:35:38Z",
			expected:  from,
			hasError:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.rangeYaml))
			if err != nil {
				t.Fatal(err)
			}

			var rangeCfg Range
			err = cfg.Unpack(&rangeCfg)
			if err != nil {
				t.Fatal(err)
			}

			v, err := rangeCfg.FromAsTime()
			if testCase.hasError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !testCase.hasError && err != nil {
				t.Fatal("expected no error but got one")
			}
			if testCase.expected != v {
				t.Fatalf("expected %v, got %v", testCase.expected, v)
			}
		})
	}
}

func TestRange_ToAsTime(t *testing.T) {
	to, err := time.Parse("2006-01-02T15:04:05Z", "2023-11-23T08:35:38Z")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		scenario  string
		rangeYaml string
		expected  time.Time
		hasError  bool
	}{
		{
			scenario:  "to nil",
			rangeYaml: "from: 2023-11-23T08:35:38Z",
			expected:  time.Time{},
			hasError:  true,
		},
		{
			scenario:  "to not nil",
			rangeYaml: "to: 2023-11-23T08:35:38Z",
			expected:  to,
			hasError:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.rangeYaml))
			if err != nil {
				t.Fatal(err)
			}

			var rangeCfg Range
			err = cfg.Unpack(&rangeCfg)
			if err != nil {
				t.Fatal(err)
			}

			v, err := rangeCfg.ToAsTime()
			if testCase.hasError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !testCase.hasError && err != nil {
				t.Fatal("expected no error but got one")
			}
			if testCase.expected != v {
				t.Fatalf("expected %v, got %v", testCase.expected, v)
			}
		})
	}
}

func TestPeriod(t *testing.T) {
	testCases := []struct {
		scenario   string
		periodYaml string
		expected   time.Duration
		hasField   bool
	}{
		{
			scenario:   "time duration as number",
			periodYaml: "- name: testField\n  period: 10",
			expected:   10 * time.Second,
			hasField:   true,
		},
		{
			scenario: "empty period",
			hasField: false,
		},
		{
			scenario:   "1h",
			periodYaml: "- name: testField\n  period: 1h",
			expected:   3600 * time.Second,
			hasField:   true,
		},
		{
			scenario:   "-1h",
			periodYaml: "- name: testField\n  period: -1h",
			expected:   -3600 * time.Second,
			hasField:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			cfg, err := yaml.NewConfig([]byte(testCase.periodYaml))
			if err != nil {
				t.Fatal(err)
			}

			var periodCfg []ConfigField
			err = cfg.Unpack(&periodCfg)
			if err != nil {
				t.Fatal(err)
			}

			ok := len(periodCfg) == 1
			if !testCase.hasField && ok {
				t.Fatalf("expected missing field but got: %v", periodCfg[0])
			}
			if testCase.hasField && !ok {
				t.Fatal("expected field but missing")
			}

			if ok && testCase.expected != periodCfg[0].Period {
				t.Fatalf("expected %v, got %v", testCase.expected, periodCfg[0].Period)
			}
		})
	}
}

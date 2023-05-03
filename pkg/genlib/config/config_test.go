package config

import (
	"github.com/elastic/go-ucfg/yaml"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
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
	//fmt.Printf("%+v\n", cfg)

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

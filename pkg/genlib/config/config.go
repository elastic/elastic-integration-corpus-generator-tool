package config

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"math"
	"os"

	"github.com/elastic/go-ucfg/yaml"
	"github.com/spf13/afero"
)

var rangeBoundNotSet = errors.New("range bound not set")
var rangeTimeNotSet = errors.New("range time not set")
var rangeInvalidConfig = errors.New("range defining both `period` and `from`/`to`")
var counterInvalidConfig = errors.New("both `range` and `counter` defined")

type TimeRange struct {
	time.Time
}

func (ct *TimeRange) Unpack(t string) error {
	var err error
	ct.Time, err = time.Parse("2006-01-02T15:04:05.999999999-07:00", t)
	return err
}

type Range struct {
	// NOTE: we want to distinguish when Min/Max/From/To are explicitly set to zero value or are not set at all. We use a pointer, such that when not set will be `nil`.
	Min  *float64   `config:"min"`
	Max  *float64   `config:"max"`
	From *TimeRange `config:"from"`
	To   *TimeRange `config:"to"`
}

type Config struct {
	m map[string]ConfigField
}

// HasFieldsMappings checks if all fields have a type defined.
// This is useful to determine if a fields definitions file is required.
func (c Config) HasFieldsMappings() bool {
	for _, f := range c.m {
		if f.Type == "" {
			return false
		}
	}
	return true
}

// FieldMapping represents the mapping of a field.
// It defines the structure and properties of a field, including its name,
// data type, associated object type, example value, and the actual value.
type FieldMapping struct {
	Name       string
	Type       string
	ObjectType string
	Example    string
	Value      any
}

// FieldsMappings is a collection of FieldMapping.
type FieldsMappings []FieldMapping

func (f FieldsMappings) Len() int           { return len(f) }
func (f FieldsMappings) Less(i, j int) bool { return f[i].Name < f[j].Name }
func (f FieldsMappings) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func normaliseFields(fields FieldsMappings) (FieldsMappings, error) {
	sort.Sort(fields)
	normalisedFields := make(FieldsMappings, 0, len(fields))
	for _, field := range fields {
		if !strings.Contains(field.Name, "*") {
			normalisedFields = append(normalisedFields, field)
			continue
		}

		normalizationPattern := strings.NewReplacer(".", "\\.", "*", ".+").Replace(field.Name)
		re, err := regexp.Compile(normalizationPattern)
		if err != nil {
			return nil, err
		}

		hasMatch := false
		for _, otherField := range fields {
			if otherField.Name == field.Name {
				continue
			}

			if re.MatchString(otherField.Name) {
				hasMatch = true
				break
			}
		}

		if !hasMatch {
			normalisedFields = append(normalisedFields, field)
		}
	}

	sort.Sort(normalisedFields)
	return normalisedFields, nil
}

// LoadFieldsMappings creates the fields mappings from the config itself.
// It has to be called after the config is loaded.
func (c Config) LoadFieldsMappings() (FieldsMappings, error) {
	var mappings FieldsMappings

	for _, f := range c.m {
		mappings = append(mappings, FieldMapping{
			Name:       f.Name,
			Type:       f.Type,
			ObjectType: f.ObjectType,
			Example:    f.Example,
			Value:      f.Value,
		})
	}

	return normaliseFields(mappings)
}

type ConfigField struct {
	Name        string        `config:"name"`
	Type        string        `config:"type"`
	ObjectType  string        `config:"object_type"`
	Example     string        `config:"example"`
	Fuzziness   float64       `config:"fuzziness"`
	Range       Range         `config:"range"`
	Cardinality int           `config:"cardinality"`
	Period      time.Duration `config:"period"`
	Enum        []string      `config:"enum"`
	ObjectKeys  []string      `config:"object_keys"`
	Value       any           `config:"value"`
	Counter     bool          `config:"counter"`
}

func (cf ConfigField) ValidForDateField() error {
	if cf.Period.Abs() > 0 && (cf.Range.From != nil || cf.Range.To != nil) {
		return rangeInvalidConfig
	}

	return nil
}

func (cf ConfigField) ValidCounter() error {
	if cf.Counter && (cf.Range.Min != nil || cf.Range.Max != nil) {
		return counterInvalidConfig
	}

	return nil
}

func (r Range) FromAsTime() (time.Time, error) {
	if r.From == nil {
		return time.Time{}, rangeTimeNotSet
	}

	return r.From.Time, nil
}

func (r Range) ToAsTime() (time.Time, error) {
	if r.To == nil {
		return time.Time{}, rangeTimeNotSet
	}

	return r.To.Time, nil
}

func (r Range) MinAsInt64() (int64, error) {
	if r.Min == nil {
		return 0, rangeBoundNotSet
	}

	return int64(*r.Min), nil
}

func (r Range) MaxAsInt64() (int64, error) {
	if r.Max == nil {
		return math.MaxInt64, rangeBoundNotSet
	}

	return int64(*r.Max), nil
}

func (r Range) MinAsFloat64() (float64, error) {
	if r.Min == nil {
		return 0, rangeBoundNotSet
	}

	return *r.Min, nil
}

func (r Range) MaxAsFloat64() (float64, error) {
	if r.Max == nil {
		return math.MaxFloat64, rangeBoundNotSet
	}

	return *r.Max, nil
}

type ConfigFile struct {
	Fields []ConfigField `config:"fields"`
}

func LoadConfig(fs afero.Fs, configFile string) (Config, error) {
	if len(configFile) == 0 {
		return Config{}, nil
	}

	configFile = os.ExpandEnv(configFile)
	if _, err := fs.Stat(configFile); err != nil {
		return Config{}, err
	}

	data, err := afero.ReadFile(fs, configFile)
	if err != nil {
		return Config{}, err
	}

	return LoadConfigFromYaml(data)
}

func LoadConfigFromYaml(c []byte) (Config, error) {

	cfg, err := yaml.NewConfig(c)
	if err != nil {
		return Config{}, err
	}

	var cfgfile ConfigFile
	err = cfg.Unpack(&cfgfile)
	if err != nil {
		return Config{}, err
	}

	outCfg := Config{
		m: make(map[string]ConfigField),
	}

	for _, c := range cfgfile.Fields {
		outCfg.m[c.Name] = c
	}

	return outCfg, nil
}

func (c Config) GetField(fieldName string) (ConfigField, bool) {
	v, ok := c.m[fieldName]
	return v, ok
}

func (c Config) SetField(fieldName string, configField ConfigField) {
	configField.Name = fieldName
	c.m[fieldName] = configField
}

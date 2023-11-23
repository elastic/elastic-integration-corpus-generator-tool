package config

import (
	"errors"
	"time"

	"math"
	"os"

	"github.com/elastic/go-ucfg/yaml"
	"github.com/spf13/afero"
)

var rangeBoundNotSet = errors.New("range bound not set")
var rangeTimeNotSet = errors.New("range time not set")

type TimeRange struct {
	time.Time
}

func (ct *TimeRange) Unpack(t string) error {
	var err error
	ct.Time, err = time.Parse(time.RFC3339Nano, t)
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

type ConfigField struct {
	Name        string        `config:"name"`
	Fuzziness   float64       `config:"fuzziness"`
	Range       Range         `config:"range"`
	Cardinality int           `config:"cardinality"`
	Period      time.Duration `config:"period"`
	Enum        []string      `config:"enum"`
	ObjectKeys  []string      `config:"object_keys"`
	Value       any           `config:"value"`
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

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

type ConfigField struct {
	Name              string        `config:"name"`
	Fuzziness         float64       `config:"fuzziness"`
	Range             Range         `config:"range"`
	Cardinality       int           `config:"cardinality"`
	Period            time.Duration `config:"period"`
	Enum              []string      `config:"enum"`
	ObjectKeys        []string      `config:"object_keys"`
	Value             any           `config:"value"`
	Counter           bool          `config:"counter"`
	CounterReset      *CounterReset `config:"counter_reset"`
	FormattingPattern string        `config:"formatting_pattern"`
}

const (
	CounterResetStrategyRandom        string = "random"
	CounterResetStrategyProbabilistic string = "probabilistic"
	CounterResetStrategyAfterN        string = "after_n"
)

type CounterReset struct {
	Strategy    string  `config:"strategy"`
	Probability *uint64 `config:"probability"`
	ResetAfterN *uint64 `config:"reset_after_n"`
}

func (cf ConfigField) ValidateCounterResetStrategy() error {
	if cf.Counter && cf.CounterReset != nil &&
		cf.CounterReset.Strategy != CounterResetStrategyRandom &&
		cf.CounterReset.Strategy != CounterResetStrategyProbabilistic &&
		cf.CounterReset.Strategy != CounterResetStrategyAfterN {
		return errors.New("counter_reset strategy must be one of 'random', 'probabilistic', 'after_n'")
	}

	return nil
}

func (cf ConfigField) ValidateCounterResetAfterN() error {
	if cf.Counter && cf.CounterReset != nil && cf.CounterReset.Strategy == CounterResetStrategyAfterN && cf.CounterReset.ResetAfterN == nil {
		return errors.New("counter_reset after_n requires 'reset_after_n' value to be set")
	}

	return nil
}

func (cf ConfigField) ValidateCounterResetProbabilistic() error {
	if cf.Counter && cf.CounterReset != nil && cf.CounterReset.Strategy == CounterResetStrategyProbabilistic && cf.CounterReset.Probability == nil {
		return errors.New("counter_reset probabilistic requires 'probability' value to be set")
	}

	return nil
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

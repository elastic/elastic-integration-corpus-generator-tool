package config

import (
	"os"

	"github.com/elastic/go-ucfg/yaml"
	"github.com/spf13/afero"
)

type Ratio struct {
	Numerator   int `config:"numerator"`
	Denominator int `config:"denominator"`
}

type Range struct {
	Min interface{} `config:"min"`
	Max interface{} `config:"max"`
}

type Config struct {
	m map[string]ConfigField
}

type ConfigField struct {
	Name        string      `config:"name"`
	Fuzziness   Ratio       `config:"fuzziness"`
	Range       Range       `config:"range"`
	Cardinality Ratio       `config:"cardinality"`
	Enum        []string    `config:"enum"`
	ObjectKeys  []string    `config:"object_keys"`
	Value       interface{} `config:"value"`
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

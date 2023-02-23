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

	var cfgList []ConfigField
	err = cfg.Unpack(&cfgList)
	if err != nil {
		return Config{}, err
	}

	outCfg := Config{
		m: make(map[string]ConfigField),
	}

	for _, c := range cfgList {
		outCfg.m[c.Name] = c
	}

	return outCfg, nil
}

func (c Config) GetField(fieldName string) (ConfigField, bool) {
	v, ok := c.m[fieldName]
	return v, ok
}

package config

import (
	"github.com/elastic/go-ucfg/yaml"
	"io/ioutil"
	"os"
)

type Config struct {
	m map[string]ConfigField
}

type ConfigField struct {
	Name        string      `config:"name"`
	Fuzziness   int         `config:"fuzziness"`
	Range       int         `config:"range"`
	Cardinality int         `config:"cardinality"`
	Enum        []string    `config:"enum"`
	ObjectKeys  []string    `config:"object_keys"`
	Value       interface{} `config:"value"`
	Expression  string      `config:"expression"`
}

func LoadConfig(configFile string) (Config, error) {
	if len(configFile) == 0 {
		return Config{}, nil
	}

	configFile = os.ExpandEnv(configFile)
	if _, err := os.Stat(configFile); err != nil {
		return Config{}, err
	}

	data, err := ioutil.ReadFile(configFile)
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

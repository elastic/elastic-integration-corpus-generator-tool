package config_test

import (
	"testing"

	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const sampleConfigFile = `---
- name: field
  value: foobar
`

func TestLoadConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	configFile := "/cfg.yml"

	data := []byte(sampleConfigFile)
	afero.WriteFile(fs, configFile, data, 0666)

	cfg, err := config.LoadConfig(fs, configFile)
	assert.Nil(t, err)
	//fmt.Printf("%+v\n", cfg)

	f, ok := cfg.GetField("field")
	assert.True(t, ok)
	assert.Equal(t, "field", f.Name)
	assert.Equal(t, "foobar", f.Value.(string))

}

// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package settings

import (
	"os"
	"path"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/spf13/viper"
)

// Init initalize settings and default values
func Init() {
	viper.AutomaticEnv()
	// NOTE: err value is ignored as it only checks for missing argument
	_ = viper.BindEnv("ELASTIC_INTEGRATION_CORPUS")

	setDefaults()
	setConstants()
}

func setDefaults() {
	viper.SetDefault("cache_dir", xdg.CacheHome())
	viper.SetDefault("config_dir", xdg.ConfigHome())
	viper.SetDefault("data_dir", xdg.DataHome())

	// fragment_root supports env var expansion
	viper.SetDefault("corpora_root", path.Join(viper.GetString("data_dir"), "elastic-integration-corpus-generator-tool"))
	viper.SetDefault("corpora_path", "corpora")
	viper.SetDefault("corpora_location", path.Join(
		os.ExpandEnv(viper.GetString("corpora_root")),
		viper.GetString("corpora_path")))
}

func setConstants() {
	// viper.Set()
}

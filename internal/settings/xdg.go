// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package settings

import (
	"github.com/spf13/viper"
)

func CacheDir() string {
	return viper.GetString("cache_dir")
}

func ConfigDir() string {
	return viper.GetString("config_dir")
}

func DataDir() string {
	return viper.GetString("data_dir")
}

// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package settings_test

import (
	"os"
	"testing"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/elastic-integration-corpus-generator-tool/internal/settings"
)

func TestCacheDir(t *testing.T) {
	settings.Init()

	expected := xdg.CacheHome()
	got := settings.CacheDir()

	assert.Equal(t, expected, got)
}

func TestCacheDir_customValue(t *testing.T) {
	settings.Init()

	expected := "foobar"
	viper.Set("cache_dir", expected)
	got := settings.CacheDir()

	assert.Equal(t, expected, got)
}

func TestCacheDir_valueFromEnv(t *testing.T) {
	settings.Init()

	expected := "foobar"
	os.Setenv("ELASTIC_INTEGRATION_CORPUS_CACHE_DIR", expected)
	got := settings.CacheDir()

	assert.Equal(t, expected, got)
}

func TestConfigDir(t *testing.T) {
	settings.Init()

	expected := xdg.ConfigHome()
	got := settings.ConfigDir()

	assert.Equal(t, expected, got)
}

func TestConfigDir_customValue(t *testing.T) {
	settings.Init()

	expected := "foobar"
	viper.Set("config_dir", expected)
	got := settings.ConfigDir()

	assert.Equal(t, expected, got)
}

func TestConfigDir_valueFromEnv(t *testing.T) {
	settings.Init()

	expected := "foobar"
	os.Setenv("ELASTIC_INTEGRATION_CORPUS_CONFIG_DIR", expected)
	got := settings.ConfigDir()

	assert.Equal(t, expected, got)
}

func TestDataDir(t *testing.T) {
	settings.Init()

	expected := xdg.DataHome()
	got := settings.DataDir()

	assert.Equal(t, expected, got)
}

func TestDataDir_customValue(t *testing.T) {
	settings.Init()

	expected := "foobar"
	viper.Set("data_dir", expected)
	got := settings.DataDir()

	assert.Equal(t, expected, got)
}

func TestDataDir_valueFromEnv(t *testing.T) {
	settings.Init()

	expected := "foobar"
	os.Setenv("ELASTIC_INTEGRATION_CORPUS_DATA_DIR", expected)
	got := settings.DataDir()

	assert.Equal(t, expected, got)
}

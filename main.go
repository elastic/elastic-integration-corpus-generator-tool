// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package main

import (
	"os"

	"github.com/elastic/elastic-integration-corpus-generator-tool/cmd"
	"github.com/elastic/elastic-integration-corpus-generator-tool/internal/settings"
)

func main() {
	settings.Init()

	rootCmd := cmd.RootCmd()
	rootCmd.AddCommand(cmd.GenerateCmd())
	rootCmd.AddCommand(cmd.GenerateWithTemplateCmd())
	rootCmd.AddCommand(cmd.VersionCmd())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

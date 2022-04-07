// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd creates and returns root cmd for elastic-integration-corpus-generator-tool.
func RootCmd() *cobra.Command {

	rootCmd := &cobra.Command{
		Use:          "elastic-integration-corpus-generator-tool",
		Long:         "elastic-integration-corpus-generator-tool - Command line tool used for generating events corpus dynamically given a specific integration.",
		SilenceUsage: true,
	}

	return rootCmd
}

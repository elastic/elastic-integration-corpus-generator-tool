// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package cmd

import (
	"errors"
	"fmt"
	"github.com/elastic/elastic-integration-corpus-generator-tool/internal/corpus"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
)

var integrationPackage string
var dataStream string
var packageVersion string

func GenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate integration data_stream version",
		Short: "Generate a corpus",
		Long:  "Generate a bulk request corpus for a given integration data stream downloaded from a package registry",
		Args: func(cmd *cobra.Command, args []string) error {
			var errs []error
			if len(args) != 3 {
				return errors.New("you must pass the integration package the data stream and the package vesion")
			}

			if packageRegistryBaseURL == "" {
				errs = append(errs, errors.New("you must provide a not empty --package-registry-base-url flag value"))
			}

			integrationPackage = args[0]
			if integrationPackage == "" {
				errs = append(errs, errors.New("you must provide a not empty integration argument"))
			}

			dataStream = args[1]
			if dataStream == "" {
				errs = append(errs, errors.New("you must provide a not empty data stream argument"))
			}

			packageVersion = args[2]
			if packageVersion == "" {
				errs = append(errs, errors.New("you must provide a not empty package version argument"))
			}

			if len(errs) > 0 {
				return multierr.Combine(errs...)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := afero.NewOsFs()
			location := viper.GetString("corpora_location")

			cfg, err := config.LoadConfig(fs, configFile)
			if err != nil {
				return err
			}

			fc, err := corpus.NewGenerator(cfg, fs, location)
			if err != nil {
				return err
			}

			payloadFilename, err := fc.Generate(packageRegistryBaseURL, integrationPackage, dataStream, packageVersion, totEvents)
			if err != nil {
				return err
			}

			fmt.Println("File generated:", payloadFilename)

			return nil
		},
	}

	generateCmd.Flags().StringVarP(&packageRegistryBaseURL, "package-registry-base-url", "r", "https://epr.elastic.co/", "base url of the package registry with schema")
	generateCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "path to config file for generator settings")
	generateCmd.Flags().Uint64VarP(&totEvents, "tot-events", "t", 0, "total events of the corpus to generate")
	return generateCmd
}

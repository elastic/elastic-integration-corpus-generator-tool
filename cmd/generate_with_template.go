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

var templatePath string
var fieldsDefinitionPath string
var dataType string

func GenerateWithTemplateCmd() *cobra.Command {
	generateWithTemplateCmd := &cobra.Command{
		Use:   "generate-with-template template-path fields-definition-path data-type",
		Short: "Generate a corpus",
		Long:  "Generate a bulk request corpus given a template path and a data type",
		Args: func(cmd *cobra.Command, args []string) error {
			var errs []error
			if len(args) != 3 {
				return errors.New("you must pass the template path and the data type")
			}

			if totSize == "" {
				errs = append(errs, errors.New("you must provide a not empty --tot-size flag value"))
			}

			templatePath = args[0]
			if templatePath == "" {
				errs = append(errs, errors.New("you must provide a not empty template path argument"))
			}

			fieldsDefinitionPath = args[1]
			if fieldsDefinitionPath == "" {
				errs = append(errs, errors.New("you must provide a not empty fields definition path argument"))
			}
			dataType = args[2]
			if dataType == "" {
				errs = append(errs, errors.New("you must provide a not empty data type argument"))
			}

			if len(errs) > 0 {
				return multierr.Combine(errs...)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			location := viper.GetString("corpora_location")
			config, err := config.LoadConfig(configFile)
			if err != nil {
				return err
			}

			fc, err := corpus.NewGenerator(config, afero.NewOsFs(), location)
			if err != nil {
				return err
			}

			payloadFilename, err := fc.GenerateWithTemplate(templatePath, fieldsDefinitionPath, dataType, totSize)
			if err != nil {
				return err
			}

			fmt.Println("File generated:", payloadFilename)

			return nil
		},
	}

	generateWithTemplateCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "path to config file for generator settings")
	generateWithTemplateCmd.Flags().StringVarP(&totSize, "tot-size", "t", "", "total size of the corpus to generate")
	return generateWithTemplateCmd
}

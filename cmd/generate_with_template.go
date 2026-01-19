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

var templateType string

var (
	templatePath         string
	fieldsDefinitionPath string
)

func GenerateWithTemplateCmd() *cobra.Command {
	generateWithTemplateCmd := &cobra.Command{
		Use:   "generate-with-template template-path fields-definition-path",
		Short: "Generate a corpus",
		Long:  "Generate a bulk request corpus given a template path and a fields definition path",
		Args: func(cmd *cobra.Command, args []string) error {
			var errs []error
			if len(args) != 2 {
				return errors.New("you must pass the template path and the fields definition path")
			}

			templatePath = args[0]
			if templatePath == "" {
				errs = append(errs, errors.New("you must provide a not empty template path argument"))
			}

			fieldsDefinitionPath = args[1]
			if fieldsDefinitionPath == "" {
				errs = append(errs, errors.New("you must provide a not empty fields definition path argument"))
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

			fc, err := corpus.NewGeneratorWithTemplate(cfg, fs, location, templateType)
			if err != nil {
				return err
			}

			timeNow, err := getTimeNowFromFlag(timeNowAsString)
			if err != nil {
				return err
			}

			payloadFilename, err := fc.GenerateWithTemplate(templatePath, fieldsDefinitionPath, totEvents, timeNow, randSeed)
			if err != nil {
				return err
			}

			fmt.Println("File generated:", payloadFilename)

			return nil
		},
	}

	generateWithTemplateCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "path to config file for generator settings")
	generateWithTemplateCmd.Flags().StringVarP(&templateType, "template-type", "y", "placeholder", "either 'placeholder' or 'gotext'")
	generateWithTemplateCmd.Flags().Uint64VarP(&totEvents, "tot-events", "t", 1, "total events of the corpus to generate")
	generateWithTemplateCmd.Flags().StringVarP(&timeNowAsString, "now", "n", "", "time to use for generation based on now (`date` type)")
	generateWithTemplateCmd.Flags().Int64VarP(&randSeed, "seed", "s", 1, "seed to set as source of rand")

	return generateWithTemplateCmd
}

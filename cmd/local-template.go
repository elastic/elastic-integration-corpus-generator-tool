// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/elastic/elastic-integration-corpus-generator-tool/internal/corpus"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
)

var flag_schema string

func TemplateCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "local-template package dataset",
		Example: "local-template aws billing",
		Short:   "Generate a corpus from a local template",
		Long:    "Generate a bulk request corpus for the specified package dataset in the assets/templates folder",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("package and dataset arguments are required")
			}

			datasetFolder := filepath.Join("assets", "templates", fmt.Sprintf("%s.%s", args[0], args[1]))
			if _, err := os.Stat(datasetFolder); errors.Is(err, os.ErrNotExist) {
				return errors.New(fmt.Sprintf("dataset folder %s does not exists", datasetFolder))
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

			var errs []error
			datasetFolder := fmt.Sprintf("%s.%s", args[0], args[1])
			schema := fmt.Sprintf("schema-%s", flag_schema)
			datasetFolderPath := filepath.Join("assets", "templates", datasetFolder, schema)

			templateFile := fmt.Sprintf("%s.tpl", templateType)
			templatePath := filepath.Join(datasetFolderPath, templateFile)
			if _, err := os.Stat(templatePath); errors.Is(err, os.ErrNotExist) {
				errs = append(errs, errors.New(fmt.Sprintf("template file %s does not exist", templatePath)))
			}

			fieldsDefinitionFile := "fields.yml"
			fieldsDefinitionPath := filepath.Join(datasetFolderPath, fieldsDefinitionFile)
			if _, err := os.Stat(templatePath); errors.Is(err, os.ErrNotExist) {
				errs = append(errs, errors.New(fmt.Sprintf("fields definition file %s does not exist", fieldsDefinitionPath)))
			}

			fieldsConfigFile := "configs.yml"
			fieldsConfigFilePath := filepath.Join(datasetFolderPath, fieldsConfigFile)
			if _, err := os.Stat(fieldsConfigFilePath); errors.Is(err, os.ErrNotExist) {
				log.Printf("fields config file %s does not exist", fieldsConfigFilePath)
			}

			if len(errs) > 0 {
				return multierr.Combine(errs...)
			}

			fc, err := corpus.NewGeneratorWithTemplate(cfg, afero.NewOsFs(), location, templateType)
			if err != nil {
				return err
			}

			payloadFilename, err := fc.GenerateWithTemplate(templatePath, fieldsDefinitionPath, totEvents)
			if err != nil {
				return err
			}

			fmt.Println("File generated:", payloadFilename)

			return nil
		},
	}

	command.Flags().StringVarP(&configFile, "config-file", "c", "", "path to config file for generator settings")
	command.Flags().StringVarP(&templateType, "engine", "e", "gotext", "either 'placeholder' or 'gotext'")
	command.Flags().Uint64VarP(&totEvents, "tot-events", "t", 0, "total events of the corpus to generate")
	command.Flags().StringVarP(&flag_schema, "schema", "", "b", "schema to generate data for; valid values: a, b")
	return command
}

/*
Copyright © 2023 The tf2d2 Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/tf2d2/tf2d2/internal/diagram"
	"github.com/tf2d2/tf2d2/internal/inframap"
	"github.com/tf2d2/tf2d2/internal/tfcloud"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CliRuntime represents the CLI execution runtime
type CliRuntime struct {
	ctx    context.Context
	logger hclog.Logger

	cmd    *cobra.Command
	Config *Config
}

// Config represents the CLI runtime config
type Config struct {
	CfgFile      string `mapstructure:"-"`
	Hostname     string `mapstructure:"hostname"`
	Organization string `mapstructure:"organization"`
	Workspace    string `mapstructure:"workspace"`
	Token        string `mapstructure:"token"`
	OutputFile   string `mapstructure:"output-file"`
	OutputPath   string `mapstructure:"output-path"`
	StateFile    string `mapstructure:"state-file"`
	Verbose      bool   `mapstructure:"verbose"`
	DryRun       bool   `mapstructure:"dry-run"`
}

// NewCliRuntime returns a new CLI runtime instance
func NewCliRuntime(ctx context.Context, logger hclog.Logger) *CliRuntime {
	return &CliRuntime{
		cmd:    nil,
		ctx:    hclog.WithContext(ctx, logger),
		logger: logger,
		Config: &Config{},
	}
}

// PreRunE executes before `RunE` to configure Viper and logging level
func (r *CliRuntime) PreRunE(cmd *cobra.Command, _ []string) error {
	r.cmd = cmd

	if err := r.bindFlags(); err != nil {
		return err
	}

	if err := r.readConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(r.Config); err != nil {
		return err
	}

	if r.Config.Verbose {
		r.logger.SetLevel(hclog.LevelFromString("debug"))
	}

	return nil
}

// RunE executes the logic to generate a d2 diagram from Terraform state
func (r *CliRuntime) RunE(_ *cobra.Command, _ []string) error {
	// get local or remote Terraform state
	var tfJsonState []byte
	if r.Config.StateFile != "" {
		// open and read the Terraform state JSON file
		bytes, err := os.ReadFile(filepath.Clean(r.Config.StateFile))
		if err != nil {
			return err
		}
		tfJsonState = bytes
		r.logger.Info("loaded local terraform state", "file", r.Config.StateFile)
	} else {
		tfe, err := tfcloud.NewTFClient(r.Config.Hostname, r.Config.Token)
		if err != nil {
			r.logger.Error("failed to initialize terraform cloud client", "error", err)
			return err
		}

		c := tfcloud.NewTFCloud(tfe)
		stateCommand := &tfcloud.GetStateVersion{
			Context:      r.ctx,
			TFCloud:      c,
			Organization: r.Config.Organization,
			Workspace:    r.Config.Workspace,
		}
		tfJsonState, err = stateCommand.Run()
		if err != nil {
			return err
		}
		r.logger.Info("loaded remote terraform state", "organization", r.Config.Organization, "workspace", r.Config.Workspace)
	}

	if err := r.generateContent(tfJsonState); err != nil {
		return err
	}

	return nil
}

// readConfig loads Viper config
func (r *CliRuntime) readConfig() error {
	if r.cmd.Flags().Changed("config") {
		viper.SetConfigFile(r.Config.CfgFile)
	} else {
		viper.SetConfigName(".tf2d2")
		viper.SetConfigType("yml")
	}

	// Search config in the following directories with name ".tf2d2" (without extension)
	viper.AddConfigPath("$HOME/") // home directory
	viper.AddConfigPath(".")      // current directory

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	var pathError *os.PathError
	var notFoundError viper.ConfigFileNotFoundError
	switch {
	case err != nil && errors.As(err, &pathError):
		r.logger.Warn("no config file found", "warning", err.Error())
	case err != nil && !errors.As(err, &notFoundError):
		// pathError and notFoundError are produced when no config file is found
		// here we check and return an error produced when reading the config file
		r.logger.Error("failed to read config", "error", err.Error())
		return err
	default:
		r.logger.Info("using config file", "path", viper.ConfigFileUsed())
	}
	return nil
}

// bindFlags binds all Cobra flags to Viper config
func (r *CliRuntime) bindFlags() error {
	fs := r.cmd.Flags()
	if err := viper.BindPFlags(fs); err != nil {
		return err
	}
	return nil
}

// generateContent uses Terraform state to generate and render a d2 diagram
func (r *CliRuntime) generateContent(state []byte) error {
	// build output file path for generated output
	filepath := filepath.Join(r.Config.OutputPath, r.Config.OutputFile)

	// generate output
	infraMap, err := inframap.GenerateInfraMap(r.ctx, state)
	if err != nil {
		return err
	}

	d := diagram.NewDiagram(r.ctx, infraMap, filepath)

	if err = d.Initialize(); err != nil {
		return err
	}

	if err = d.Generate(); err != nil {
		return err
	}

	if err = d.Render(r.Config.DryRun); err != nil {
		return err
	}

	return nil
}

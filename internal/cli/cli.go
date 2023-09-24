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
	"strings"

	"github.com/tf2d2/tf2d2/internal/diagram"
	"github.com/tf2d2/tf2d2/internal/inframap"
	"github.com/tf2d2/tf2d2/internal/tfcloud"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Runtime represents the CLI execution runtime
type Runtime struct {
	ctx    context.Context
	cmd    *cobra.Command
	Config *config
}

// Config represents the CLI runtime config
type config struct {
	CfgFile      string `mapstructure:"-"`
	Hostname     string `mapstructure:"hostname"`
	Organization string `mapstructure:"organization"`
	Workspace    string `mapstructure:"workspace"`
	Token        string `mapstructure:"token"`
	OutputFile   string `mapstructure:"output-file"`
	StateFile    string `mapstructure:"state-file"`
	Verbose      bool   `mapstructure:"verbose"`
	DryRun       bool   `mapstructure:"dry-run"`
}

// NewRuntime returns a new CLI runtime instance
func NewRuntime(ctx context.Context) *Runtime {
	return &Runtime{
		cmd:    nil,
		ctx:    ctx,
		Config: &config{},
	}
}

// PreRunE executes before `RunE` to configure Viper and logging level
func (r *Runtime) PreRunE(cmd *cobra.Command, _ []string) error {
	logger := hclog.FromContext(r.ctx)

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
		logger.SetLevel(hclog.LevelFromString("debug"))
	}

	logger.Debug("successfully executed pre-run cli configuration")

	return nil
}

// RunE executes the logic to generate a d2 diagram from Terraform state
func (r *Runtime) RunE(_ *cobra.Command, _ []string) error {
	logger := hclog.FromContext(r.ctx)

	// read local or remote Terraform state
	var tfJsonState []byte
	if r.Config.Token == "" {
		bytes, err := os.ReadFile(filepath.Clean(r.Config.StateFile))
		if err != nil {
			return err
		}
		tfJsonState = bytes
		logger.Info("loaded local terraform state", "file", r.Config.StateFile)
	} else {
		tfe, err := tfcloud.NewTFClient(r.Config.Hostname, r.Config.Token)
		if err != nil {
			logger.Error("failed to initialize terraform cloud client", "error", err)
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
			logger.Error(err.Error())
			return err
		}
		logger.Info("loaded remote terraform state", "organization", r.Config.Organization, "workspace", r.Config.Workspace)
	}

	if err := r.generateContent(tfJsonState); err != nil {
		return err
	}

	return nil
}

// readConfig loads Viper config
func (r *Runtime) readConfig() error {
	logger := hclog.FromContext(r.ctx)

	if r.cmd.Flags().Changed("config") {
		viper.SetConfigFile(r.Config.CfgFile)
	} else {
		viper.SetConfigName(".tf2d2")
		viper.SetConfigType("yml")
	}

	// Search config in the following directories with name ".tf2d2" (without extension)
	viper.AddConfigPath("$HOME/") // home directory
	viper.AddConfigPath(".")      // current directory

	// read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("TF")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	var pathError *os.PathError
	var notFoundError viper.ConfigFileNotFoundError
	switch {
	case err != nil && errors.As(err, &pathError):
		logger.Warn("no config file found", "warning", err.Error())
	case err != nil && !errors.As(err, &notFoundError):
		// pathError and notFoundError are produced when no config file is found
		// here we check and return an error produced when reading the config file
		logger.Error("failed to read config", "error", err.Error())
		return err
	default:
		logger.Info("using config file", "path", viper.ConfigFileUsed())
	}

	return nil
}

// bindFlags binds all Cobra flags to Viper config
func (r *Runtime) bindFlags() error {
	fs := r.cmd.Flags()
	if err := viper.BindPFlags(fs); err != nil {
		return err
	}
	return nil
}

// generateContent uses Terraform state to generate and render a d2 diagram
func (r *Runtime) generateContent(state []byte) error {
	logger := hclog.FromContext(r.ctx)

	// generate output
	infraMap, err := inframap.GenerateInfraMap(r.ctx, state)
	if err != nil {
		logger.Error("error generating terraform infra map", "error", err)
		return err
	}

	d := diagram.NewDiagram(r.ctx, infraMap, filepath.Clean(r.Config.OutputFile))

	if err = d.Initialize(); err != nil {
		logger.Error("error initializing diagram", "error", err)
		return err
	}

	if err = d.Generate(r.Config.DryRun); err != nil {
		logger.Error("error generating diagram", "error", err)
		return err
	}

	logger.Info("successfully generated d2 diagram")

	return nil
}

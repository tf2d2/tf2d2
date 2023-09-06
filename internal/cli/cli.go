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

	"github.com/tf2d2/tf2d2/internal/diagram"
	"github.com/tf2d2/tf2d2/internal/inframap"
	"github.com/tf2d2/tf2d2/internal/tfcloud"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CliRuntime struct {
	cmd       *cobra.Command
	cliCtx    context.Context
	cliLogger hclog.Logger
	Config    *Config
}

type Config struct {
	CfgFile    string
	Host       string
	Org        string
	Workspace  string
	StateFile  string
	OutputFile string
}

func NewCliRuntime(ctx context.Context, logger hclog.Logger) *CliRuntime {
	return &CliRuntime{
		cmd:       &cobra.Command{},
		cliCtx:    ctx,
		cliLogger: logger,
		Config:    &Config{},
	}
}

func (r *CliRuntime) PreRunE(_ *cobra.Command, _ []string) error {
	verbose := viper.GetBool("verbose")
	if verbose {
		r.cliLogger.SetLevel(hclog.LevelFromString("debug"))
	}
	err := r.readConfig()
	if err != nil {
		return err
	}
	return nil
}

func (r *CliRuntime) RunE(_ *cobra.Command, _ []string) error {
	r.cliCtx = hclog.WithContext(r.cliCtx, r.cliLogger)
	vOrg := viper.GetString("org")
	vWorkspace := viper.GetString("workspace")
	vStateFile := viper.GetString("state-file")

	if vWorkspace != "" {
		tfe, err := tfcloud.NewTFClient(viper.GetString("host"), viper.GetString("token"))
		if err != nil {
			r.cliLogger.Error("failed to initialize Terraform Cloud client", "error", err.Error())
			return err
		}

		c := tfcloud.NewTFCloud(tfe)
		stateCommand := &tfcloud.GetStateVersion{
			Context:   r.cliCtx,
			TFCloud:   c,
			Workspace: vWorkspace,
		}
		stateCommand.Run(vOrg, vWorkspace)
	} else {
		if vStateFile != "" {
			m, err := inframap.GenerateInfraMap(r.cliCtx, viper.GetString("state-file"))
			if err != nil {
				r.cliLogger.Error("error generating terraform infra map", "error", err)
			}
			d := diagram.NewDiagram(r.cliCtx, m)
			if err = d.Initialize(); err != nil {
				return err
			}
			if err = d.Generate(); err != nil {
				return err
			}
			if err = d.Render(); err != nil {
				return err
			}
		} else {
			return errors.New("path to state file is empty")
		}
	}

	return nil
}

func (r *CliRuntime) readConfig() error {
	if r.Config.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(r.Config.CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			r.cliLogger.Error("failed to find $HOME directory", "error", err.Error())
			return err
		}

		// Search config in the following directories with name ".checkov-docs" (without extension)
		viper.AddConfigPath(home) // home directory
		viper.AddConfigPath(".")  // current directory
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tf2d2")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	var pathError *os.PathError
	var notFoundError viper.ConfigFileNotFoundError
	switch {
	case err != nil && errors.As(err, &pathError):
		r.cliLogger.Warn("no config file found", "warning", err.Error())
	case err != nil && !errors.As(err, &notFoundError):
		// pathError and notFoundError are produced when no config file is found
		// here we check and return an error produced when reading the config file
		r.cliLogger.Error("failed to read config", "error", err.Error())
		return err
	default:
		r.cliLogger.Info("using config file", "path", viper.ConfigFileUsed())
	}
	return nil
}

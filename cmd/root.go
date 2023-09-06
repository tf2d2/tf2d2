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

package cmd

import (
	"context"

	"github.com/tf2d2/tf2d2/internal/cli"
	"github.com/tf2d2/tf2d2/internal/logging"
	"github.com/tf2d2/tf2d2/internal/version"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cliCtx    context.Context
	cliLogger hclog.Logger
)

// Execute initializes, configures and runs 'root' command
func Execute() error {
	cliCtx = context.Background()
	cliLogger = logging.NewLogger()
	return NewCommand().Execute()
}

// NewCommand returns a new 'root' command
func NewCommand() *cobra.Command {
	r := cli.NewCliRuntime(cliCtx, cliLogger)
	cmd := &cobra.Command{
		Use:           "tf2d2",
		Short:         "Generate d2 diagrams from Terraform",
		Long:          "Generate d2 diagrams from Terraform",
		Annotations:   map[string]string{"command": "root"},
		Version:       version.GetVersion(),
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		PreRunE:       r.PreRunE,
		RunE:          r.RunE,
	}
	cmd.SetVersionTemplate("{{.Version}}")
	cmd.PersistentFlags().StringVarP(&r.Config.CfgFile, "config", "c", ".tf2d2.yaml", "Configuration file")
	cmd.PersistentFlags().StringVarP(&r.Config.Host, "host", "", "app.terraform.io", "Terraform API hostname")
	cmd.PersistentFlags().StringVarP(&r.Config.Org, "org", "", "", "Terraform organization")
	cmd.PersistentFlags().StringVarP(&r.Config.Workspace, "workspace", "w", "", "Terraform workspace")
	cmd.PersistentFlags().StringVarP(&r.Config.StateFile, "state-file", "", "", "Path of Terraform state file")
	cmd.PersistentFlags().StringVarP(&r.Config.OutputFile, "output-file", "o", "", "Path of output file")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Show debug output")
	cmd.PersistentFlags().Bool("dry-run", false, "Only print generated output")
	cobra.CheckErr(viper.BindPFlag("host", cmd.PersistentFlags().Lookup("host")))
	cobra.CheckErr(viper.BindPFlag("org", cmd.PersistentFlags().Lookup("org")))
	cobra.CheckErr(viper.BindPFlag("workspace", cmd.PersistentFlags().Lookup("workspace")))
	cobra.CheckErr(viper.BindPFlag("state-file", cmd.PersistentFlags().Lookup("state-file")))
	cobra.CheckErr(viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose")))
	cobra.CheckErr(viper.BindPFlag("dry-run", cmd.PersistentFlags().Lookup("dry-run")))

	return cmd
}

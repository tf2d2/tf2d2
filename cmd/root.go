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
)

// Execute initializes, configures and runs 'root' command
func Execute(ctx context.Context) error {
	return NewCommand(ctx).Execute()
}

// NewCommand returns a new 'root' command
func NewCommand(ctx context.Context) *cobra.Command {
	cmdLogger := logging.NewLogger()
	cmdCtx := hclog.WithContext(ctx, cmdLogger)
	r := cli.NewCliRuntime(cmdCtx)
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
	cmd.PersistentFlags().StringVarP(&r.Config.CfgFile, "config", "c", ".tf2d2.yml", "config file name")
	cmd.PersistentFlags().StringVarP(&r.Config.Hostname, "hostname", "", "app.terraform.io", "hostname of terraform organization")
	cmd.PersistentFlags().StringVarP(&r.Config.Organization, "organization", "", "", "name of terraform organization")
	cmd.PersistentFlags().StringVarP(&r.Config.Workspace, "workspace", "w", "", "name of terraform workspace")
	cmd.PersistentFlags().StringVarP(&r.Config.Token, "token", "t", "", "terraform api token")
	cmd.PersistentFlags().StringVarP(&r.Config.StateFile, "state-file", "f", "", "file path of terraform state")
	cmd.PersistentFlags().StringVarP(&r.Config.OutputFile, "output-file", "o", "tf2d2.svg", "file path of output diagram in .svg or .png format")
	cmd.PersistentFlags().BoolVarP(&r.Config.Verbose, "verbose", "v", false, "show debug output")
	cmd.PersistentFlags().BoolVarP(&r.Config.DryRun, "dry-run", "", false, "only print output")

	return cmd
}

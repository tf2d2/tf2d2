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

package version

import (
	_ "embed"

	goversion "github.com/caarlos0/go-version"
)

var (
	version = ""
	commit  = ""
	date    = ""
	builtBy = ""
)

//go:embed art.txt
var asciiArt string

const projectURL = "https://github.com/tf2d2/tf2d2"

// GetVersion returns the latest version including runtime GOOS and GOARCH and more.
func GetVersion() string {
	v := goversion.GetVersionInfo(
		goversion.WithAppDetails("tf2d2", "Generate d2 diagrams from Terraform", projectURL),
		goversion.WithASCIIName(asciiArt),
		goversion.WithBuiltBy(builtBy),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if version != "" {
				i.GitVersion = version
			}
			if date != "" {
				i.BuildDate = date
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)

	return v.String()
}

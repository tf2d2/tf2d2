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

package logging

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

const envLogLevel = "TF_LOG"

// NewLogger returns a new logger
func NewLogger() hclog.Logger {
	var logLevel string
	envLevel := os.Getenv(envLogLevel)
	if envLevel != "" {
		logLevel = envLevel
	} else {
		logLevel = "info"
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Name:                     "tf2d2",
		Level:                    hclog.LevelFromString(logLevel),
		Output:                   nil,
		Mutex:                    nil,
		JSONFormat:               false,
		IncludeLocation:          false,
		AdditionalLocationOffset: 0,
		TimeFormat:               "",
		DisableTime:              false,
		Color:                    hclog.AutoColor,
		ColorHeaderOnly:          false,
		ColorHeaderAndFields:     false,
		IndependentLevels:        false,
	})
	return logger
}

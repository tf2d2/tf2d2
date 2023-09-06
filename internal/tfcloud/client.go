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

package tfcloud

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
)

const (
	defaultHostname = "app.terraform.io"
	baseUserAgent   = "tf2d2"
)

// NewTFClient returns a new Terraform API client
func NewTFClient(host, token string) (*tfe.Client, error) {
	tfeConfig := tfe.DefaultConfig()

	hostname := host
	if hostname == "" {
		hostname = defaultHostname
	}

	apiToken := token
	if apiToken == "" {
		apiTokenEnv := os.Getenv("TF_API_TOKEN")
		if apiTokenEnv != "" {
			apiToken = apiTokenEnv
		}
	}

	tfeConfig.Headers.Set("User-Agent", baseUserAgent)
	tfeConfig.Address = fmt.Sprintf("https://%s", hostname)
	tfeConfig.Token = apiToken

	if tfeConfig.Token == "" {
		return nil, fmt.Errorf("Terraform API token is not set")
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return nil, err
	}
	client.RetryServerErrors(true)

	return client, nil
}

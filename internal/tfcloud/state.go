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
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hashicorp/go-hclog"
)

// GetStateVersion allows to retrieve the state version of a Terraform workspace
type GetStateVersion struct {
	Context      context.Context
	TFCloud      *TFCloud
	Organization string
	Workspace    string
}

// Run retrieves the state version of a Terraform workspace
func (c *GetStateVersion) Run() (json.RawMessage, error) {
	logger := hclog.FromContext(c.Context)

	stateVersionRes, err := c.TFCloud.StateVersionService.GetState(c.Context, c.Organization, c.Workspace)
	if err != nil {
		logger.Error("error getting terraform state version", "error", err)
		return nil, err
	}

	stateDownloadReq, err := http.NewRequestWithContext(c.Context, http.MethodGet, stateVersionRes.DownloadURL, nil)
	if err != nil {
		logger.Error("error downloading json state", "error", err)
		return nil, err
	}

	stateDownloadRes, err := http.DefaultClient.Do(stateDownloadReq)
	if err != nil {
		logger.Error("error making http request to download json state", "error", err)
		return nil, err
	}
	defer func() {
		err = stateDownloadRes.Body.Close()
		if err != nil {
			logger.Error("error closing download response body", "error", err)
		}
	}()

	stateJSON, err := io.ReadAll(stateDownloadRes.Body)
	if err != nil {
		logger.Error("error reading state download response body", "error", err)
		return nil, err
	}

	return stateJSON, nil
}

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
	"io"
	"net/http"

	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
)

// GetStateVersion allows to retrieve the state version of a Terraform workspace
type GetStateVersion struct {
	Context   context.Context
	TFCloud   *TFCloud
	Workspace string
}

// Run retrieves the state version of a Terraform workspace
func (c *GetStateVersion) Run(orgName string, workspaceName string) (*tfjson.State, error) {
	logger := hclog.FromContext(c.Context)

	stateVersionRes, err := c.TFCloud.StateVersionService.GetState(c.Context, orgName, workspaceName)
	if err != nil {
		logger.Error("error getting Terraform state version", err.Error())
		return nil, err
	}
	logger.Info("successfully retrieved state version", "download-url", stateVersionRes.JSONDownloadURL)

	stateDownloadReq, err := http.NewRequestWithContext(c.Context, http.MethodGet, stateVersionRes.JSONDownloadURL, nil)
	if err != nil {
		logger.Error("error downloading json state", "error", err.Error())
		return nil, err
	}

	stateDownloadRes, err := http.DefaultClient.Do(stateDownloadReq)
	if err != nil {
		logger.Error("error making http request to download json state", "error", err.Error())
		return nil, err
	}
	defer func() {
		err = stateDownloadRes.Body.Close()
		if err != nil {
			logger.Error("error closing download response body", "error", err.Error())
		}
	}()

	stateJSON, err := io.ReadAll(stateDownloadRes.Body)
	if err != nil {
		logger.Error("error reading state download response body", "error", err.Error())
		return nil, err
	}

	var state tfjson.State

	err = state.UnmarshalJSON(stateJSON)
	if err != nil {
		logger.Error("error unmarshalling state", "error", err)
		return nil, err
	}

	return &state, nil
}

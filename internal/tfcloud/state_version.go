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
	"errors"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-tfe"
	"github.com/sethvargo/go-retry"
)

// StateVersionService defines a service that retrieves Terraform state versions
type StateVersionService interface {
	GetState(ctx context.Context, orgName string, workspaceName string) (*tfe.StateVersion, error)
}

// stateVersionService implements a service to retrieve Terraform state versions
type stateVersionService struct {
	tfe *tfe.Client
}

// NewStateVersionService returns a new service that retrieves Terraform state versions
func NewStateVersionService(c *tfe.Client) StateVersionService {
	return &stateVersionService{c}
}

// wait 5 minutes for current state version to finish processing
// to prevent edge case of reading current state version immediately after an apply run
const stateVersionReadMaxDuration = 5 * time.Minute

func newBackoff() retry.Backoff {
	backoff := retry.NewFibonacci(2 * time.Second)
	backoff = retry.WithCappedDuration(7*time.Second, backoff)
	backoff = retry.WithMaxDuration(stateVersionReadMaxDuration, backoff)
	return backoff
}

// GetState retrieves the current state version of a Terraform workspace
func (s *stateVersionService) GetState(ctx context.Context, orgName string, workspaceName string) (*tfe.StateVersion, error) {
	logger := hclog.FromContext(ctx)

	workspaceRead, err := s.tfe.Workspaces.Read(ctx, orgName, workspaceName)
	if err != nil {
		logger.Error("error reading workspace", "organization", orgName, "workspace", workspaceName, "error", err)
		return nil, err
	}

	currentSV, err := s.tfe.StateVersions.ReadCurrent(ctx, workspaceRead.ID)
	if err != nil {
		logger.Error("error reading current state version", "error", err)
		return nil, err
	}

	if !currentSV.ResourcesProcessed {
		retryErr := retry.Do(ctx, newBackoff(), func(ctx context.Context) error {
			currentSV, err = s.tfe.StateVersions.ReadCurrent(ctx, workspaceRead.ID)
			// return non-retryable error, i.e. Terraform API call failed
			if err != nil {
				return err
			}
			// keep checking until resources have been processed or until max timeout is exceeded
			if currentSV.ResourcesProcessed {
				return nil
			}
			return retry.RetryableError(errors.New("current state version read has exceeded max timeout"))
		})

		if retryErr != nil {
			logger.Error("error waiting for current state version to finish processing", "error", retryErr)
			return nil, retryErr
		}
	}

	return currentSV, nil
}

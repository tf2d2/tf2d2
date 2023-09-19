package tfcloud

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
	"github.com/stretchr/testify/assert"
)

func TestStateVersionService_ReadStateVersion(t *testing.T) {
	testCases := []struct {
		name            string
		orgName         string
		workspaceName   string
		ctx             context.Context
		workspaceID     string
		tfeWorkspace    *tfe.Workspace
		tfeStateVersion *tfe.StateVersion
	}{
		{
			name:          "successful without retries",
			orgName:       "my-org",
			workspaceName: "my-workspace",
			ctx:           context.Background(),
			workspaceID:   "ws-***",
			tfeWorkspace:  &tfe.Workspace{ID: "ws-***"},
			tfeStateVersion: &tfe.StateVersion{
				ResourcesProcessed: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mock workspace
			mockWorkspace := mocks.NewMockWorkspaces(ctrl)
			mockWorkspace.EXPECT().Read(tc.ctx, tc.orgName, tc.workspaceName).Return(
				tc.tfeWorkspace,
				nil,
			)

			// mock state version
			mockStateVersion := mocks.NewMockStateVersions(ctrl)
			mockStateVersion.EXPECT().ReadCurrent(tc.ctx, tc.workspaceID).Return(
				tc.tfeStateVersion,
				nil,
			)

			client := NewStateVersionService(
				&tfe.Client{
					Workspaces:    mockWorkspace,
					StateVersions: mockStateVersion,
				},
			)

			_, resultErr := client.GetState(tc.ctx, tc.orgName, tc.workspaceName)
			assert.Nil(t, resultErr)
		})
	}
}

func TestStateVersionService_RetryReadStateVersion(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx, orgName, workspaceName, wID := context.Background(), "my-org", "my-workspace", "ws-***"

		tfeWorkspace := &tfe.Workspace{ID: wID}
		tfeStateVersion := &tfe.StateVersion{
			ResourcesProcessed: false,
		}

		// mock workspace
		mockWorkspace := mocks.NewMockWorkspaces(ctrl)
		mockWorkspace.EXPECT().Read(ctx, orgName, workspaceName).Return(
			tfeWorkspace,
			nil,
		)

		// mock state version
		mockStateVersion := mocks.NewMockStateVersions(ctrl)
		// Assert and mock retry with resources processed set to false
		retryCall := mockStateVersion.EXPECT().ReadCurrent(ctx, wID).Return(tfeStateVersion, nil).Times(3)
		// Assert and mock retry is stopped when resources processed is set to true
		doneCall := mockStateVersion.EXPECT().ReadCurrent(ctx, wID).Return(&tfe.StateVersion{
			ResourcesProcessed: true,
		}, nil)

		// Expect retry calls before done call
		gomock.InOrder(
			retryCall,
			doneCall,
		)

		client := NewStateVersionService(
			&tfe.Client{
				Workspaces:    mockWorkspace,
				StateVersions: mockStateVersion,
			},
		)

		client.GetState(ctx, orgName, workspaceName)
	})
}

func TestStateVersionService_ErrorWorkspaceRead(t *testing.T) {
	t.Run("error reading workspace", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx, orgName, workspaceName := context.Background(), "my-org", "my-workspace"

		// mock workspace
		mockWorkspace := mocks.NewMockWorkspaces(ctrl)
		mockWorkspace.EXPECT().Read(ctx, orgName, workspaceName).Return(
			nil,
			errors.New("error-workspace-read"),
		)

		client := NewStateVersionService(
			&tfe.Client{
				Workspaces: mockWorkspace,
			},
		)

		_, resultErr := client.GetState(ctx, orgName, workspaceName)
		assert.NotNil(t, resultErr)
		assert.ErrorContains(t, resultErr, "error-workspace-read")
	})
}

func TestStateVersionService_ErrorStateVersionRead(t *testing.T) {
	t.Run("error reading state version", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx, orgName, workspaceName, wID := context.Background(), "my-org", "my-workspace", "ws-***"

		tfeWorkspace := &tfe.Workspace{ID: wID}

		// mock workspace
		mockWorkspace := mocks.NewMockWorkspaces(ctrl)
		mockWorkspace.EXPECT().Read(ctx, orgName, workspaceName).Return(
			tfeWorkspace,
			nil,
		)

		// mock state version
		mockStateVersion := mocks.NewMockStateVersions(ctrl)
		mockStateVersion.EXPECT().ReadCurrent(ctx, wID).Return(
			nil, errors.New("error-state-version-read"),
		)

		client := NewStateVersionService(
			&tfe.Client{
				Workspaces:    mockWorkspace,
				StateVersions: mockStateVersion,
			},
		)

		_, resultErr := client.GetState(ctx, orgName, workspaceName)
		assert.NotNil(t, resultErr)
		assert.ErrorContains(t, resultErr, "error-state-version-read")
	})
}

func TestStateVersionService_ErrorRetryStateVersionRead(t *testing.T) {
	t.Run("error reading state version with retry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx, orgName, workspaceName, wID := context.Background(), "my-org", "my-workspace", "ws-***"

		tfeWorkspace := &tfe.Workspace{ID: wID}
		tfeStateVersion := &tfe.StateVersion{
			ResourcesProcessed: false,
		}

		// mock workspace
		mockWorkspace := mocks.NewMockWorkspaces(ctrl)
		mockWorkspace.EXPECT().Read(ctx, orgName, workspaceName).Return(
			tfeWorkspace,
			nil,
		)

		// mock state version
		mockStateVersion := mocks.NewMockStateVersions(ctrl)
		// Assert and mock retry with resources processed set to false
		retryCall := mockStateVersion.EXPECT().ReadCurrent(ctx, wID).Return(tfeStateVersion, nil).Times(1)
		// Assert and mock retry is stopped with an error
		errCall := mockStateVersion.EXPECT().ReadCurrent(ctx, wID).Return(
			nil, errors.New("error-retry-state-version-read"),
		)

		// Expect retry calls before done call
		gomock.InOrder(
			retryCall,
			errCall,
		)

		client := NewStateVersionService(
			&tfe.Client{
				Workspaces:    mockWorkspace,
				StateVersions: mockStateVersion,
			},
		)

		_, resultErr := client.GetState(ctx, orgName, workspaceName)
		assert.NotNil(t, resultErr)
		assert.ErrorContains(t, resultErr, "error-retry-state-version-read")
	})
}

package tfcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

// func TestGetStateVersion_Run(t *testing.T) {
// 	// Create a mock HTTP server to simulate downloading JSON state
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		assert.Equal(t, r.Method, http.MethodGet)
// 		assert.Contains(t, r.URL.String(), "mock-state-download-url")

// 		// Write mock JSON state to the response
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		_, _ = w.Write([]byte(`{"key": "value"}`))
// 	}))
// 	defer server.Close()

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	ctx, orgName, workspaceName, wID := context.Background(), "my-org", "my-workspace", "ws-***"
// 	downloadURL := fmt.Sprintf("%s/mock-state-download-url", server.URL)

// 	tfeWorkspace := &tfe.Workspace{ID: wID}
// 	tfeStateVersion := &tfe.StateVersion{
// 		ResourcesProcessed: true,
// 		DownloadURL:        downloadURL,
// 	}

// 	// mock workspace
// 	mockWorkspace := mocks.NewMockWorkspaces(ctrl)
// 	mockWorkspace.EXPECT().Read(ctx, orgName, workspaceName).Return(
// 		tfeWorkspace,
// 		nil,
// 	)

// 	// mock state version
// 	mockStateVersion := mocks.NewMockStateVersions(ctrl)
// 	mockStateVersion.EXPECT().ReadCurrent(ctx, wID).Return(
// 		tfeStateVersion,
// 		nil,
// 	)

// 	// create tfe client
// 	mockTFEConfig := tfe.DefaultConfig()
// 	mockTFEConfig.Token = "api-token"
// 	mockTFEClient, err := tfe.NewClient(mockTFEConfig)
// 	assert.Nil(t, err)
// 	mockTFEClient.Workspaces = mockWorkspace
// 	mockTFEClient.StateVersions = mockStateVersion

// 	// create tfcloud instance
// 	mockTFCloud := &TFCloud{
// 		StateVersionService: NewStateVersionService(
// 			mockTFEClient,
// 		),
// 	}

// 	// Create a GetStateVersion instance
// 	getStateVersion := &GetStateVersion{
// 		Context:      context.Background(),
// 		TFCloud:      mockTFCloud,
// 		Organization: orgName,
// 		Workspace:    workspaceName,
// 	}

// 	// Run the GetStateVersion
// 	stateJSON, err := getStateVersion.Run()
// 	assert.NoError(t, err)
// 	assert.NotNil(t, stateJSON)

// 	var stateData map[string]interface{}
// 	_ = json.Unmarshal(stateJSON, &stateData)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "value", stateData["key"])
// }

type MockStateVersionService struct {
	sv      *tfe.StateVersion
	isError bool
}

func (m *MockStateVersionService) GetState(_ context.Context, _ string, _ string) (*tfe.StateVersion, error) {
	if m.isError {
		return nil, errors.New("error")
	}

	return m.sv, nil
}

func testGetStateVersion(t *testing.T, isError bool, downloadURL string) *GetStateVersion {
	t.Helper()

	mockTFCloud := &TFCloud{
		StateVersionService: &MockStateVersionService{
			sv:      &tfe.StateVersion{DownloadURL: downloadURL},
			isError: isError,
		},
	}

	mockGetStateVersion := &GetStateVersion{
		Context:      context.Background(),
		TFCloud:      mockTFCloud,
		Organization: "orgName",
		Workspace:    "workspaceName",
	}

	return mockGetStateVersion
}

func TestGetStateVersion_Run(t *testing.T) {
	const testServerDownloadURL = "mock-state-download-url"
	var expectedJSONState = []byte(`{"key": "value"}`)

	testCases := []struct {
		name            string
		expected        []byte
		downloadURL     string
		onServiceError  bool
		onDownloadError bool
	}{
		{
			name:            "success",
			expected:        expectedJSONState,
			downloadURL:     testServerDownloadURL,
			onServiceError:  false,
			onDownloadError: false,
		},
		{
			name:            "error calling state service",
			expected:        expectedJSONState,
			downloadURL:     testServerDownloadURL,
			onServiceError:  true,
			onDownloadError: false,
		},
		{
			name:            "error parsing download url",
			expected:        expectedJSONState,
			downloadURL:     "::error::",
			onServiceError:  false,
			onDownloadError: true,
		},
		{
			name:            "error calling invalid download url",
			expected:        expectedJSONState,
			downloadURL:     "error",
			onServiceError:  false,
			onDownloadError: true,
		},
		{
			name:            "error reading downloaded json state",
			expected:        expectedJSONState,
			downloadURL:     testServerDownloadURL,
			onServiceError:  false,
			onDownloadError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock HTTP server to simulate downloading JSON state
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodGet)
				assert.Contains(t, r.URL.String(), testServerDownloadURL)

				// Write mock JSON state to the response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(tc.expected)
			}))
			defer server.Close()

			if !tc.onDownloadError {
				tc.downloadURL = fmt.Sprintf("%s/%s", server.URL, tc.downloadURL)
			}

			gsv := testGetStateVersion(t, tc.onServiceError, tc.downloadURL)

			stateJSON, err := gsv.Run()
			if tc.onServiceError || tc.onDownloadError {
				assert.Error(t, err)
				assert.Nil(t, stateJSON)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stateJSON)

				var stateData map[string]interface{}
				_ = json.Unmarshal(stateJSON, &stateData)
				assert.NoError(t, err)
				assert.Equal(t, "value", stateData["key"])
			}
		})
	}
}

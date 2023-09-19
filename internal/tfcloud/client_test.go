package tfcloud

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTFClient(t *testing.T) {
	testCases := []struct {
		name          string
		host          string
		token         string
		isEnvToken    bool
		expectedError bool
	}{
		{name: "new client with input default host and token", host: "app.terraform.io", token: "api_token", isEnvToken: false, expectedError: false},
		{name: "new client with input token", host: "", token: "api_token", isEnvToken: false, expectedError: false},
		{name: "new client with input token as env var", host: "", token: "", isEnvToken: true, expectedError: false},
		{name: "new client without input token", host: "", token: "", isEnvToken: false, expectedError: true},
		{name: "new client with input non-default host", host: "custom-host", token: "", isEnvToken: false, expectedError: true},
		{name: "new client with input non-default host and token", host: "custom-host", token: "api_token", isEnvToken: false, expectedError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for TF_API_TOKEN
			if tc.isEnvToken {
				os.Setenv("TF_API_TOKEN", "api_token")
				defer os.Unsetenv("TF_API_TOKEN")
			}

			client, err := NewTFClient(tc.host, tc.token)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

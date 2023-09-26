package tfcloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTFClient(t *testing.T) {
	testCases := []struct {
		name    string
		host    string
		token   string
		isError bool
	}{
		{name: "default host with token", host: "", token: "api_token", isError: false},
		{name: "default host without token", host: "", token: "", isError: true},
		{name: "custom host no token", host: "custom-host", token: "", isError: true},
		{name: "custom host with token", host: "custom-host", token: "api_token", isError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewTFClient(tc.host, tc.token)

			if tc.isError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

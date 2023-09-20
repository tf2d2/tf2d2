package inframap

import (
	"context"
	"testing"

	"github.com/tf2d2/tf2d2/internal/utils"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInfraMap_Success(t *testing.T) {
	ctx := context.Background()

	mockState, err := utils.GetExpected("valid.golden")
	assert.Nil(t, err)

	logger := hclog.New(&hclog.LoggerOptions{})
	ctx = hclog.WithContext(ctx, logger)

	infraMap, err := GenerateInfraMap(ctx, []byte(mockState))
	assert.NoError(t, err)
	assert.NotNil(t, infraMap)
}

func TestGenerateInfraMap_Error(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{
			name:     "invalid state",
			filename: "invalid.golden",
		},
		{
			name:     "invalid version",
			filename: "invalid_version.golden",
		},
		{
			name:     "missing version",
			filename: "missing_version.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockState, err := utils.GetExpected(tc.filename)
			assert.Nil(t, err)

			logger := hclog.New(&hclog.LoggerOptions{})
			ctx = hclog.WithContext(ctx, logger)

			infraMap, err := GenerateInfraMap(ctx, []byte(mockState))
			assert.Error(t, err)
			assert.Nil(t, infraMap)
		})
	}

}

// func TestGenerateInfraMap_ErrorGeneratingMap(t *testing.T) {
// 	// Create a mock context
// 	ctx := context.Background()

// 	// Create a mock Terraform state
// 	mockState := []byte("mocked-terraform-state")

// 	// Create a logger with debugging enabled
// 	logger := hclog.New(&hclog.LoggerOptions{
// 		Output:     nil,
// 		Level:      hclog.Debug,
// 		JSONFormat: false,
// 	})

// 	// Set the logger as the logger for the context
// 	ctx = hclog.WithContext(ctx, logger)

// 	// Call the GenerateInfraMap function
// 	infraMap, err := GenerateInfraMap(ctx, mockState)

// 	// Assertions
// 	assert.Error(t, err)
// 	assert.Nil(t, infraMap)
// 	assert.Contains(t, err.Error(), "error generating infra map")
// }

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

	mockState, err := utils.GetExpected("valid.expected")
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
			filename: "invalid.expected",
		},
		{
			name:     "invalid version",
			filename: "invalid_version.expected",
		},
		{
			name:     "missing version",
			filename: "missing_version.expected",
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

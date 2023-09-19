package tfcloud

import (
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

func TestNewTFCloud(t *testing.T) {
	// Create a new TFCloud instance
	cloud := NewTFCloud(&tfe.Client{})

	// Assertions
	assert.NotNil(t, cloud)
	assert.NotNil(t, cloud.StateVersionService)

	// Ensure that StateVersionService is of the correct type
	_, result := cloud.StateVersionService.(*stateVersionService)
	assert.True(t, result)
}

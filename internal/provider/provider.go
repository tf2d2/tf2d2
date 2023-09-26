package provider

import (
	"strings"

	"github.com/tf2d2/tf2d2/internal/provider/aws"
	"github.com/tf2d2/tf2d2/internal/provider/azurerm"
	"github.com/tf2d2/tf2d2/internal/provider/google"
)

// ValidateResource checks that a provider resource is a valid node
// which can be included in the output diagram
func ValidateResource(name string) bool {
	provider := strings.Split(name, "_")[0]

	var result bool
	switch provider {
	case "aws":
		if _, ok := aws.Nodes[name]; ok {
			result = ok
		}
	case "azurerm":
		if _, ok := azurerm.Nodes[name]; ok {
			result = ok
		}
	case "google":
		if _, ok := google.Nodes[name]; ok {
			result = ok
		}
	}

	return result
}

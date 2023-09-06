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

package inframap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cycloidio/inframap/generate"
	"github.com/cycloidio/inframap/graph"
	"github.com/hashicorp/go-hclog"
)

type TFInfraMap struct {
	Graph     *graph.Graph
	GraphDesc map[string]interface{}
}

// GenerateInfraMap generates an infra map from Terraform state
func GenerateInfraMap(ctx context.Context, filename string) (*TFInfraMap, error) {
	logger := hclog.FromContext(ctx)
	logger.Info("generate terraform infra map")

	// open and read the Terraform state JSON file
	stateFile, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}

	opt := generate.Options{
		Raw:           true,
		Clean:         true,
		Connections:   true,
		ExternalNodes: true,
	}

	g, gDesc, err := generate.FromState(stateFile, opt)
	if err != nil {
		return nil, err
	}

	tfinframap := &TFInfraMap{
		Graph:     g,
		GraphDesc: gDesc,
	}

	logger.Info(fmt.Sprintf("total nodes: %d", len(g.Nodes)))
	logger.Info(fmt.Sprintf("total edges: %d", len(g.Edges)))

	return tfinframap, nil
}

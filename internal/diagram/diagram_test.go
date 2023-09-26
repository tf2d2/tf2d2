package diagram

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/tf2d2/tf2d2/internal/utils"

	"github.com/cycloidio/inframap/graph"
	"github.com/cycloidio/tfdocs/resource"
	"github.com/stretchr/testify/assert"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2lib"
)

func testGetMockDiagram(t *testing.T, mockGraph *graph.Graph, outputFile string) *Diagram {
	t.Helper()

	ctx := context.Background()

	if mockGraph == nil {
		mockGraph = &graph.Graph{}
	}

	mockDiagram := NewDiagram(ctx, mockGraph, outputFile)
	assert.NotNil(t, mockDiagram)
	assert.Equal(t, ctx, mockDiagram.ctx)
	assert.Equal(t, outputFile, mockDiagram.Filepath)
	assert.Equal(t, mockGraph, mockDiagram.TFInfraMap)
	assert.Nil(t, mockDiagram.d2Diagram)
	assert.Nil(t, mockDiagram.d2Graph)
	assert.Nil(t, mockDiagram.d2CompileOpts)
	assert.Nil(t, mockDiagram.d2RenderOpts)

	err := mockDiagram.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, mockDiagram.d2Diagram)
	assert.NotNil(t, mockDiagram.d2Graph)
	assert.NotNil(t, mockDiagram.d2CompileOpts)
	assert.NotNil(t, mockDiagram.d2RenderOpts)

	return mockDiagram
}

func TestGenerate_Success(t *testing.T) {
	mockGraph := graph.New()
	mockGraph.AddNode(
		&graph.Node{
			ID:        "node_foo",
			Canonical: "node_foo",
			Resource: resource.Resource{
				Name: "aws_ecs_cluster",
			},
		},
	)
	mockGraph.AddNode(
		&graph.Node{
			ID:        "node_bar",
			Canonical: "node_bar",
			Resource: resource.Resource{
				Name: "aws_ecs_cluster",
			},
		},
	)
	mockGraph.AddEdge(
		&graph.Edge{
			ID:     "edge_baz",
			Source: "node_foo",
			Target: "node_bar",
		},
	)
	outputScript, outputDiagram := "output.d2", "output.svg"
	expectedScript, err := utils.GetTestData("script.golden")
	assert.NoError(t, err)
	d := testGetMockDiagram(t, mockGraph, outputDiagram)
	err = d.Generate(false)
	assert.NoError(t, err)

	outScript, err := os.ReadFile(outputScript)
	assert.NoError(t, err)
	assert.NotEmpty(t, outScript)
	assert.Equal(t, expectedScript, string(outScript))

	outDiagram, err := os.ReadFile(outputDiagram)
	assert.NoError(t, err)
	assert.NotEmpty(t, outDiagram)
	assert.Contains(t, string(outDiagram), "node_foo")
	assert.Contains(t, string(outDiagram), "node_bar")

	// remove output files
	_ = os.Remove(outputScript)
	_ = os.Remove(outputDiagram)
}

func TestGenerate_SkipSourceNodes(t *testing.T) {
	mockGraph := graph.New()
	err := mockGraph.AddNode(
		&graph.Node{ID: "node_foo", Canonical: "node_foo", Resource: resource.Resource{Name: "aws_ecs_cluster"}},
	)
	assert.NoError(t, err)
	err = mockGraph.AddNode(
		&graph.Node{ID: "node_bar", Canonical: "node_bar", Resource: resource.Resource{Name: "aws_ecs_cluster"}},
	)
	assert.NoError(t, err)
	err = mockGraph.AddNode(
		&graph.Node{ID: "node_foo_skip", Canonical: "node_foo_skip", Resource: resource.Resource{Name: "aws_iam_policy"}},
	)
	assert.NoError(t, err)
	err = mockGraph.AddEdge(&graph.Edge{ID: "edge_baz", Source: "node_foo", Target: "node_bar"})
	assert.NoError(t, err)
	err = mockGraph.AddEdge(&graph.Edge{ID: "edge_skip", Source: "node_foo_skip", Target: "node_bar"})
	assert.NoError(t, err)

	testCases := []struct {
		name        string
		isTargetErr bool
	}{
		{
			name:        "skip edge source node",
			isTargetErr: false,
		},
		{
			name:        "skip edge target node",
			isTargetErr: true,
		},
	}
	outputScript, outputDiagram := "output.d2", "output.svg"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isTargetErr {
				mockGraph.InvertEdge("edge_skip")
			}

			expectedScript, err := utils.GetTestData("script.golden")
			assert.NoError(t, err)

			d := testGetMockDiagram(t, mockGraph, outputDiagram)
			err = d.Generate(false)
			assert.NoError(t, err)

			outScript, err := os.ReadFile(outputScript)
			assert.NoError(t, err)
			assert.NotEmpty(t, outScript)
			assert.Equal(t, expectedScript, string(outScript))

			outDiagram, err := os.ReadFile(outputDiagram)
			assert.NoError(t, err)
			assert.NotEmpty(t, outDiagram)
			assert.Contains(t, string(outDiagram), "node_foo")
			assert.Contains(t, string(outDiagram), "node_bar")

			// remove output files
			_ = os.Remove(outputScript)
			_ = os.Remove(outputDiagram)
		})
	}

}

func TestGenerate_Errors(t *testing.T) {
	sourceNodeErrGraph := graph.New()
	sourceNodeErrGraph.Nodes = []*graph.Node{
		{
			ID:        "foo",
			Canonical: "foo",
			Resource: resource.Resource{
				Name: "foo",
			},
		},
	}
	sourceNodeErrGraph.Edges = []*graph.Edge{
		{
			ID:     "invalid",
			Source: "invalid",
			Target: "invalid",
		},
	}

	targetNodeErrGraph := graph.New()
	err := targetNodeErrGraph.AddNode(&graph.Node{
		ID:        "foo",
		Canonical: "foo",
		Resource:  resource.Resource{Name: "foo"},
	})
	assert.NoError(t, err)
	err = targetNodeErrGraph.AddNode(&graph.Node{
		ID:        "bar",
		Canonical: "bar",
		Resource:  resource.Resource{Name: "bar"},
	})
	assert.NoError(t, err)
	targetNodeErrGraph.Edges = []*graph.Edge{
		{
			Source: "foo",
			Target: "invalid",
		},
	}

	testCases := []struct {
		name     string
		expected string
		graph    *graph.Graph
	}{
		// {
		// 	name:     "resource icon not found",
		// 	expected: "resource \"\" not found",
		// 	graph: &graph.Graph{
		// 		Nodes: []*graph.Node{
		// 			{
		// 				ID:        "aws_invalid_resource",
		// 				Canonical: "aws_invalid_resource",
		// 				Resource:  resource.Resource{Name: "aws_invalid_resource"},
		// 			},
		// 		},
		// 	},
		// },
		{
			name:     "empty terraform inframap",
			expected: "no shapes found",
			graph:    &graph.Graph{},
		},
		{
			name:     "source node not found",
			expected: "error getting source node: invalid",
			graph:    sourceNodeErrGraph,
		},
		{
			name:     "target node not found",
			expected: "error getting target node: invalid",
			graph:    targetNodeErrGraph,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := testGetMockDiagram(t, tc.graph, "output.svg")

			err = d.Generate(true)
			assert.Error(t, err)
			assert.EqualError(t, err, tc.expected)
		})
	}
}

func TestWrite_FileOutput(t *testing.T) {
	outExpected, outputScript, outputSVG, outputPNG := "a -> b\n", "output.d2", "output.svg", "output.png"
	testCases := []struct {
		name     string
		filename string
		expected string
		isError  bool
	}{
		{
			name:     "empty d2 diagram",
			expected: "",
			filename: outputSVG,
			isError:  false,
		},
		{
			name:     "svg d2 diagram",
			expected: outExpected,
			filename: outputSVG,
			isError:  false,
		},
		{
			name:     "png d2 diagram",
			expected: outExpected,
			filename: outputPNG,
			isError:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := testGetMockDiagram(t, nil, tc.filename)

			// create mock d2 graph from script
			var err error
			d.d2Diagram, d.d2Graph, err = d2lib.Compile(d.ctx, tc.expected, d.d2CompileOpts, d.d2RenderOpts)
			assert.NoError(t, err)

			// Verify output files are created for d2 script and diagram
			err = d.Write(false)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, d2format.Format(d.d2Graph.AST))

			outScript, err := os.ReadFile(outputScript)
			assert.NoError(t, err)
			assert.Equal(t, string(outScript), tc.expected)

			outDiagram, err := os.ReadFile(tc.filename)
			assert.NoError(t, err)
			assert.NotEqual(t, outDiagram, "")

			// remove output files
			_ = os.Remove(outputScript)
			_ = os.Remove(tc.filename)
		})
	}
}

func TestWrite_DryRun(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		isError  bool
	}{
		{
			name:     "empty d2 graph",
			expected: "",
			isError:  false,
		},
		{
			name:     "non-empty d2 graph",
			expected: "a -> b\n",
			isError:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := testGetMockDiagram(t, nil, "")

			// create mock d2 graph from script
			_, d.d2Graph, _ = d2lib.Compile(d.ctx, tc.expected, d.d2CompileOpts, d.d2RenderOpts)

			// Store the original os.Stdout and redirect it to a pipe
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Verify the script is written to stdout
			err := d.Write(true)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, d2format.Format(d.d2Graph.AST))

			// Restore the original stdout and read the captured output from the pipe
			w.Close()
			os.Stdout = originalStdout
			capturedOutput, _ := io.ReadAll(r)
			assert.Contains(t, string(capturedOutput), tc.expected)
		})
	}
}

func TestGetIconURL(t *testing.T) {
	testCases := []struct {
		name     string
		resource string
		expected string
	}{
		{
			name:     "aws icon",
			resource: "aws_ecs_cluster",
			expected: "https://raw.githubusercontent.com/tf2d2/icons/main/aws/service/Containers/64/Amazon-Elastic-Container-Service.svg",
		},
		{
			name:     "azure icon",
			resource: "azurerm_aadb2c_directory",
			expected: "https://raw.githubusercontent.com/tf2d2/icons/main/azure/identity/Azure-AD-B2C.svg",
		},
		{
			name:     "google icon",
			resource: "google_compute_instance",
			expected: "",
		},
		{
			name:     "unknown cloud provider",
			resource: "unknown_cloud_provider",
			expected: "",
		},
		{
			name:     "invalid aws resource",
			resource: "aws_invalid_resource",
			expected: "",
		},
		{
			name:     "invalid azure resource",
			resource: "azurerm_invalid_resource",
			expected: "",
		},
		{
			name:     "invalid google resource",
			resource: "google_invalid_resource",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := getIconURL(tc.resource)
			assert.Equal(t, tc.expected, url.String())
		})
	}
}

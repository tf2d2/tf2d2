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

package diagram

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tf2d2/tf2d2/internal/inframap"
	d2tpl "github.com/tf2d2/tf2d2/internal/template"

	"github.com/hashicorp/go-hclog"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

// IDiagram defines how to generate a d2 diagram
type IDiagram interface {
	Initialize() error
	Generate() error
	Render() error
}

// Diagram implements IDiagram to generate a d2 diagram
type Diagram struct {
	ctx           context.Context // parent context
	Filepath      string
	TFInfraMap    *inframap.TFInfraMap
	d2Graph       *d2graph.Graph
	d2CompileOpts *d2lib.CompileOptions
	d2RenderOpts  *d2svg.RenderOpts
}

// NewDiagram creates a new Diagram instance
func NewDiagram(ctx context.Context, m *inframap.TFInfraMap, filepath string) *Diagram {
	return &Diagram{
		ctx:           ctx,
		Filepath:      filepath,
		TFInfraMap:    m,
		d2Graph:       nil,
		d2CompileOpts: nil,
		d2RenderOpts:  nil,
	}
}

// Initialize creates a new empty D2 graph
func (d *Diagram) Initialize() error {
	logger := hclog.FromContext(d.ctx)

	// initialize a ruler to measure font glyphs
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return err
	}

	// initialize layout resolver
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}

	// initialize compile options
	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}

	// initialize render options
	renderOpts := &d2svg.RenderOpts{
		Pad: go2.Pointer(int64(d2svg.DEFAULT_PADDING)),
	}

	// initialize empty diagram
	_, graph, err := d2lib.Compile(d.ctx, "", nil, nil)
	if err != nil {
		return fmt.Errorf("error initializing diagram: %w", err)
	}

	d.d2Graph = graph
	d.d2CompileOpts = compileOpts
	d.d2RenderOpts = renderOpts

	logger.Info("initialized empty d2 diagram")

	return nil
}

// Generate generates a D2 diagram
func (d *Diagram) Generate() error {
	logger := hclog.FromContext(d.ctx)

	// compute d2 shapes
	shapes := []*d2target.Shape{}
	for _, n := range d.TFInfraMap.Graph.Nodes {
		logger.Debug(fmt.Sprintf("%#v\n", n))

		s := d2target.BaseShape()
		s.ID = strings.ReplaceAll(n.Canonical, ".", "_")
		s.Label = n.Resource.Name
		iconLink, err := url.Parse("")
		if err != nil {
			return err
		}
		s.Icon = iconLink
		shapes = append(shapes, s)
	}

	// compute d2 connections between shapes
	conns := []*d2target.Connection{}
	for _, e := range d.TFInfraMap.Graph.Edges {
		sourceN, err := d.TFInfraMap.Graph.GetNodeByID(e.Source)
		if err != nil {
			return fmt.Errorf("error getting source node: %w", err)
		}
		targetN, err := d.TFInfraMap.Graph.GetNodeByID(e.Target)
		if err != nil {
			return fmt.Errorf("error getting target node: %w", err)
		}

		c := d2target.BaseConnection()
		c.Src = strings.ReplaceAll(sourceN.Canonical, ".", "_")
		c.Dst = strings.ReplaceAll(targetN.Canonical, ".", "_")
		c.DstArrow = d2target.ArrowArrowhead
		conns = append(conns, c)
	}

	// render d2 template with computed shapes and connections
	t := d2tpl.New(shapes, conns)
	out, err := t.Render("d2")
	if err != nil {
		return err
	}

	// compile d2 diagram from template render output
	_, d.d2Graph, err = d2lib.Compile(d.ctx, out, d.d2CompileOpts, d.d2RenderOpts)
	if err != nil {
		return err
	}

	logger.Info("generated d2 diagram")

	return nil
}

// Render renders both a D2 diagram and script if not `dryRun`.
// Otherwise, D2 script is written to stdout
func (d *Diagram) Render(dryRun bool) error {
	logger := hclog.FromContext(d.ctx)

	// turn the graph into a script
	script := d2format.Format(d.d2Graph.AST)

	// compile the script into a diagram
	diagram, _, err := d2lib.Compile(d.ctx, script, d.d2CompileOpts, d.d2RenderOpts)
	if err != nil {
		return err
	}

	if dryRun {
		_, err = os.Stdout.WriteString(script + "\n")
		if err != nil {
			return err
		}
	} else {
		// build filenames
		scriptFilename := fmt.Sprintf("%s.d2", d.Filepath)
		diagramFilename := fmt.Sprintf("%s.svg", d.Filepath)
		// render to svg
		out, _ := d2svg.Render(diagram, d.d2RenderOpts)

		// write d2 script to output file path
		err = os.WriteFile(filepath.Clean(scriptFilename), []byte(script), 0600)
		if err != nil {
			return err
		}

		// write d2 diagram to output file path
		err = os.WriteFile(filepath.Clean(diagramFilename), out, 0600)
		if err != nil {
			return err
		}

		logger.Info("rendered d2 diagram")
		logger.Info("output files", "diagram", diagramFilename, "script", scriptFilename)
	}

	return nil
}

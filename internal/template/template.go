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

package template

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/url"
	gtpl "text/template"

	"oss.terrastruct.com/d2/d2target"

	"github.com/Masterminds/sprig"
)

//go:embed d2.tpl
var tpl string

// Template represents the template data used to render a d2 script
type Template struct {
	Shapes      []*d2target.Shape
	Connections []*d2target.Connection
	funcMap     gtpl.FuncMap
}

// New returns a new Template instance
func New(shapes []*d2target.Shape, conns []*d2target.Connection) *Template {
	return &Template{
		Shapes:      shapes,
		Connections: conns,
		funcMap:     defaultFuncs(),
	}
}

func defaultFuncs() gtpl.FuncMap {
	fns := gtpl.FuncMap{
		"urlToString": func(u url.URL) string {
			return u.String()
		},
	}

	// add other useful template functions
	for name, fn := range sprig.FuncMap() {
		if _, found := fns[name]; !found {
			fns[name] = fn
		}
	}

	return fns
}

// Render template
func (t *Template) Render(name string) (string, error) {
	if len(t.Shapes) < 1 {
		return "", fmt.Errorf("no shapes found")
	}

	var buffer bytes.Buffer

	tmpl := gtpl.New(name)
	tmpl.Funcs(t.funcMap)
	gtpl.Must(tmpl.Parse(tpl))

	if err := tmpl.ExecuteTemplate(&buffer, name, t); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

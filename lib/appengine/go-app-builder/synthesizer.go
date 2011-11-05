// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"template"
)

// MakeMain creates the synthetic main package for a Go App Engine app.
func MakeMain(app *App, extraImports []string) (string, error) {
	buf := new(bytes.Buffer)
	data := &templateData{
		App:          app,
		ExtraImports: extraImports,
	}
	if err := mainTemplate.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type templateData struct {
	App          *App
	ExtraImports []string
}

var mainTemplate = template.Must(template.New("main").Parse(
	`package main

import (
	"appengine_internal"
	{{range .ExtraImports}}
	_ "{{.}}"
	{{end}}

	// Top-level app packages
	{{range .App.RootPackages}}
	_ "{{.ImportPath}}"
	{{end}}
)

func main() {
	appengine_internal.Main()
}
`))

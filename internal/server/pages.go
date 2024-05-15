package server

import (
	"bytes"
	"errors"
	"html/template"
	"log"
)

const getAllHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Metrics list</title>
  </head>
  <body>
	<ul>{{ range .Data}}
	<li>{{ .Name }} {{ .Value }}</li>{{ end }}
	</ul>
  </body>
</html>`

type Item struct {
	Name  string
	Value string
}

type templateArgs struct {
	Data []Item
}

func renderGetAll(data []Item) (*bytes.Buffer, error) {
	indexTemplate := template.Must(template.New("metrics").Parse(getAllHTML))
	buf := new(bytes.Buffer)
	err := indexTemplate.Execute(buf, templateArgs{Data: data})
	if err != nil {
		log.Println(err)
		return nil, errors.Unwrap(err)
	}

	return buf, nil
}
package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/client/*html
var tmplFS embed.FS

//go:embed templates/static/*
var content embed.FS

// holds static folder data
func Static() embed.FS {
	return content
}

var fm template.FuncMap

type Template struct {
	templates *template.Template
}

func New() *Template {
	templates := template.Must(template.New("").Funcs(fm).ParseFS(tmplFS, "templates/client/*html"))
	return &Template{
		templates: templates,
	}
}
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(tmplFS, "templates/client/"+name))
	return tmpl.ExecuteTemplate(w, name, data)
}

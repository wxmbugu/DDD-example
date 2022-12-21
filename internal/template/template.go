package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/client/*.html
var tmplFS embed.FS

type Template struct {
	templates *template.Template
}

func New() *Template {
	funcMap := template.FuncMap{
		"inc": "inc",
	}

	templates := template.Must(template.New("").Funcs(funcMap).ParseFS(tmplFS, "./templates/client/*.html"))
	return &Template{
		templates: templates,
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(tmplFS, "./templates/client/"+name))
	return tmpl.ExecuteTemplate(w, name, data)
}

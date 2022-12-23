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
	templates := template.Must(template.New("").ParseFS(tmplFS, "./templates/client/*.html"))
	return &Template{
		templates: templates,
	}
}

// Adds custom functions to templates
func (t *Template) CustomFunc(name string, f func(any) any) *Template {
	t.templates.Funcs(template.FuncMap{
		name: f,
	})
	return t
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(tmplFS, "./templates/client/"+name))
	return tmpl.ExecuteTemplate(w, name, data)
}

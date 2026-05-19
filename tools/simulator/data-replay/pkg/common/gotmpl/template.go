package gotmpl

import "text/template"

func New(name string) *template.Template {
	tmpl := template.New(name)
	tmpl.Funcs(fm)

	return tmpl
}

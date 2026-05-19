package gotmpl

import (
	"text/template"
)

var fm template.FuncMap

func init() {
	fm = template.FuncMap{
		"now":      now,
		"location": location,
	}
}

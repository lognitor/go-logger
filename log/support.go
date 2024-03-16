package log

import "github.com/valyala/fasttemplate"

func newTemplate(t string) *fasttemplate.Template {
	return fasttemplate.New(t, "${", "}")
}

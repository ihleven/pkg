package tmplt

import (
	"html/template"
	"net/http"
)

type Tmpl struct {
	// opts          Options
	templates     map[string]*template.Template
	globTemplates *template.Template
	headers       map[string]string
}

func New() *Tmpl {
	return &Tmpl{}
}

func (t *Tmpl) Render(data interface{}) {

}

func handler(w http.ResponseWriter, r *http.Request) {

}

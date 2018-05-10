package tmpl

import (
	"html/template"
	"io/ioutil"
	"log"

	"github.com/jonomacd/forcedhappyness/site/statik"
)

const (
	templateDir = "/templates/"
	ext         = ".tmpl"
)

var (
	Templates = map[string]*TemplateInfo{
		"home": &TemplateInfo{
			Name:     "home",
			filepath: templateDir + "feed" + ext,
		},
		"login": &TemplateInfo{
			Name:     "login",
			filepath: templateDir + "login" + ext,
		},
		"post": &TemplateInfo{
			Name:     "post",
			filepath: templateDir + "post" + ext,
		},
		"register": &TemplateInfo{
			Name:     "register",
			filepath: templateDir + "register" + ext,
		},
	}
)

type TemplateInfo struct {
	Name     string
	filepath string
	Template *template.Template
}

func MustInit() {
	baseTmpl := template.Must(template.New("base").Parse(loadTemplateString(templateDir + "base" + ext)))

	for _, ti := range Templates {
		ti.LoadTemplate(template.Must(baseTmpl.Clone()))
	}
}

func GetTemplate(name string) *template.Template {
	return Templates[name].Template
}

func (ti *TemplateInfo) LoadTemplate(base *template.Template) {
	ti.Template = template.Must(base.Parse(loadTemplateString(ti.filepath)))
}

func loadTemplateString(filepath string) string {
	log.Printf("Loading template: %s", filepath)
	f, err := statik.StatikFS.Open(filepath)
	if err != nil {
		panic(err)
	}
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return string(bb)
}

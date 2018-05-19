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
			Name: "home",
			filepaths: []string{
				templateDir + "submitform" + ext,
				templateDir + "feed" + ext,
				templateDir + "post" + ext,
			},
		},
		"comments": &TemplateInfo{
			Name: "comments",
			filepaths: []string{
				templateDir + "comments" + ext,
				templateDir + "post" + ext,
			},
		},
		"user": &TemplateInfo{
			Name: "user",
			filepaths: []string{
				templateDir + "user" + ext,
				templateDir + "post" + ext,
			},
		},
		"login": &TemplateInfo{
			Name:      "login",
			filepaths: []string{templateDir + "login" + ext},
		},
		"submit": &TemplateInfo{
			Name: "submit",
			filepaths: []string{
				templateDir + "submitform" + ext,
				templateDir + "submit" + ext,
			},
		},
		"register": &TemplateInfo{
			Name:      "register",
			filepaths: []string{templateDir + "register" + ext},
		},
	}
)

type TemplateInfo struct {
	Name      string
	filepaths []string
	Template  *template.Template
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
	if len(ti.filepaths) == 0 {
		return
	}

	log.Printf("Running template group %s", ti.Name)
	ti.Template = base

	for _, filepath := range ti.filepaths {
		ti.Template = template.Must(ti.Template.Parse(loadTemplateString(filepath)))
	}
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

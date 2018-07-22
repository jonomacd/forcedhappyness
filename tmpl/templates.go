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
				templateDir + "happyness" + ext,
			},
		},
		"comments": &TemplateInfo{
			Name: "comments",
			filepaths: []string{
				templateDir + "comments" + ext,
				templateDir + "post" + ext,
				templateDir + "submitform" + ext,
				templateDir + "happyness" + ext,
			},
		},
		"user": &TemplateInfo{
			Name: "user",
			filepaths: []string{
				templateDir + "user" + ext,
				templateDir + "post" + ext,
				templateDir + "submitform" + ext,
				templateDir + "happyness" + ext,
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
		"settings": &TemplateInfo{
			Name:      "settings",
			filepaths: []string{templateDir + "settings" + ext},
		},
		"notifications": &TemplateInfo{
			Name:      "notifications",
			filepaths: []string{templateDir + "notifications" + ext},
		},
		"postonly": &TemplateInfo{
			Name:   "postonly",
			noBase: true,
			filepaths: []string{
				templateDir + "post" + ext,
				templateDir + "submitform" + ext,
				templateDir + "happyness" + ext,
			},
		},
		"error": &TemplateInfo{
			Name: "error",
			filepaths: []string{
				templateDir + "error" + ext,
			},
		},
		"angryban": &TemplateInfo{
			Name: "angryban",
			filepaths: []string{
				templateDir + "angryban" + ext,
			},
		},
		"welcome": &TemplateInfo{
			Name: "welcome",
			filepaths: []string{
				templateDir + "welcome" + ext,
			},
		},
		"rottenposts": &TemplateInfo{
			Name: "rottenposts",
			filepaths: []string{
				templateDir + "rottenposts" + ext,
			},
		},
		"moderate": &TemplateInfo{
			Name: "moderate",
			filepaths: []string{
				templateDir + "post" + ext,
				templateDir + "moderation" + ext,
			},
		},
	}
)

type TemplateInfo struct {
	Name      string
	filepaths []string
	Template  *template.Template
	noBase    bool
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
	if !ti.noBase {
		ti.Template = base
	} else {
		ti.Template = template.New(ti.Name)
	}

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

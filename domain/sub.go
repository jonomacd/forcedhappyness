package domain

import "html/template"

type Sub struct {
	Name        string
	DisplayName string
	Description template.HTML `datastore:",noindex"`
	Owners      []string
	Moderators  []string
	Claimed     bool `datastore:",noindex"`
}

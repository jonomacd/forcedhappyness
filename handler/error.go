package handler

import (
	"log"
	"net/http"

	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

func renderError(w http.ResponseWriter, message string, hasSession bool) {

	be := &domain.BasePage{
		HasSession: hasSession,
		Error:      message,
	}
	err := tmpl.GetTemplate("error").Execute(w, be)
	if err != nil {
		log.Printf("Error page failed. Terrible: %v", err)
	}
}

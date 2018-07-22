package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jonomacd/forcedhappyness/site/images"
)

type UploadHandler struct {
	ss sessions.Store
}

func NewUploadHandler(ss sessions.Store) *UploadHandler {
	return &UploadHandler{
		ss: ss,
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ff, ffh, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		w.WriteHeader(500)
		return
	}
	contentType := ffh.Header.Get("Content-Type")
	switch contentType {
	case "image/bmp", "image/gif", "image/jpeg", "image/tiff", "image/png":
	default:
		log.Printf("bad content type: %s", contentType)
		w.WriteHeader(400)
		return
	}

	ext := ""
	fnext := strings.Split(ffh.Filename, ".")
	if len(fnext) > 0 {
		ext = fnext[len(fnext)-1]
	}
	url, err := images.UploadImage(context.Background(), images.BucketImages, contentType, ext, ff)
	if err != nil {
		log.Printf("Error uploading to GCS: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Write([]byte(`{"url":"` + url + `"}`))
}

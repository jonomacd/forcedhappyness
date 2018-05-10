package statik

import (
	"log"
	"net/http"

	"github.com/rakyll/statik/fs"
)

var StatikFS http.FileSystem

func Init() {
	var err error
	StatikFS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/jonomacd/forcedhappyness/site/certificate"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/embedder"
	"github.com/jonomacd/forcedhappyness/site/handler"
	"github.com/jonomacd/forcedhappyness/site/images"
	"github.com/jonomacd/forcedhappyness/site/push"
	"github.com/jonomacd/forcedhappyness/site/sentiment"
	"github.com/jonomacd/forcedhappyness/site/statik"
	"github.com/jonomacd/forcedhappyness/site/tmpl"
)

//go:generate statik -src=static/
func main() {

	init := flag.Bool("init", false, "Use on first run on a new server")
	port := flag.String("port", "80", "Port number to use")
	domain := flag.String("domain", ".iwillbenice.com", "Domain to serve on")
	flag.Parse()
	if *init {
		if err := certificate.Run("jonomacd@gmail.com"); err != nil {
			log.Printf("error getting cert: %v", err)
		}
	}

	statik.Init()
	tmpl.MustInit()
	if err := images.Init(); err != nil {
		log.Printf("image buckets failed to init: %v", err)
	}

	f, err := statik.StatikFS.Open("/indexes/index.yaml")
	if err != nil {
		panic(err)
	}
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := dao.Init(bb); err != nil {
		log.Printf("error init dao: %v", err)
	}

	provs, err := statik.StatikFS.Open("/embeds/providers.json")
	if err != nil {
		panic(err)
	}
	if err := embedder.Init(provs); err != nil {
		log.Printf("error init embedder: %v", err)
	}

	sentiment.InitNLP()
	publicPushKey := push.Init()
	registerHandlers(*domain, publicPushKey)
	addr := "0.0.0.0:" + *port
	log.Printf("Serving on %s", addr)
	err = http.ListenAndServe(addr, nil)
	log.Printf("Failed to serve: %v", err)
}

func registerHandlers(domain, publicPushKey string) {
	log.Printf("Registering handlers")
	sessionStore := dao.NewSessionStore(domain)

	if domain == "" {
		// Use phony key
		publicPushKey = "BFxNWyMfHFEcaZSb9kRkCoCCpQG5I_wNNGwR1CucRySfnH7Qv-8s6bHOHJpUIkA6K5HBZ00zCe7lenDy33ADUr8"
	}

	http.Handle("/", handler.NewHomeFeedHandler(sessionStore))
	http.Handle("/n/", handler.NewSubHandler(sessionStore))
	http.Handle("/sub", handler.NewSubCRUDHandler(sessionStore))
	http.Handle("/post/", handler.NewPostHandler(sessionStore))
	http.Handle("/settings", handler.NewSettingsHandler(sessionStore))
	http.Handle("/user/", handler.NewUserHandler(sessionStore))
	http.Handle("/submit", handler.NewSubmitHandler(sessionStore))
	http.Handle("/reply/", handler.NewReplyHandler(sessionStore))
	http.Handle("/login", handler.NewLoginHandler(sessionStore))
	http.Handle("/logout", handler.NewLogoutHandler(sessionStore))
	http.Handle("/register", handler.NewRegisterHandler(sessionStore))
	http.Handle("/register/google", handler.NewGoogleRegisterHandler(sessionStore))
	http.Handle("/like", handler.NewLikeHandler(sessionStore))
	http.Handle("/follow", handler.NewFollowHandler(sessionStore))
	http.Handle("/welcome", handler.NewWelcomeHandler(sessionStore))
	http.Handle("/search", handler.NewSearchHandler(sessionStore))
	http.Handle("/shame/", handler.NewRottenPostsHandler(sessionStore))
	http.Handle("/notifications", handler.NewNotificationHandler(sessionStore, publicPushKey))
	http.Handle("/upload", handler.NewUploadHandler(sessionStore))
	http.Handle("/moderation/", handler.NewDeletePostHandler(sessionStore))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(statik.StatikFS)))

	// Service worker MUST be at root of domain
	serveRoot("/js/service-worker.js", "application/javascript")
	for _, favicon := range faviconRoot {
		serveRoot("/img/favicon/"+favicon, "")
	}

}

var (
	faviconRoot = []string{"android-icon-144x144.png", "apple-icon-114x114.png", "apple-icon-60x60.png", "favicon-16x16.png", "ms-icon-150x150.png",
		"android-icon-192x192.png", "apple-icon-120x120.png", "apple-icon-72x72.png", "favicon-32x32.png", "ms-icon-310x310.png",
		"android-icon-36x36.png", "apple-icon-144x144.png", "apple-icon-76x76.png", "favicon-96x96.png", "ms-icon-70x70.png",
		"android-icon-48x48.png", "apple-icon-152x152.png", "apple-icon-precomposed.png", "favicon.ico",
		"android-icon-72x72.png", "apple-icon-180x180.png", "apple-icon.png", "manifest.json",
		"android-icon-96x96.png", "apple-icon-57x57.png", "browserconfig.xml", "ms-icon-144x144.png"}
)

func serveRoot(file, contentType string) {
	// Service worker MUST be at root of domain

	f, err := statik.StatikFS.Open(file)
	if err != nil {
		panic(err)
	}
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	fileA := strings.Split(file, "/")

	http.HandleFunc("/"+fileA[len(fileA)-1], func(w http.ResponseWriter, r *http.Request) {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Write(bb)
	})
}

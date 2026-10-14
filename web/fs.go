package web

import (
	"embed"
	"log"
	"net/http"
	"os"
	"strings"
)

func listContent(fs embed.FS) {
	entries, err := fs.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		log.Print(entry.Name())
	}
}

type filehander struct{}

func (fs *filehander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	w.Header().Add("Cache-Control", "no-cache")
	buf, err := os.ReadFile("web/resources/" + path)
	if err != nil {
		log.Fatalf("failed to read: %v\n", err)
	}
	if strings.HasPrefix(path, "css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasPrefix(path, "js") {
		w.Header().Set("Content-Type", "application/javascript")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

var fs = &filehander{}

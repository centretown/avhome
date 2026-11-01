package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	var (
		// sourceDir  = "/mnt/molly/recordings"
		symLink   = "recordings"
		sourceDir = flag.String("source", "/mnt/molly/recordings",
			"the path of the source directory")
	)
	flag.Parse()
	log.Print(*sourceDir)

	err := os.Symlink(*sourceDir, symLink)
	if err != nil {
		log.Fatal(err)
	}

	var fs = &filehander{
		symLink: symLink,
		handler: http.FileServer(http.Dir(symLink)),
	}

	mux := &http.ServeMux{}
	server := &http.Server{
		Addr:    ":9900",
		Handler: mux,
	}

	mux.Handle("/", fs)
	go serve(server)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	os.Remove(symLink)
}

func serve(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type filehander struct {
	handler http.Handler
	symLink string
}

func (fs *filehander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache")
	fs.handler.ServeHTTP(w, r)
}

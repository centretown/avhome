package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
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

type Runtime struct {
	MtxUrl string
	Paths  *MtxPathList
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

func Run() {
	var (
		err     error
		runtime = &Runtime{
			MtxUrl: "10.0.0.7",
		}
		mux    = http.NewServeMux()
		server = &http.Server{
			Addr:    ":9000",
			Handler: mux,
		}
	)
	defer server.Close()

	runtime.Paths, err = GetMtxPaths(runtime.MtxUrl, "dave", "football")
	if err != nil {
		log.Printf("failed to contact Mtx: %v\n", err)
		return
	}

	// fs := http.FileServer(http.Dir("web/resources/"))
	var httpTemplate *template.Template
	refreshTemplate := func() {
		t, err := template.ParseGlob("web/resources/html/*.html")
		if err != nil {
			log.Printf("failed to create html template: %v\n", err)
		}
		httpTemplate = t
	}

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		refreshTemplate()
		mainTmpl := httpTemplate.Lookup("main")
		w.WriteHeader(http.StatusOK)
		mainTmpl.Execute(w, runtime)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		buf, _ := os.ReadFile("web/resources/favicon.ico")
		w.Write(buf)
	})

	go serve(server)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	for {
		select {
		case sig := <-sigs:
			log.Printf("Signal: %v", sig)
			return
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}

}

func serve(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

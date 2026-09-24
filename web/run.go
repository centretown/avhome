package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed resources/*
var content embed.FS

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

func Run() {
	listContent(content)
	// return
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))
	t, err := template.ParseFS(content, "resources/html/*.html")
	if err != nil {
		log.Printf("failed to create html template: %v\n", err)
		return
	}

	mainTmpl := t.Lookup("main")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("root %s\n", r.RequestURI)
		w.WriteHeader(http.StatusOK)
		mainTmpl.Execute(w, runtime)
	})
	mux.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("css %s\n", r.RequestURI)
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		buf, _ := content.ReadFile("resources" + r.RequestURI)
		w.Write(buf)
	})
	mux.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("js %s\n", r.RequestURI)
		w.Header().Set("Cache-Control", "max-age=3600")
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		buf, _ := content.ReadFile("resources" + r.RequestURI)
		w.Write(buf)
	})
	mux.HandleFunc("/home.js", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("home %s\n", r.RequestURI)
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		buf, _ := content.ReadFile("resources/home.js")
		w.Write(buf)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		buf, _ := content.ReadFile("resources/favicon.ico")
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

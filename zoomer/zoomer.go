package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

func main() {
	var (
		sourceDir = flag.String("source", "web",
			"the path of the source directory")
	)
	flag.Parse()
	log.Print(*sourceDir)

	mux := &http.ServeMux{}
	server := &http.Server{
		Addr:    ":9800",
		Handler: mux,
	}

	mtxUrl := ""
	fs := http.FileServer(http.Dir(""))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, "192.168.10") {
			mtxUrl = "192.168.10.7"
		} else if strings.HasPrefix(r.Host, "10.0.0") {
			mtxUrl = "10.0.0.7"
		}
		log.Print("root", r.Host, "mtxUrl", mtxUrl)
		fs.ServeHTTP(w, r)
		// // debugging purposes only
		// webRun.refreshTemplate()
		// mainTmpl := webRun.template.Lookup("main")
		// w.Header().Add("Cache-Control", "no-cache")
		// w.WriteHeader(http.StatusOK)
		// err := mainTmpl.Execute(w, webRun)
		// if err != nil {
		// 	log.Printf("failed to create html template: %v\n", err)
		// }
	})

	go serve(server)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	log.Print("done")

}

func serve(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

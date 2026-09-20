package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed reader.js
var script []byte

//go:embed main.html
var page []byte

//go:embed favicon.ico
var icon []byte

var (
	url = "192.168.10.188:9000"
)

func Run() {
	GetPaths()

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    url,
		Handler: mux,
	}
	defer server.Close()
	// mux.Handle("/reader.js", http.FileServer(http.FS(script)))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s\n", r.RequestURI)
		w.Write(page)
	})
	mux.HandleFunc("/reader.js", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s\n", r.RequestURI)
		w.Header().Set("Cache-Control", "max-age=3600")
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write(script)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(icon)
	})

	go serve(server)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Printf("Signal: %v", sig)
}

func serve(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func GetPaths() (list *APIPathList, err error) {
	list = &APIPathList{}
	query := "http://192.168.10.188:9997/v3/paths/list"
	err = queryAndDecode(query, list)
	if err != nil {
		log.Printf("query current attempt: %v\n\terror: %v", 1, err)
		return
	}
	log.Print(list.ItemCount)
	for _, item := range list.Items {
		log.Printf("name: %s, available: %v, source: %v, type: %v",
			item.Name, item.Available, item.Source.ID, item.Source.Type)
	}

	return
}

func queryAndDecode(query string, w any) (err error) {
	var resp *http.Response

	for attempt := range 3 {
		resp, err = http.Get(query)
		if err != nil {
			log.Printf("query current attempt: %v\n\terror: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}

		err = readAndDecode(resp.Body, w)
		if err != nil {
			log.Printf("decode current attempt: %v, error: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}

		return
	}

	return
}

func readAndDecode(r io.ReadCloser, w any) error {
	defer r.Close()
	buf, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %v", err)
	}
	err = json.Unmarshal(buf, w)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %v\n%v", err, string(buf))
	}
	return nil
}

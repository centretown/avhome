package web

import (
	"avhome/socket"
	"avhome/weather"
	"bytes"
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

type WebRun struct {
	MtxUrl  string
	Paths   *MtxPathList
	Weather *weather.Runtime
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
		err    error
		webRun = &WebRun{
			MtxUrl: "10.0.0.7",
		}
		mux    = http.NewServeMux()
		server = &http.Server{
			Addr:    ":9000",
			Handler: mux,
		}
		sockServer *socket.Server
	)

	webRun.Paths, err = GetMtxPaths(webRun.MtxUrl, "dave", "football")
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
	refreshTemplate()

	sockServer = socket.NewServer()
	sockServer.Run()
	defer sockServer.Done()
	mux.HandleFunc("/events", sockServer.Events)
	// mux.HandleFunc("/msghook", sockServer.MessageHook)

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		refreshTemplate()
		mainTmpl := httpTemplate.Lookup("main")
		w.WriteHeader(http.StatusOK)
		mainTmpl.Execute(w, webRun)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		buf, _ := os.ReadFile("web/resources/favicon.ico")
		w.Write(buf)
	})

	go serve(server)

	webRun.Weather = weather.NewRuntime(mux, httpTemplate, sockServer)
	err = webRun.Weather.Connect()
	if err != nil {
		log.Printf("failed to connect weather service: %v\n", err)
		return
	}
	defer webRun.Weather.Done()
	webRun.Weather.QueryCurrent()
	webRun.Weather.QueryDaily()
	webRun.Weather.QueryHourly()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	resetTicker := true
	span := 15
	ticker := time.NewTicker(firstTicker(span))
	rt := webRun.Weather
	buf := bytes.Buffer{}
	now := time.Now()
	for {
		select {
		case sig := <-sigs:
			log.Printf("Signal: %v", sig)
			return
		case now = <-ticker.C:
			if resetTicker {
				ticker.Reset(time.Duration(span) * time.Minute)
				resetTicker = false
			}
			if now.Minute() == 0 {
				rt.QueryHourly()
				if now.Hour()%4 == 0 {
					rt.QueryDaily()
				}
			}

			rt.QueryCurrent()
			t := httpTemplate.Lookup("current.summary")
			t.Execute(&buf, rt)
			sockServer.Broadcast(buf.String())
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

func firstTicker(span int) (ticker time.Duration) {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(),
		now.Minute(), 0, 0, now.Location())
	minute := span - next.Minute()%span
	if minute == 0 {
		minute = span
	}
	next = next.Add(time.Duration(minute) * time.Minute)
	if next.Compare(now) < 0 {
		log.Fatal(minute, now, next)
	}
	ticker = next.Sub(now)
	// log.Printf("ticker=%v, minute=%v now=%v next=%v\n", ticker, minute, now, next)
	return
}

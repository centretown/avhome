package web

import (
	"avhome/socket"
	"avhome/weather"
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	webRun = &Runtime{
		MtxUrl: "10.0.0.7",
	}
	sockServer *socket.Server
	mux        = http.NewServeMux()
	server     = &http.Server{
		Addr:    ":9000",
		Handler: mux,
	}
)

func Run() {
	var err error

	webRun.Paths, err = GetMtxPaths(webRun.MtxUrl, "dave", "football")
	if err != nil {
		log.Printf("failed to contact Mtx: %v\n", err)
		return
	}

	webRun.Weather = weather.NewRuntime()
	err = webRun.Weather.Connect("database/location.db")
	if err != nil {
		log.Printf("failed to connect weather database: %v\n", err)
		return
	}
	defer webRun.Weather.Done()
	err = webRun.Weather.LoadLocations()
	if err != nil {
		log.Printf("failed to load weather locations: %v\n", err)
		return
	}

	webRun.addCameraSections()
	for _, a := range webRun.Sections {
		log.Print(a)
	}

	sockServer = socket.NewServer()
	sockServer.Run()
	defer sockServer.Done()

	webRun.addHandlers(mux, sockServer)
	handlePlayers(mux, sockServer, webRun.MtxUrl, webRun.Paths, webRun.template)
	go serve(server)

	const (
		startMinute = 4  // 4 minutes after the hour
		spanMinute  = 15 // every 15 minutes
	)
	// webRun.r
	webRun.poll(startMinute, spanMinute)

}

func serve(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func firstTicker(startMinute int, spanMinute int) (ticker time.Duration) {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(),
		now.Minute(), 0, 0, now.Location())
	minute := spanMinute - next.Minute()%spanMinute + startMinute
	if minute == startMinute {
		minute = spanMinute
	}
	next = next.Add(time.Duration(minute) * time.Minute)
	if next.Compare(now) < 0 {
		log.Fatal(minute, now, next)
	}
	ticker = next.Sub(now)
	log.Printf("ticker=%v, minute=%v now=%v next=%v\n", ticker, minute, now, next)
	return
}

func (webRun *Runtime) refreshTemplate() {
	t, err := template.New("").Funcs(funcMap).ParseGlob("web/resources/html/*.html")
	if err != nil {
		log.Printf("failed to create html template: %v\n", err)
	}
	webRun.template = t
}

func (webRun *Runtime) addHandlers(mux *http.ServeMux, sockServer *socket.Server) {
	webRun.refreshTemplate()

	mux.HandleFunc("/events", sockServer.Events)
	// mux.HandleFunc("/msghook", sockServer.MessageHook)

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("root", r.Host)
		if strings.HasPrefix(r.Host, "192.168.10") {
			webRun.MtxUrl = "192.168.10.7"
		} else if strings.HasPrefix(r.Host, "10.0.0") {
			webRun.MtxUrl = "10.0.0.7"
		}
		// debugging purposes only
		webRun.refreshTemplate()
		mainTmpl := webRun.template.Lookup("main")
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		err := mainTmpl.Execute(w, webRun)
		if err != nil {
			log.Printf("failed to create html template: %v\n", err)
		}
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		buf, _ := os.ReadFile("web/resources/favicon.ico")
		w.Write(buf)
	})
}

func (webRun *Runtime) poll(startMinute int, spanMinute int) {

	webRun.Weather.QueryCurrent()
	webRun.Weather.QueryDaily()
	webRun.Weather.QueryHourly()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	resetTicker := true
	ticker := time.NewTicker(firstTicker(startMinute, spanMinute))
	buf := bytes.Buffer{}
	now := time.Now()

	var (
		currentSummaryT *template.Template
		currentT        *template.Template
		currentHistoryT *template.Template
		hourlyT         *template.Template
		hourlyDetailT   *template.Template
		dailyT          *template.Template
		dailyDetailT    *template.Template
	)

	refreshTemplates := func() {
		t := webRun.template
		currentSummaryT = t.Lookup("current.summary")
		currentT = t.Lookup("weather.current")
		currentHistoryT = t.Lookup("weather.current.history")
		hourlyT = t.Lookup("weather.hourly")
		hourlyDetailT = t.Lookup("weather.hourly.detail")
		dailyT = t.Lookup("weather.daily")
		dailyDetailT = t.Lookup("weather.daily.detail")
	}

	webRun.refreshTemplate()
	refreshTemplates()

	for {
		select {
		case sig := <-sigs:
			log.Printf("Signal: %v", sig)
			return
		case now = <-ticker.C:
			if resetTicker {
				ticker.Reset(time.Duration(spanMinute) * time.Minute)
				resetTicker = false
			}

			// go func() {
			rt := webRun.Weather
			rt.QueryCurrent()

			webRun.refreshTemplate()
			refreshTemplates()
			currentSummaryT.Execute(&buf, webRun)
			sockServer.Broadcast(buf.String())

			locations := rt.Locations
			broadcastLocations(locations, currentT, sockServer)
			broadcastLocations(locations, currentHistoryT, sockServer)

			if now.Minute() == startMinute {
				rt.QueryHourly()
				locations = rt.Locations
				broadcastLocations(locations, hourlyT, sockServer)
				broadcastLocations(locations, hourlyDetailT, sockServer)

				if now.Hour()%4 == 0 {
					rt.QueryDaily()
					locations = rt.Locations
					broadcastLocations(locations, dailyT, sockServer)
					broadcastLocations(locations, dailyDetailT, sockServer)
				}
			}
			// }()

		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

var funcMap = template.FuncMap{
	"sub": func(i int, j int) int {
		return i - j
	},
}

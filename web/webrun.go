package web

import (
	"avhome/socket"
	"avhome/weather"
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
	webRun.Weather = weather.NewRuntime(mux, httpTemplate)
	err = webRun.Weather.Run()
	if err != nil {
		log.Printf("failed to start weather service: %v\n", err)
		return
	}

	sockServer = socket.NewServer(httpTemplate)
	sockServer.Run()
	defer sockServer.Done()
	mux.HandleFunc("/events", sockServer.Events)
	mux.HandleFunc("/msghook", sockServer.MessageHook)

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

func Monitor(rt *weather.Runtime) {
	var (
		now         time.Time
		resetTicker = true
	)

	for {
		select {
		case now = <-rt.Ticker.C:
			if resetTicker {
				rt.Ticker.Reset(time.Minute * 15)
				resetTicker = false
			}
			rt.QueryCurrent()
			if now.Minute() == 0 {
				rt.QueryHourly()
				if now.Hour()%4 == 0 {
					rt.QueryDaily()
				}
			}
			rt.BroadcastTemperature()
		default:
			time.Sleep(time.Second)
		}
	}
}

//
// func (rt *Runtime) QueryDaily() {
// 	log.Println("Retrieving daily weather forecast...")
// 	for _, location := range rt.Locations {
// 		daily := &WeatherDaily{}
// 		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, dailyTrailer)
// 		err := queryAndDecode(query, daily)
// 		if err != nil {
// 			log.Printf("QueryDaily: queryAndDecode %v", err)
// 			continue
// 		}
// 		location.WeatherDaily = daily
// 		location.WeatherDaily.UpdateTime = time.Now()
// 		location.BuildDailyProperties()
// 	}
// }
//
// type LocationData struct {
// 	Index    int
// 	Location *Location
// }
//
// func (rt *Runtime) QueryHourly() {
// 	log.Println("Retrieving hourly weather forecast...")
// 	for _, location := range rt.Locations {
// 		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, hourlyTrailer)
// 		hourly := &WeatherHourly{}
// 		err := queryAndDecode(query, hourly)
// 		if err != nil {
// 			log.Printf("QueryHourly queryAndDecode: %v", err)
// 			continue
// 		}
// 		location.WeatherHourly = hourly
// 		location.WeatherHourly.UpdateTime = time.Now()
// 		location.BuildHourlyProperties()
// 	}
// }
//
// func (rt *Runtime) QueryCurrent() {
// 	log.Println("Retrieving current weather conditions...")
// 	for _, location := range rt.Locations {
// 		current := &WeatherCurrent{}
// 		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, currentTrailer)
//
// 		err := queryAndDecode(query, current)
// 		if err != nil {
// 			continue
// 		}
// 		if current.Current == nil {
// 			log.Println("current.Current is nil!")
// 			continue
// 		}
// 		// log.Println("rt.db ", rt.db, "current ", current)
// 		err = InsertHistory(rt.db, location.ID, current.Current)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
//
// 		location.WeatherCurrent = current
// 		location.WeatherCurrent.UpdateTime = time.Now()
// 	}
//
// 	err := rt.LoadHistory()
// 	if err != nil {
// 		log.Printf("Weather current load history error: %v", err)
// 	}
// }

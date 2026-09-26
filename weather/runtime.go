package weather

import (
	"avhome/action"
	"avhome/socket"
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type Runtime struct {
	Location      *Location
	Locations     []*Location
	LocationIndex int
	WebcamUrl     string
	WebcamIndex   int
	ActionsHome   []*action.Action
	ActionMap     map[string]*action.Action
	WebSocket     *socket.Server
	Template      *template.Template
	Ticker        *time.Ticker
	retry         *time.Ticker
	db            *sqlx.DB
	mux           *http.ServeMux
}

func NewRuntime(mux *http.ServeMux, tmpl *template.Template) (rt *Runtime) {
	rt = &Runtime{
		mux:      mux,
		Template: tmpl,
		ActionsHome: []*action.Action{
			{Name: "weather_current", Title: "Current Weather", Icon: "thunderstorm", Group: action.Home},
			{Name: "weather_hourly", Title: "24 Hour Forecast", Icon: "schedule", Group: action.Home},
			{Name: "weather_daily", Title: "7 Day Forecast", Icon: "calendar_view_week", Group: action.Home},
			// {Name: "lights", Title: "LED Lights", Icon: "backlight_high", Group: action.Home},
		},

		ActionMap: make(map[string]*action.Action),
	}
	for _, action := range rt.ActionsHome {
		rt.ActionMap[action.Name] = action
	}
	return
}

func (rt *Runtime) Run() (err error) {
	rt.Ticker = time.NewTicker(FirstTicker())
	err = rt.ConnectLocationData()
	if err != nil {
		log.Print(err)
		return
	}

	rt.Locations, err = SelectLocations(rt.db)
	if err != nil {
		log.Print(err)
		return
	}

	rt.Location = rt.Locations[0]
	for _, l := range rt.Locations {
		log.Printf("%v", l.City)
	}
	return
}

func (rt *Runtime) ConnectLocationData() (err error) {
	rt.db, err = OpenDB("database/location.db")
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func (rt *Runtime) SelectHistory(ID uint64, after string, before string) (history []*Current, err error) {
	return SelectHistoryInterval(rt.db, ID, after, before, "ASC")
}

func (rt *Runtime) LoadHistory() (err error) {
	after, before := BeforeTime(time.Now(), 6*time.Hour)
	for _, loc := range rt.Locations {
		history, err := rt.SelectHistory(loc.ID, after, before)
		if err != nil {
			log.Print(err)
		}
		loc.BuildCurrentProperties(history)
	}
	return
}

func (rt *Runtime) Done() {
	if rt.db != nil {
		rt.db.Close()
	}
}

type DailySummary struct {
	City            string
	High            string
	Low             string
	Precipitation   string
	Probability     string
	WindSpeed       string
	WindDirecection string
	WindGusts       string
	Code            string
	Color           string
}

func (rt *Runtime) CurrentWeatherDaily(index int) (hs DailySummary) {
	if index > len(rt.Locations) {
		return
	}

	loc := rt.Locations[index]
	daily := loc.WeatherDaily
	if len(daily.Daily.Time) < 1 {
		return
	}
	hs.City = loc.City
	hs.High = fmt.Sprintf("%4.1f %s",
		daily.Daily.High[0],
		daily.DailyUnits.High)
	hs.Low = fmt.Sprintf("%4.1f %s",
		daily.Daily.Low[0],
		daily.DailyUnits.Low)
	hs.Precipitation = fmt.Sprintf("%4.1f %s",
		daily.Daily.Precipitation[0],
		daily.DailyUnits.Precipitation)
	hs.Probability = fmt.Sprintf("%.0f%s",
		daily.Daily.Probability[0],
		daily.DailyUnits.Probability)
	code := WeatherCodes[daily.Daily.Code[0]]
	hs.Code = code.Icon
	hs.Color = code.Color
	return
}

type HourlySummary struct {
	City          string
	Temperature   string
	FeelsLike     string
	Precipitation string
	Probability   string
	WindSpeed     string
	WindDirection string
	WindGusts     string
	Humidity      string
	Pressure      string
	Code          string
	Color         string
}

func (rt *Runtime) CurrentTemperature() string {
	hourly := rt.Location.WeatherHourly
	if len(hourly.Hourly.Temperature) == 0 {
		return "99.9 ?"
	}
	return fmt.Sprintf("%2.1f%s",
		hourly.Hourly.Temperature[0],
		hourly.HourlyUnits.Temperature)
}

func (rt *Runtime) BroadcastTemperature() {
	buf := bytes.Buffer{}
	t := rt.Template.Lookup("weather.clock")
	t.Execute(&buf, rt)
	rt.WebSocket.Broadcast(buf.String())
}

type FormData struct {
	Action  *action.Action
	Data    any
	Codes   any
	Runtime *Runtime
}

type WeatherFormData struct {
	Action  *action.Action
	Data    any
	Codes   map[int32]*WeatherCode
	Runtime *Runtime
}

func (rt *Runtime) HandleAction(path string, templ string, data *WeatherFormData) {
	rt.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(path) < 2 {
			return
		}
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		data.Action = rt.ActionMap[path[1:]]

		err := rt.Template.Lookup(templ).Execute(w, data)
		if err != nil {
			log.Fatal(path, err)
		}
	})

}

func (rt *Runtime) HandleWeather() {
	data := &WeatherFormData{
		Codes:   WeatherCodes,
		Data:    rt.Locations,
		Runtime: rt}

	rt.HandleAction("/weather_daily", "weather.daily", data)
	rt.HandleAction("/weather_hourly", "weather.hourly", data)
	rt.HandleAction("/weather_current", "weather.current", data)

}

func wrapStatus(id, msg string) []byte {
	var buf []byte
	buf = fmt.Appendf(buf, `<div id="%s" class="status">%s</div>`, id, msg)
	return buf
}

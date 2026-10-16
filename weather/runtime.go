package weather

import (
	"fmt"
	"log"
	"strings"
	"time"

	"avhome/section"

	"github.com/jmoiron/sqlx"
)

type Runtime struct {
	Location      *Location
	Locations     []*Location
	LocationIndex int
	Sections      []*section.Section
	Section       *section.Section
	retry         *time.Ticker
	db            *sqlx.DB
}

const weather_section = "weather-section"

func NewRuntime() (rt *Runtime) {
	rt = &Runtime{
		Section: &section.Section{
			Name:   "weather",
			Title:  "Weather",
			Icon:   "thunderstorm",
			Parent: weather_section,
		},
		Sections: make([]*section.Section, 0),
	}
	return
}

func (rt *Runtime) Connect(fname string) (err error) {
	rt.db, err = OpenDB(fname)
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func (rt *Runtime) LoadLocations() (err error) {
	rt.Locations, err = SelectLocations(rt.db)
	if err != nil {
		log.Print(err)
		return
	}

	if len(rt.Locations) < 1 {
		err = fmt.Errorf("No locations found.")
		return
	}

	rt.Location = rt.Locations[0]
	for _, l := range rt.Locations {
		l.WeatherDaily = &WeatherDaily{}
		l.DailyProperties = &LocationProperties{}
		l.WeatherHourly = &WeatherHourly{}
		l.HourlyProperties = &LocationProperties{}
		l.WeatherCurrent = &WeatherCurrent{}
		l.CurrentProperties = &LocationProperties{}

		city := strings.ToLower(l.City)
		rt.Sections = append(rt.Sections, &section.Section{
			Name:   "weather-" + city,
			Title:  l.City + " Weather",
			Icon:   "thunderstorm",
			Parent: weather_section + "-" + city,
			Sections: []*section.Section{
				{
					Name:   "weather-current-detail" + city,
					Title:  "Recent Details",
					Icon:   "thunderstorm",
					Parent: weather_section,
				},
				{
					Name:   "weather-hourly-detail" + city,
					Title:  "Hourly Details",
					Icon:   "schedule",
					Parent: weather_section,
				},
				{
					Name:   "weather-daily-detail" + city,
					Title:  "Daily Details",
					Icon:   "calendar_view_week",
					Parent: weather_section + "-" + city,
				},
			},
		})

		log.Printf("%v", city)
	}

	return
}

func (rt *Runtime) SelectHistory(ID uint64, after string, before string) (current []Current, err error) {
	current = make([]Current, HISTORY_INTERVALS)
	history, err := SelectHistoryInterval(rt.db, ID, after, before, "ASC")
	if err != nil {
		return
	}
	tmAfter, err := time.Parse(TIME_LAYOUT, after)
	if err != nil {
		return
	}
	tmAfter = tmAfter.Truncate(time.Minute)
	historyLength := len(history)

	rec := 0
	nextTime := tmAfter
	var tm time.Time

	for i := range current {
		nextTime = nextTime.Add(time.Minute * HISTORY_MINUTE_INTERVAL)
		if rec < historyLength-1 {
			nextRec := rec + 1
			tm, err = time.Parse(TIME_LAYOUT, history[nextRec].Time)
			if err != nil {
				return
			}
			tm = tm.Truncate(time.Minute)
			if nextTime.Hour() >= tm.Hour() && nextTime.Minute() >= tm.Minute() {
				rec = nextRec
			}
		}

		current[i] = *history[rec]
		current[i].Time = nextTime.Format(TIME_LAYOUT)
	}
	// slices.Reverse(current)
	return
}

func (rt *Runtime) LoadHistory() (err error) {
	after, before := AfterBeforeString(time.Now().
		Truncate(time.Minute), HISTORY_DURATION, HISTORY_MINUTE_INTERVAL)
	for _, loc := range rt.Locations {
		history, err := rt.SelectHistory(loc.ID, after, before)
		if err != nil {
			log.Print(err)
			continue
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

type FormData struct {
	Action  *section.Section
	Data    any
	Codes   any
	Runtime *Runtime
}

type WeatherFormData struct {
	Action  *section.Section
	Data    any
	Codes   map[int32]*WeatherCode
	Runtime *Runtime
}

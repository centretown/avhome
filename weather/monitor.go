package weather

import (
	"avhome/query"
	_ "avhome/query"
	"fmt"
	"log"
	"time"
)

func (rt *Runtime) QueryDaily() {
	log.Println("Retrieving daily weather forecast...")
	for _, location := range rt.Locations {
		daily := &WeatherDaily{}
		q := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, dailyTrailer)
		err := query.QueryAndDecode(q, daily)
		if err != nil {
			log.Printf("QueryDaily: queryAndDecode %v", err)
			continue
		}
		if len(daily.Daily.Time) == 0 {
			log.Print("Query Daily: no data")
			continue
		}
		location.WeatherDaily = daily
		location.WeatherDaily.UpdateTime = time.Now()
		location.BuildDailyProperties()
	}
}

type LocationData struct {
	Index    int
	Location *Location
}

func (rt *Runtime) QueryHourly() {
	log.Println("Retrieving hourly weather forecast...")
	for _, location := range rt.Locations {
		q := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, hourlyTrailer)
		hourly := &WeatherHourly{}
		err := query.QueryAndDecode(q, hourly)
		if err != nil {
			log.Printf("QueryHourly queryAndDecode: %v", err)
			continue
		}

		if len(hourly.Hourly.Time) == 0 {
			log.Printf("Query Hourly: no data\n\t%v", hourly)
			continue
		}

		location.WeatherHourly = hourly
		_ = location.WeatherHourly.Hourly.Code[0]
		location.WeatherHourly.UpdateTime = time.Now()
		location.BuildHourlyProperties()
	}
}

func (rt *Runtime) QueryCurrent() {
	log.Println("Retrieving current weather conditions...")
	for _, location := range rt.Locations {
		current := &WeatherCurrent{}
		q := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, currentTrailer)
		err := query.QueryAndDecode(q, current)
		if err != nil {
			continue
		}
		if len(current.Current.Time) == 0 {
			log.Print("Query Current: no data")
			continue
		}

		location.WeatherCurrent = current
		location.WeatherCurrent.UpdateTime = time.Now()
		err = InsertHistory(rt.db, location.ID, &current.Current)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	err := rt.LoadHistory()
	if err != nil {
		log.Printf("Weather current load history error: %v", err)
	}
}

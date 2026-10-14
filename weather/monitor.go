package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func (rt *Runtime) QueryDaily() {
	log.Println("Retrieving daily weather forecast...")
	for _, location := range rt.Locations {
		daily := &WeatherDaily{}
		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, dailyTrailer)
		err := queryAndDecode(query, daily)
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
		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, hourlyTrailer)
		hourly := &WeatherHourly{}
		err := queryAndDecode(query, hourly)
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
		query := fmt.Sprintf(weatherFormat, weatherHeader, location.Latitude, location.Longitude, location.Zone, currentTrailer)
		err := queryAndDecode(query, current)
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

	if len(buf) < 100 {
		e := &ServiceError{}
		log.Print(string(buf))
		json.Unmarshal(buf, e)
		// {"reason":"The service is overloaded","error":true}
		return fmt.Errorf("readAndDecode: small buffer returned.\n\t"+
			"Reason: %s, Error:%v", e.Reason, e.Error)
	}

	err = json.Unmarshal(buf, w)
	if err != nil {
		log.Print(string(buf))
		return fmt.Errorf("json.Unmarshal: %v\n%v", err, string(buf))
	}
	return nil
}

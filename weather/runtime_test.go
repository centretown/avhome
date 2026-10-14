package weather

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	hourlyJsonName = "hourly_test.json"
	hourlyHtmlName = "hourly_test.html"
)

func TestWeather(t *testing.T) {
	rt := NewRuntime()
	err := rt.Connect("../database/location.db")
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Done()
	err = rt.LoadLocations()
	if err != nil {
		log.Printf("failed to load weather locations: %v\n", err)
		return
	}

	testCurrent(t, rt)
	testHourlyQuery(t, rt)
	// testHourlyFile(t, rt)
}

func testHourlyQuery(t *testing.T, rt *Runtime) {
	testHourly(t, rt, true)
}
func testHourlyFile(t *testing.T, rt *Runtime) {
	testHourly(t, rt, false)
}

func testCurrent(t *testing.T, rt *Runtime) {
	loc := rt.Locations[0]
	tm := time.Now().Truncate(time.Minute)
	tmAfter, tmBefore := AfterBeforeTime(tm, HISTORY_DURATION, HISTORY_MINUTE_INTERVAL)
	after, before := tmAfter.Format(TIME_LAYOUT), tmBefore.Format(TIME_LAYOUT)
	t.Log(loc.ID, tm, after, before)
	current, err := rt.SelectHistory(loc.ID, after, before)
	if err != nil {
		t.Log(tm, HISTORY_DURATION, after, before)
		t.Fatal(err)
	}
	loc.BuildCurrentProperties(current)

	for _, h := range current {
		t.Logf("%v %v %v", h.LocationID, h.Time, h.Temperature)
	}
	for _, item := range loc.CurrentProperties.Items {
		// t.Logf("min %v max %v", item.ScaleMin, item.ScaleMax)
		t.Logf("min %v max %v", item.Min, item.Max)
	}

	t.Log(len(current))
}

func testHourly(t *testing.T, rt *Runtime, bQuery bool) {
	var err error
	location := rt.Locations[0]
	hourly := &WeatherHourly{}
	if bQuery {
		query := fmt.Sprintf(weatherFormat, weatherHeader,
			location.Latitude, location.Longitude, location.Zone,
			hourlyTrailer)
		err = queryDecodeSave(query, &hourly)
		if err != nil {
			log.Printf("queryDecodeSave: %v", err)
			t.Fatal(err)
		}
	} else {
		err = readAndDecodeFile(&hourly)
		if err != nil {
			log.Printf("readAndDecodeFile: %v", err)
			t.Fatal(err)
		}
	}
	location.WeatherHourly = hourly
	location.WeatherHourly.UpdateTime = time.Now()
	location.BuildHourlyProperties()
	for _, item := range location.HourlyProperties.Items {
		t.Logf("min %v max %v", item.Min, item.Max)
	}

	log.Print(location.WeatherHourly.Hourly.Code)
	index := location.WeatherHourly.Hourly.Code[0]
	log.Print(index)

	tmpl, err := template.ParseGlob("../web/resources/html/weather-hourly.html")
	if err != nil {
		t.Fatal(err)
	}
	tmplHourly := tmpl.Lookup("weather.hourly")
	out, err := os.Create(hourlyHtmlName)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	err = tmplHourly.Execute(out, location)
	if err != nil {
		t.Fatal(err)
	}
}

func queryDecodeSave(query string, w any) error {
	resp, err := http.Get(query)
	if err != nil {
		log.Printf("query %s\n\terror: %v", query, err)
	}
	r := resp.Body
	defer r.Close()
	buf, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %v", err)
	}
	err = json.Unmarshal(buf, w)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %v\n%v", err, string(buf))
	}
	out, err := os.Create(hourlyJsonName)
	if err != nil {
		return err
	}
	defer out.Close()
	out.Write(buf)
	return nil
}

func readAndDecodeFile(w any) error {
	// r io.ReadCloser
	r, err := os.Open(hourlyJsonName)
	if err != nil {
		log.Print(err)
		return err
	}
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

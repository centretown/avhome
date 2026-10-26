package web

import (
	"avhome/weather"
	"bytes"
	"html/template"
	"log"
	"time"
)

type Broadcaster interface {
	Broadcast(string)
}

func broadcastLocation(location *weather.Location, t *template.Template, broadcaster Broadcaster) {
	buf := bytes.Buffer{}
	err := t.Execute(&buf, location)
	if err != nil {
		log.Printf("broadcastLocation: %v", err)
		return
	}
	broadcaster.Broadcast(buf.String())
}

func broadcastLocations(locations []*weather.Location, t *template.Template, broadcaster Broadcaster) {
	for _, location := range locations {
		time.Sleep(time.Second)
		broadcastLocation(location, t, broadcaster)
	}
}

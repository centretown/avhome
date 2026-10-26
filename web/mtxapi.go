package web

import (
	"avhome/query"
	"avhome/socket"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

type MtxPathSourceType string

// MtxPathSource is a source.
type MtxPathSource struct {
	Type MtxPathSourceType `json:"type"`
	ID   string            `json:"id"`
}

type MtxItem struct {
	Name                 string         `json:"name"`
	ConfName             string         `json:"confName"`
	Available            bool           `json:"available"`
	AvailableTime        *time.Time     `json:"availableTime"`
	Online               bool           `json:"online"`
	OnlineTime           *time.Time     `json:"onlineTime"`
	Source               *MtxPathSource `json:"source"`
	InboundBytes         uint64         `json:"inboundBytes"`
	OutboundBytes        uint64         `json:"outboundBytes"`
	InboundFramesInError uint64         `json:"inboundFramesInError"`
	PlayerSource         string         `json:"-"`
}

// MtxPathList is a list of paths.
type MtxPathList struct {
	ItemCount int       `json:"itemCount"`
	PageCount int       `json:"pageCount"`
	Items     []MtxItem `json:"items"`
}

func GetMtxPaths(url string, user string, password string) (list *MtxPathList, err error) {
	list = &MtxPathList{}
	q := fmt.Sprintf("http://%s:%s@%s:9997/v3/paths/list", user, password, url)
	err = query.QueryAndDecode(q, list)
	if err != nil {
		log.Printf("paths query: %v: error: %v", q, err)
		return
	}
	log.Print(list.ItemCount)
	for i, item := range list.Items {
		var source string
		log.Printf("name: %s, available: %v, source: %v, type: %v",
			item.Name, item.Available, item.Source.ID, item.Source.Type)
		source, err = cameraSource(url, item.Name, time.Now().Add(time.Minute*-15))
		if err != nil {
			log.Printf("source query: %v: error: %v", q, err)
			// return
		}
		list.Items[i].PlayerSource = source
	}

	return
}

type MtxRecording struct {
	Start    time.Time `json:"start"`
	Duration float64   `json:"duration"`
	Url      string    `json:"url"`
}
type MtxRecordings []*MtxRecording

const (
	dtLayout     = "2006-01-02T15:04 MST"
	playerPrefix = "/player/"
)

func writeError(w http.ResponseWriter, r *http.Request, err error, status int) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
	log.Print(r.RequestURI, " Error: ", err)
}

func handlePlayers(mux *http.ServeMux, sockServer *socket.Server, mtxUrl string, paths *MtxPathList, template *template.Template) {
	pathLookup := make(map[string]*MtxItem)
	for _, item := range paths.Items {
		pathLookup[item.Name] = &item
	}

	mux.HandleFunc(playerPrefix, func(w http.ResponseWriter, r *http.Request) {
		cameraName := strings.TrimPrefix(r.RequestURI, playerPrefix)
		log.Print(cameraName)
		camera, ok := pathLookup[cameraName]
		if !ok {
			err := fmt.Errorf("%s not mapped", cameraName)
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		log.Print(camera.Name)

		err := r.ParseForm()
		if err != nil {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		// request is sent via local time so get our time zone before converting
		zone, _ := time.Now().Zone()
		requestTime := r.FormValue("player-time") + " " + zone
		log.Print("requestTime: ", requestTime)
		start, err := time.Parse(dtLayout, requestTime)
		if err != nil {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		source, err := cameraSource(mtxUrl, cameraName, start)
		if err != nil {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		var data struct {
			Name         string
			PlayerSource string
		}
		data.Name = camera.Name
		data.PlayerSource = source

		sourceTmpl := template.Lookup("camera.video.source")
		err = sourceTmpl.Execute(w, &data)
		if err != nil {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}
		// w.WriteHeader(http.StatusOK)
	})
}

func cameraSource(mtxUrl string, cameraName string, start time.Time) (source string, err error) {
	startf, endf := start.Format(time.RFC3339),
		start.Add(15*time.Minute).Format(time.RFC3339)
	const playerPort = "9996"
	var recordings MtxRecordings
	q := fmt.Sprintf("http://%s:%s/list?path=%s&start=%s&end=%s",
		mtxUrl, playerPort, cameraName, startf, endf)
	log.Print(q)

	err = query.QueryAndDecode(q, &recordings)
	if err != nil {
		return
	}

	if len(recordings) < 1 {
		err = fmt.Errorf("no recordings were found")
		return
	}
	source = recordings[0].Url + "&format=mp4"
	return
}

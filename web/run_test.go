package web

import (
	"avhome/query"
	"fmt"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	const (
		startMinute = 4  // 4 minutes after the hour
		spanMinute  = 15 // every 15 minutes
	)

	tick := firstTicker(startMinute, spanMinute)
	t.Logf("%v", tick.String())
	// next := time.Now().Add(tick)
	// if next.Minute() != startMinute {
	// 	t.Log(next.Minute(), startMinute)
	// 	t.Fatal()
	// }
}

func TestRecording(t *testing.T) {
	const (
		url          = "http://10.0.0.7"
		port         = "9996"
		start        = "2026-10-17T00:00:00-04:00"
		end          = "2026-10-17T02:00:00-04:00"
		requestListF = "http://localhost:9996/list?path=[mypath]&start=[start]&end=[end]"
		requestPathF = "http://localhost:9996/get?path=mypath&start=2024-01-14T16%3A33%3A17%2B00%3A00&duration=200.5"
	)
	var (
		recordings = &MtxRecordings{}
		paths      = []string{"cam", "cam1", "cam4"}
	)
	for _, path := range paths {
		q := fmt.Sprintf("%s:%s/list?path=%s&start=%s&end=%s",
			url, port, path, start, end)
		err := query.QueryAndDecode(q, recordings)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(q)
		for _, recording := range *recordings {
			d := time.Duration(recording.Duration * float64(time.Second))
			t.Log(d, recording)
		}
	}
}

func TestRun(t *testing.T) {
	var err error

	webRun.Paths, err = GetMtxPaths(webRun.MtxUrl, "dave", "football")
	if err != nil {
		t.Fatalf("failed to contact Mtx: %v\n", err)
		return
	}

	webRun.addCameraSections()
	for _, a := range webRun.Sections {
		t.Log(a, a.Sections[0])
	}

}

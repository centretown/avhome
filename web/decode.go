package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func QueryAndDecode(query string, w any) (err error) {
	var resp *http.Response
	for attempt := range 3 {
		resp, err = http.Get(query)
		if err != nil {
			log.Printf("query current attempt: %v\n\terror: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}
		err = ReadAndDecode(resp.Body, w)
		if err != nil {
			log.Printf("decode current attempt: %v, error: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return
}

func ReadAndDecode(r io.ReadCloser, w any) error {
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

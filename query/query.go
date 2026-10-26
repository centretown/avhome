package query

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ServiceError struct {
	Reason string `json:"reason"`
	Error  bool   `json:"error"`
}

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
		resp.Body.Close()
		if err != nil {
			log.Printf("decode current attempt: %v, error: %v", attempt, err)
			time.Sleep(time.Second)
			continue
		}

		return
	}

	return
}

func ReadAndDecode(r io.ReadCloser, w any) error {
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

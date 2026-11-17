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

func CheckServiceError(buf []byte) error {
	var serviceError ServiceError
	json.Unmarshal(buf, &serviceError)
	if serviceError.Error {
		// {"reason":"The service is overloaded","error":true}
		return fmt.Errorf("service error:'%v' '%s'",
			serviceError.Error, serviceError.Reason)
	}
	return nil
}

func ReadAndDecode(r io.ReadCloser, w any) (err error) {
	var buf []byte
	defer func() {
		if err != nil {
			log.Printf("ReadAndDecode error '%v'", err)
		}
	}()

	buf, err = io.ReadAll(r)
	if err != nil {
		return
	}

	err = CheckServiceError(buf)
	if err != nil {
		return
	}

	err = json.Unmarshal(buf, w)
	if err != nil {
		log.Print(string(buf))
		return
	}

	return nil
}

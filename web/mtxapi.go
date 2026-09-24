package web

import (
	"fmt"
	"log"
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
}

// MtxPathList is a list of paths.
type MtxPathList struct {
	ItemCount int       `json:"itemCount"`
	PageCount int       `json:"pageCount"`
	Items     []MtxItem `json:"items"`
}

func GetMtxPaths(url string, user string, password string) (list *MtxPathList, err error) {
	list = &MtxPathList{}
	query := fmt.Sprintf("http://%s:%s@%s:9997/v3/paths/list", user, password, url)
	err = QueryAndDecode(query, list)
	if err != nil {
		log.Printf("paths query: %v: error: %v", query, err)
		return
	}
	log.Print(list.ItemCount)
	for _, item := range list.Items {
		log.Printf("name: %s, available: %v, source: %v, type: %v",
			item.Name, item.Available, item.Source.ID, item.Source.Type)
	}

	return
}

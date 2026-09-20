package web

import "time"

type APIPathSourceType string

// APIPathSource is a source.
type APIPathSource struct {
	Type APIPathSourceType `json:"type"`
	ID   string            `json:"id"`
}

type APIPath struct {
	Name                 string         `json:"name"`
	ConfName             string         `json:"confName"`
	Available            bool           `json:"available"`
	AvailableTime        *time.Time     `json:"availableTime"`
	Online               bool           `json:"online"`
	OnlineTime           *time.Time     `json:"onlineTime"`
	Source               *APIPathSource `json:"source"`
	InboundBytes         uint64         `json:"inboundBytes"`
	OutboundBytes        uint64         `json:"outboundBytes"`
	InboundFramesInError uint64         `json:"inboundFramesInError"`
}

// APIPathList is a list of paths.
type APIPathList struct {
	ItemCount int       `json:"itemCount"`
	PageCount int       `json:"pageCount"`
	Items     []APIPath `json:"items"`
}

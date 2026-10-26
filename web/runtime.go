package web

import (
	"avhome/section"
	"avhome/weather"
	"html/template"
)

type Runtime struct {
	MtxUrl   string
	Sections []*section.Section
	Paths    *MtxPathList
	Weather  *weather.Runtime
	template *template.Template
}

func (webRun *Runtime) addCameraSections() {
	for _, p := range webRun.Paths.Items {
		item := &section.Section{
			Name:   p.Name,
			Icon:   "videocam",
			Parent: "Cameras",
		}
		switch p.Name {
		case "cam":
			item.Title = "Front"
		case "cam1":
			item.Title = "West"
		case "cam4":
			item.Title = "East"
		default:
			item.Title = p.Name
		}
		item.Sections = []*section.Section{
			{
				Name:   p.Name + "-player",
				Title:  item.Title + " (recording)",
				Icon:   "playlist_play",
				Parent: "Cameras",
			},
		}
		item.Title += " (live)"
		webRun.Sections = append(webRun.Sections, item)
	}
}

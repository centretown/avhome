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
		title := ""
		switch p.Name {
		case "cam":
			title = "Front"
		case "cam1":
			title = "West"
		case "cam4":
			title = "East"
		default:
			title = p.Name
		}
		item.Title = title + " (live)"
		item.Sections = []*section.Section{
			{
				Name:   p.Name + "-player",
				Title:  title + " (recording)",
				Icon:   "playlist_play",
				Parent: "Cameras",
			},
		}
		webRun.Sections = append(webRun.Sections, item)
	}
}

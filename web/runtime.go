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

func (webRun *Runtime) addCameraActions() {
	for _, p := range webRun.Paths.Items {
		section := &section.Section{
			Name:   p.Name,
			Icon:   "videocam",
			Parent: "Cameras",
		}
		switch p.Name {
		case "cam":
			section.Title = "Front"
		case "cam1":
			section.Title = "West"
		case "cam4":
			section.Title = "East"
		default:
			section.Title = p.Name
		}
		webRun.Sections = append(webRun.Sections, section)
	}
}

package controllers

import "go-imagecloud/views"

func NewStatic() *Static {
	return &Static{
		Home: views.NewView(
			"bootstrap", "static/home"),
		Contact: views.NewView(
			"bootstrap", "static/contact"),
		Faq: views.NewView(
			"bootstrap", "static/faq"),
		NotFound: views.NewView(
			"bootstrap", "static/notFound"),
	}
}

type Static struct {
	Home      *views.View
	Contact   *views.View
	Faq       *views.View
	NotFound  *views.View
}

package view

import (
	"github.com/jcarm010/kodimerce/settings"
	"net/http"
)

type View struct {
	Title string
	MetaDescription string
	Keywords string
	CompanyName string
	CompanyNameAlternate string
	CompanyUrl string
	PageUrl string
	FacebookUrl string
	InstagramUrl string
}

func NewView(request *http.Request, title string, metaDescription string) *View {
	return &View{
		Title: title,
		MetaDescription: metaDescription,
		CompanyName: settings.COMPANY_NAME,
		CompanyNameAlternate: settings.COMPANY_NAME_ALTERNATE,
		CompanyUrl: settings.ServerUrl(request),
		PageUrl: settings.ServerUrl(request) + request.URL.String(),
		FacebookUrl: settings.FACEBOOK_URL,
		InstagramUrl: settings.INSTAGRAM_URL,
	}
}

package view

import (
	"github.com/jcarm010/kodimerce/settings"
	"net/http"
	"fmt"
)

type View struct {
	Title string
	MetaDescription string
	Keywords string
	CompanyName string
	CompanyNameAlternate string
	ContactEmail string
	ContactPhone string
	CompanyUrl string
	CanonicalUrl string
	PageUrl string
	FacebookUrl string
	InstagramUrl string
	TwitterUrl string
	LinkedInUrl string
	YouTubeUrl string
}

func NewView(request *http.Request, title string, metaDescription string) *View {
	httpHeader := "http"
	if request.TLS != nil {
		httpHeader = "https"
	}

	newUrl := fmt.Sprintf("%s://%s%s", httpHeader, request.Host, request.URL.Path)
	if request.URL.RawQuery != "" {
		newUrl += "?" + request.URL.RawQuery
	}

	return &View{
		Title: title,
		MetaDescription: metaDescription,
		CompanyName: settings.COMPANY_NAME,
		CompanyNameAlternate: settings.COMPANY_NAME_ALTERNATE,
		ContactEmail: settings.COMPANY_CONTACT_EMAIL,
		ContactPhone: settings.COMPANY_CONTACT_PHONE,
		CompanyUrl: settings.ServerUrl(request),
		CanonicalUrl: newUrl,
		PageUrl: settings.ServerUrl(request) + request.URL.String(),
		FacebookUrl: settings.FACEBOOK_URL,
		InstagramUrl: settings.INSTAGRAM_URL,
		TwitterUrl: settings.TWITTER_URL,
		LinkedInUrl: settings.LINKEDIN_URL,
		YouTubeUrl: settings.YOUTUBE_URL,
	}
}

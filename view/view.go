package view

import (
	"github.com/jcarm010/kodimerce/settings"
	"net/http"
	"fmt"
	"strings"
	"log"
	"time"
	"html/template"
)

type View struct {
	Request *http.Request
	Author string
	Title string
	MetaDescription string
	Keywords string
	CompanyName string
	CompanyNameAlternate string
	CompanyMailingAddress string
	ContactEmail string
	ContactPhone string
	CompanyUrl string
	CompanyGoogleMapsUrl string
	CompanyGoogleMapsEmbedUrl string
	CanonicalUrl string
	PageUrl string
	FacebookUrl string
	InstagramUrl string
	TwitterUrl string
	LinkedInUrl string
	YouTubeUrl string
	TwitterHandle string
}

func (v *View) DateTimeFormat (d time.Time ) (template.HTML) {
	return template.HTML(d.Format("2006-01-02T15:04:05.999999-07:00"))
}

func (v *View) FullUrl(u string) string {
	log.Printf("Url U: %s", u)
	var newUrl string = u
	if strings.HasPrefix(u, "/") {
		newUrl = settings.ServerUrl(v.Request) + u
	}else if !strings.HasPrefix(u, "http") {
		newUrl = settings.ServerUrl(v.Request) + "/" + u
	}
	return newUrl
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
		Request: request,
		Author: settings.AUTHOR,
		Title: title,
		MetaDescription: metaDescription,
		CompanyName: settings.COMPANY_NAME,
		CompanyNameAlternate: settings.COMPANY_NAME_ALTERNATE,
		CompanyMailingAddress: settings.COMPANY_MAILING_ADDRESS,
		ContactEmail: settings.COMPANY_CONTACT_EMAIL,
		ContactPhone: settings.COMPANY_CONTACT_PHONE,
		CompanyUrl: settings.ServerUrl(request),
		CompanyGoogleMapsUrl: settings.COMPANY_GOOGLE_MAPS_URL,
		CompanyGoogleMapsEmbedUrl: settings.COMPANY_GOOGLE_MAPS_EMBED_URL,
		CanonicalUrl: newUrl,
		PageUrl: settings.ServerUrl(request) + request.URL.String(),
		FacebookUrl: settings.FACEBOOK_URL,
		InstagramUrl: settings.INSTAGRAM_URL,
		TwitterUrl: settings.TWITTER_URL,
		LinkedInUrl: settings.LINKEDIN_URL,
		YouTubeUrl: settings.YOUTUBE_URL,
		TwitterHandle: settings.TWITTER_HANDLE,
	}
}

package view

import (
	"github.com/jcarm010/kodimerce/settings"
	"net/http"
	"fmt"
	"strings"
	"log"
	"time"
	"html/template"
	"io/ioutil"
	"encoding/json"
	"bytes"
)

var (
	TEMPLATES *template.Template
	CUSTOM_PAGES map[string]CustomPage
)

var fns = template.FuncMap{
	"plus1": func(x int) int {
		return x + 1
	},
	"DateTimeFormat": DateTimeFormat,
	"DateTimeFormatJSStr": DateTimeFormatJSStr,
	"DateTimeFormatHTMLAttr": DateTimeFormatHTMLAttr,
	"FullUrl": FullUrl,
	"CallTemplate": func(name string, data interface{}) (ret template.HTML, err error) {
		buf := bytes.NewBuffer([]byte{})
		err = TEMPLATES.ExecuteTemplate(buf, name, data)
		ret = template.HTML(buf.String())
		return
	},
}

type CustomPage struct {
	TemplateName string `json:"template_name"`
	Title string `json:"title"`
	MetaDescription string `json:"meta_description"`
	InSiteMap bool `json:"in_site_map"`
	ChangeFrequency string `json:"change_frequency"`
	Priority int `json:"priority"`
}

func init() {
	TEMPLATES = template.New("").Funcs(fns)
	TEMPLATES.ParseGlob("views/core-templates/*")
	TEMPLATES.ParseGlob("views/core-components/*")
	TEMPLATES.ParseGlob("views/templates/*")
	TEMPLATES.ParseGlob("views/components/*")

	customPages := struct{
		Pages map[string]CustomPage `json:"pages"`
	}{
		Pages: map[string]CustomPage{},
	}

	raw, err := ioutil.ReadFile("./custom-pages.json")
	if err == nil {
		err = json.Unmarshal(raw, &customPages)
		if err == nil {
			CUSTOM_PAGES = customPages.Pages
		}
	}
}

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
	FacebookAppId string
	InstagramUrl string
	TwitterUrl string
	LinkedInUrl string
	YouTubeUrl string
	TwitterHandle string
	GoogleAnalyticsAccountId string
}

func (v *View) DateTimeFormat (d time.Time ) (string) {
	return DateTimeFormat(d)
}

func (v *View) DateTimeFormatJSStr (d time.Time ) (template.JSStr) {
	return DateTimeFormatJSStr(d)
}

func (v *View) DateTimeFormatHTMLAttr (d time.Time ) (template.HTMLAttr) {
	return DateTimeFormatHTMLAttr(d)
}

func (v *View) FullUrl(u string) string {
	return FullUrl(u, v.Request)
}

func DateTimeFormat (d time.Time ) (string) {
	return d.Format("2006-01-02T15:04:05-07:00")
}

func DateTimeFormatJSStr (d time.Time ) (template.JSStr) {
	return template.JSStr(DateTimeFormat(d))
}

func DateTimeFormatHTMLAttr (d time.Time ) (template.HTMLAttr) {
	return template.HTMLAttr(DateTimeFormat(d))
}

func FullUrl(u string, r *http.Request) string {
	log.Printf("Url U: %s", u)
	var newUrl string = u
	if strings.HasPrefix(u, "/") {
		newUrl = settings.ServerUrl(r) + u
	}else if !strings.HasPrefix(u, "http") {
		newUrl = settings.ServerUrl(r) + "/" + u
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
		FacebookAppId: settings.FACEBOOK_APP_ID,
		InstagramUrl: settings.INSTAGRAM_URL,
		TwitterUrl: settings.TWITTER_URL,
		LinkedInUrl: settings.LINKEDIN_URL,
		YouTubeUrl: settings.YOUTUBE_URL,
		TwitterHandle: settings.TWITTER_HANDLE,
		GoogleAnalyticsAccountId: settings.GOOGLE_ANALYTICS_ACCOUNT_ID,
	}
}

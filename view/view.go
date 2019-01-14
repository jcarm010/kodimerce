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
	"golang.org/x/net/context"
	"github.com/jcarm010/kodimerce/entities"
)

var (
	Templates       *template.Template
	CustomPages     map[string]CustomPage
	CustomRedirects map[string]CustomRedirect
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
		err = Templates.ExecuteTemplate(buf, name, data)
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

type CustomRedirect struct {
	ToPath string `json:"to_path"`
	StatusCode int `json:"status_code"`
}

func init() {
	Templates = template.New("").Funcs(fns)
	for _, path := range []string{"views/core-templates/*", "views/core-components/*", "views/templates/*", "views/components/*"} {
		_, err := Templates.ParseGlob(path)
		if err != nil {
			print(fmt.Sprintf("error parsing: %s: %s", path, err))
		}
	}

	customPages := struct{
		Pages map[string]CustomPage `json:"pages"`
		Redirects map[string]CustomRedirect `json:"redirects"`
	}{
		Pages: map[string]CustomPage{},
		Redirects: map[string]CustomRedirect{},
	}

	raw, err := ioutil.ReadFile("./custom-pages.json")
	if err == nil {
		err = json.Unmarshal(raw, &customPages)
		if err == nil {
			CustomPages = customPages.Pages
			CustomRedirects = customPages.Redirects
		}
	}
}

type View struct {
	ServerSettings entities.ServerSettings
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
	GoogleTagManagerId string
	FareHarborShortName string
	OgImagePath string
}

func (v *View) GetBannerPath () (string) {
	if v.OgImagePath != "" {
		return v.OgImagePath
	}

	return "/assets/images/og-banner.png"
}

func (v *View) CurrentYear () (string) {
	return fmt.Sprintf("%v", time.Now().Year())
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

func (v *View) IsPathRoot() bool {
	return v.Request.URL.Path == "" || v.Request.URL.Path == "/"
}

func (v *View) PathContains(u string) bool {
	return strings.Contains(v.Request.URL.Path, u)
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
	var newUrl = u
	if strings.HasPrefix(u, "/") {
		newUrl = settings.ServerUrl(r) + u
	}else if !strings.HasPrefix(u, "http") {
		newUrl = settings.ServerUrl(r) + "/" + u
	}
	return newUrl
}

func NewView(request *http.Request, title string, metaDescription string, ctx context.Context) *View {
	httpHeader := "http"
	if request.TLS != nil {
		httpHeader = "https"
	}

	p := request.URL.Path
	if strings.HasSuffix(p, "/") {
		p = p[0 : len(p)-1]
	}

	newUrl := fmt.Sprintf("%s://%s%s", httpHeader, request.Host, p)
	if request.URL.RawQuery != "" {
		newUrl += "?" + request.URL.RawQuery
	}

	globalSettings := settings.GetGlobalSettings(ctx)
	return &View{
		ServerSettings: globalSettings,
		Request: request,
		Author: globalSettings.Author,
		Title: title,
		MetaDescription: metaDescription,
		CompanyName: globalSettings.CompanyName,
		CompanyNameAlternate: globalSettings.CompanyNameAlternate,
		CompanyMailingAddress: globalSettings.CompanyMailingAddress,
		ContactEmail: globalSettings.CompanyContactEmail,
		ContactPhone: globalSettings.CompanyContactPhone,
		CompanyUrl: settings.ServerUrl(request),
		CompanyGoogleMapsUrl: globalSettings.CompanyGoogleMapsUrl,
		CompanyGoogleMapsEmbedUrl: globalSettings.CompanyGoogleMapsEmbedUrl,
		CanonicalUrl: newUrl,
		PageUrl: settings.ServerUrl(request) + request.URL.String(),
		FacebookUrl: globalSettings.FacebookUrl,
		FacebookAppId: globalSettings.FacebookAppId,
		InstagramUrl: globalSettings.InstagramUrl,
		TwitterUrl: globalSettings.TwitterUrl,
		LinkedInUrl: globalSettings.LinkedInUrl,
		YouTubeUrl: globalSettings.YoutubeUrl,
		TwitterHandle: globalSettings.TwitterHandle,
		GoogleAnalyticsAccountId: globalSettings.GoogleAnalyticsAccountId,
		GoogleTagManagerId: globalSettings.GoogleTagManagerId,
		FareHarborShortName: globalSettings.FareHarborShortName,
	}
}


type BlogPostView struct{
	*View
	CanonicalUrl string
	Post         *entities.Post
	LatestPosts  []*entities.Post
	AboutBlog    string
	AmpImports   []AmpImport
}

type AmpImport struct {
	Name string
	URL  string
}

func (v BlogPostView) GetBannerPath () (string) {
	return v.Post.Banner
}

type CustomPageView struct{
	*View
	CustomPage *entities.DynamicPage
}

func (v CustomPageView) GetBannerPath () (string) {
	//page.DynamicPage.Banner.Path
	if v.CustomPage.Banner == nil {
		return v.View.GetBannerPath()
	}

	return v.CustomPage.Banner.Path
}
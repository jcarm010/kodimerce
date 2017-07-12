package view

import "github.com/jcarm010/kodimerce/settings"

type View struct {
	Title string
	MetaDescription string
	Keywords string
	CompanyName string
}

func NewView(title string, metaDescription string) *View {
	return &View{
		Title: title,
		MetaDescription: metaDescription,
		CompanyName: settings.COMPANY_NAME,
	}
}

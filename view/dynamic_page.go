package view

import "html/template"

type DynamicPage struct {
	Title string `json:"title"`
	MetaDescription string `json:"meta_description"`
	HasNavigation bool `json:"has_navigation"`
	HasBanner bool `json:"has_banner"`
	Banner *DynamicPageImageComponent `json:"banner"`
	Rows []*DynamicPageRow `json:"rows"`
}

type DynamicPageRow struct {
	ComponentName string `json:"component_name"` //this should be the name of the component to use
	RowSimpleComponent *DynamicPageRowSimpleComponent `json:"row_simple_component"`
	SeparatorTop bool `json:"separator_top"`
	SeparatorBottom bool `json:"separator_bottom"`
}

type DynamicPageImageComponent struct {
	Path string `json:"path"`
	AltText string `json:"alt_text"`
	SetSize bool `json:"set_size"`
	Width string `json:"width"`
	Height string `json:"height"`
}

type DynamicPageRowSimpleComponent struct {
	Header string `json:"header"`
	IsMainHeader bool `json:"is_main_header"`
	Description template.HTML `json:"description"`
	HasImage bool `json:"has_image"`
	Image *DynamicPageImageComponent `json:"image"`
	ImagePosition string `json:"image_position"`
}

func NewDynamicPage(title string, metaDescription string) *DynamicPage {
	return &DynamicPage{
		Title: title,
		MetaDescription: metaDescription,
		Rows: make([]*DynamicPageRow, 0),
	}
}
package km

import (
	"github.com/gocraft/web"
	"entities"
	"net/http"
	"google.golang.org/appengine/log"
)

func UpdateCompanyDetails(c *entities.ServerContext, w web.ResponseWriter, r *web.Request)  {
	if c.User.Role != entities.USER_TYPE_OWNER {
		log.Errorf(c.Context, "User is not owner: %+v", c)
		c.ServeJson(http.StatusForbidden, "Forbidden")
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "The parameters could not be parsed.")
		return
	}

	log.Infof(c.Context, "Company Update: %+v", r.Form)
	companyName := r.FormValue("company_name")
	if companyName == "" {
		c.ServeJson(http.StatusBadRequest, "Missing company name.")
		return
	}

	companyEmail := r.FormValue("company_email")
	companyPhone := r.FormValue("company_phone")
	companyAddress := r.FormValue("company_address")
	c.CompanyConfig.CompanyName = companyName
	c.CompanyConfig.CompanyEmail = companyEmail
	c.CompanyConfig.CompanyPhone = companyPhone
	c.CompanyConfig.CompanyAddress = companyAddress
	err = entities.SetCompanyConfig(c.Context, c.CompanyConfig)
	if err != nil {
		log.Errorf(c.Context, "Error saving server config: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error saving company details.")
		return
	}
}


func UpdateServerDetails(c *entities.ServerContext, w web.ResponseWriter, r *web.Request)  {
	if c.User.Role != entities.USER_TYPE_OWNER {
		log.Errorf(c.Context, "User is not owner: %+v", c)
		c.ServeJson(http.StatusForbidden, "Forbidden")
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing request: %+v", err)
		c.ServeJson(http.StatusBadRequest, "The parameters could not be parsed.")
		return
	}

	log.Infof(c.Context, "New Server Config: %+v", r.Form)

	c.ServerConfig.PrimaryColor = r.FormValue("primary_color")
	c.ServerConfig.SuccessColor = r.FormValue("success_color")
	c.ServerConfig.WarningColor = r.FormValue("warning_color")
	c.ServerConfig.DangerColor = r.FormValue("danger_color")
	c.ServerConfig.DefaultColor = r.FormValue("default_color")
	c.ServerConfig.PrimaryFontColor = r.FormValue("primary_font_color")
	c.ServerConfig.SuccessFontColor = r.FormValue("success_font_color")
	c.ServerConfig.WarningFontColor = r.FormValue("warning_font_color")
	c.ServerConfig.DangerFontColor = r.FormValue("danger_font_color")
	c.ServerConfig.DefaultFontColor = r.FormValue("default_font_color")

	err = entities.SetServerConfig(c.Context, c.ServerConfig)
	if err != nil {
		log.Errorf(c.Context, "Error saving server config: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error saving platform details.")
		return
	}
}
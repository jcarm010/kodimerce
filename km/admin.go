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
	c.Config.CompanyName = companyName
	c.Config.CompanyEmail = companyEmail
	c.Config.CompanyPhone = companyPhone
	c.Config.CompanyAddress = companyAddress
	err = entities.SetServerConfig(c.Context, c.Config)
	if err != nil {
		log.Errorf(c.Context, "Error saving server config: %+v", err)
		c.ServeJson(http.StatusInternalServerError, "Unexpected error saving company details.")
		return
	}
}

package km

import (
	"github.com/gocraft/web"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"entities"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func SetupServerInit(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	context := appengine.NewContext(r.Request)

	config, err := entities.GetServerConfig(context)
	if err != entities.ENTITY_NOT_FOUND_ERROR {
		c.ServeJson(http.StatusForbidden, "Server already configured.")
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Errorf(context, "Error parsing parameters: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not get parameters.")
		return
	}

	companyName := r.FormValue("company_name")
	if companyName == "" {
		log.Errorf(context, "Missing company name")
		c.ServeJson(http.StatusBadRequest, "Missing company name.")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		log.Errorf(context, "Missing first name.")
		c.ServeJson(http.StatusBadRequest, "Missing name.")
		return
	}

	lastName := r.FormValue("last_name")
	if lastName == "" {
		log.Errorf(context, "Missing last name")
		c.ServeJson(http.StatusBadRequest, "Missing last name.")
		return
	}

	email := r.FormValue("email")
	if email == "" {
		log.Errorf(context, "Missing email")
		c.ServeJson(http.StatusBadRequest, "Missing email.")
		return
	}

	password := r.FormValue("password")
	if password == "" {
		log.Errorf(context, "Missing password")
		c.ServeJson(http.StatusBadRequest, "Missing password.")
		return
	}

	companyAddress := ""
	companyPhone := ""
	companyEmail := ""

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf(context, "Error creating password: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not generate password")
		return
	}

	user := entities.NewUser(email, name, lastName, entities.USER_TYPE_OWNER, string(passwordBytes))
	err = entities.CreateUser(context, user)
	if err != nil {
		log.Errorf(context, "Error creating first user: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Unexpected error creating user.")
		return
	}

	config = entities.NewServerConfig(companyName, companyAddress, companyEmail, companyPhone)
	log.Infof(context, "New Configuration: %+v", config)
	err = entities.SetServerConfig(context, config)
	if err != nil {
		log.Errorf(context, "Error saving configuration: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Unexpected error saving configuration")
		return
	}
}

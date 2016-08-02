package initial_config

import (
	"github.com/gocraft/web"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"entities"
	"fmt"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func ServeJson(w web.ResponseWriter, status int, value interface{}){
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	bts, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "%s", bts)
}

func SetupServerInit(w web.ResponseWriter, r *web.Request) {
	context := appengine.NewContext(r.Request)

	config, err := entities.GetServerConfig(context)
	if err != entities.ENTITY_NOT_FOUND_ERROR {
		ServeJson(w, http.StatusForbidden, "Server already configured.")
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Errorf(context, "Error parsing parameters: %+v", err)
		ServeJson(w, http.StatusBadRequest, "Could not get parameters.")
		return
	}

	companyName := r.FormValue("company_name")
	if companyName == "" {
		log.Errorf(context, "Missing company name")
		ServeJson(w, http.StatusBadRequest, "Missing company name.")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		log.Errorf(context, "Missing first name.")
		ServeJson(w, http.StatusBadRequest, "Missing name.")
		return
	}

	lastName := r.FormValue("last_name")
	if lastName == "" {
		log.Errorf(context, "Missing last name")
		ServeJson(w, http.StatusBadRequest, "Missing last name.")
		return
	}

	email := r.FormValue("email")
	if email == "" {
		log.Errorf(context, "Missing email")
		ServeJson(w, http.StatusBadRequest, "Missing email.")
		return
	}

	password := r.FormValue("password")
	if password == "" {
		log.Errorf(context, "Missing password")
		ServeJson(w, http.StatusBadRequest, "Missing password.")
		return
	}

	companyAddress := ""
	companyPhone := ""
	companyEmail := ""

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf(context, "Error creating password: %+v", err)
		ServeJson(w, http.StatusBadRequest, "Could not generate password")
		return
	}

	user := entities.NewUser(email, name, lastName, string(passwordBytes))
	err = entities.CreateUser(context, user)
	if err != nil {
		log.Errorf(context, "Error creating first user: %+v", err)
		ServeJson(w, http.StatusBadRequest, "Unexpected error creating user.")
		return
	}

	config = entities.NewServerConfig(companyName, companyAddress, companyEmail, companyPhone)
	log.Infof(context, "New Configuration: %+v", config)
	err = entities.SetServerConfig(context, config)
	if err != nil {
		log.Errorf(context, "Error saving configuration: %+v", err)
		ServeJson(w, 400, "Unexpected error saving configuration")
		return
	}
}

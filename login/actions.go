package login

import (
	"github.com/gocraft/web"
	"entities"
	"google.golang.org/appengine/log"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine/datastore"
	"github.com/satori/go.uuid"
)

func Login(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c.Context, "Error parsing parameters: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Could not get parameters.")
		return
	}

	email := r.FormValue("email")
	if email == "" {
		log.Errorf(c.Context, "Missing email.")
		c.ServeJson(http.StatusBadRequest, "Missing email.")
		return
	}

	password := r.FormValue("password")
	if password == "" {
		log.Errorf(c.Context, "Missing password.")
		c.ServeJson(http.StatusBadRequest, "Missing password.")
		return
	}

	user, err := entities.GetUser(c.Context, email)
	if err == datastore.ErrNoSuchEntity {
		log.Infof(c.Context, "Account %s not found", user)
		c.ServeJson(http.StatusBadRequest, "Account not found.")
		return
	}else if err != nil {
		log.Errorf(c.Context, "Error getting user: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Unexpected error getting account.")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Errorf(c.Context, "Password error: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Invalid credentials.")
		return
	}

	cookieToken := uuid.NewV4().String()
	err = entities.StoreSessionToken(c.Context, entities.NewSessionToken(email, cookieToken))
	if err != nil {
		log.Errorf(c.Context, "Error creating session: %+v", err)
		c.ServeJson(http.StatusBadRequest, "Unexpected error creating session.")
		return
	}

	cookie := http.Cookie{Name: "session", Value: cookieToken}
	http.SetCookie(w, &cookie)
}

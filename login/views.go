package login

import (
	"github.com/gocraft/web"
	"html/template"
	"google.golang.org/appengine/log"
	"entities"
	"net/http"
)

type Person struct {
	UserName string
}

func LoginView(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	context := c.Context
	t, err := template.ParseFiles("./login/templates/login.html")
	if err != nil {
		log.Errorf(context, "Error parsing login template: %+v", err)
	}
	data := map[string] interface{}{
		"Server": c.Config,
	}
	t.Execute(w, data)
}

func UserLanding(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	err := c.SetUserContext()
	if err != nil {
		http.Redirect(w, r.Request, "/login", http.StatusTemporaryRedirect)
		return
	}

	switch c.User.Role {
	case entities.USER_TYPE_OWNER:
		http.Redirect(w, r.Request, "/admin", http.StatusTemporaryRedirect)
		break
	default:
		http.Redirect(w, r.Request, "/login", http.StatusTemporaryRedirect)
	}
}
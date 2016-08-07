package km

import (
	"github.com/gocraft/web"
	"entities"
	"html/template"
	"google.golang.org/appengine/log"
	"io/ioutil"
	"google.golang.org/appengine"
	"net/http"
)

func ServerInitView(w web.ResponseWriter, r *web.Request) {
	context := appengine.NewContext(r.Request)
	bts, err := ioutil.ReadFile("./km/templates/server_config.html")
	if err != nil {
		log.Errorf(context, "Error reading init template: %+v", err)
	}
	w.Write(bts)
}

func AdminView(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	context := c.Context
	t, err := template.ParseFiles("./km/templates/admin.html")
	if err != nil {
		log.Errorf(context, "Error parsing login template: %+v", err)
	}
	data := map[string] interface{}{
		"Server": c.Config,
		"User": c.User,
	}
	t.Execute(w, data)
}

func LoginView(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	context := c.Context
	t, err := template.ParseFiles("./km/templates/login.html")
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
package login

import (
	"github.com/gocraft/web"
	"html/template"
	"google.golang.org/appengine/log"
	"entities"
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

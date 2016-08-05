package admin

import (
	"github.com/gocraft/web"
	"entities"
	"html/template"
	"google.golang.org/appengine/log"
)

func AdminView(c *entities.ServerContext, w web.ResponseWriter, r *web.Request) {
	context := c.Context
	t, err := template.ParseFiles("./admin/templates/admin.html")
	if err != nil {
		log.Errorf(context, "Error parsing login template: %+v", err)
	}
	data := map[string] interface{}{
		"Server": c.Config,
		"User": c.User,
	}
	t.Execute(w, data)
}

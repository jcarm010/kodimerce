package views

import (
	"github.com/gocraft/web"
	"km"
	"html/template"
	"google.golang.org/appengine/log"
	"net/http"
)

type View struct {
	Title string
}

func HomeView(c *km.ServerContext, w web.ResponseWriter, r *web.Request) {
	var templates = template.Must(template.ParseGlob("views/template/*")) // cache this globally
	p := View{
		Title: "MandalaCrame",
	}

	err := templates.ExecuteTemplate(w, "home-page", p)
	if err != nil {
		log.Errorf(c.Context, "Error parsing home html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}
}

func AdminView(c *km.AdminContext, w web.ResponseWriter, r *web.Request) {
	p := View{
		Title: "Admin",
	}

	t, err := template.ParseFiles("views/admin.html") // cache this globally
	if err != nil {
		log.Errorf(c.Context, "Error parsing admin html file: %+v", err)
		c.ServeHTML(http.StatusInternalServerError, "Unexpected Error, please try again later.")
		return
	}

	t.Execute(w, p)
}

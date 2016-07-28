package login

import (
	"github.com/gocraft/web"
	"html/template"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
)

type Person struct {
	UserName string
}

func Login(w web.ResponseWriter, r *web.Request) {
	context := appengine.NewContext(r.Request)
	t, err := template.ParseFiles("./login/templates/login.html")
	if err != nil {
		log.Errorf(context, "Error parsing login template: %+v", err)
	}
	p := Person{UserName: "Astaxie"}
	t.Execute(w, p)
}

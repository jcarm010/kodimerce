package initial_config

import (
	"github.com/gocraft/web"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"io/ioutil"
)

func ServerInit(w web.ResponseWriter, r *web.Request) {
	context := appengine.NewContext(r.Request)
	bts, err := ioutil.ReadFile("./initial_config/templates/init.html")
	if err != nil {
		log.Errorf(context, "Error reading init template: %+v", err)
	}
	w.Write(bts)
}
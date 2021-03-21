package routes

import (
	"../service"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
)

func Route(srv *service.Server) {

	r := mux.NewRouter()
	r.HandleFunc("/", Chat)
	r.HandleFunc("/ws", srv.WsReceiver)

	http.Handle("/", r)
	httpPort := os.Getenv("HTTPPORT")
	log.Println("Starting http server on", httpPort)
	log.Fatal(http.ListenAndServe("localhost:"+httpPort, nil))
}

func Chat(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("index.html").ParseFiles("./assets/index.html", "./assets/header.html")
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		panic(err)
	}
}

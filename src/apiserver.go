package main

import (
	//"fmt"
	"net/http"
	"html/template"

	"github.com/gorilla/mux"
)

// data structure for the html template
type Repo struct {
	FullName string
	Link string
}
type Datas struct {
	Repos []Repo
}



func homePage(w http.ResponseWriter, r *http.Request) {

    tmpl := template.Must(template.ParseFiles("templates/homepage.html"))
    data := Datas {
        Repos: []Repo {
            {FullName: "Test", Link: "test.html"},
        },
    }

    tmpl.Execute(w, data)
}


func main() {
	// declare router
	r := mux.NewRouter()

	// url handler
	r.HandleFunc("/", homePage)

	// start server
	http.ListenAndServe(":8080", r)
}

package main

import (
	"fmt"
	"net/http"
	"html/template"

	"github.com/gorilla/mux"
)

type Todo struct {
	Title string
	Done bool
}

type TodoPageData struct {
	PageTitle string
	Todos []Todo
}

type ContactDetails struct {
	Email string
	Subject string
	Message string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello, welcome on my site\n")
	tmpl := template.Must(template.ParseFiles("templates/forms.html"))

	// get form values
	data := ContactDetails {
		Email: r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}


	tmpl.Execute(w, data)
}

func languagePage(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// language := vars["language"]
    //
	// fmt.Fprintf(w, "You've requested the language: %s\n", language)

	tmpl := template.Must(template.ParseFiles("templates/layout.html"))
	data := TodoPageData {
		PageTitle: "My Title Page",
		Todos: []Todo {
			{Title: "Task 1", Done: true},
			{Title: "Task 2", Done: true},
			{Title: "Task 3", Done: false},
			{Title: "Task 4", Done:true },
			{Title: "Task 5", Done: false},
		},
	}

	tmpl.Execute(w, data)
}

func main() {
	// declare router
	r := mux.NewRouter()

	// url handler
	r.HandleFunc("/", homePage)
	r.HandleFunc("/search", languagePage).Methods("GET")

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static", http.StripPrefix("/static/", fs))

	// start server
	http.ListenAndServe(":8080", r)
	fmt.Println("Server listenning on 8080..")
}

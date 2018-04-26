package main

import (
	"net/http"
	"fmt"
)


func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Main page")
}

func search(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Your are on the search page")
}


func main() {
	http.HandleFunc("/", sayHello)
	http.HandleFunc("/search", search)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

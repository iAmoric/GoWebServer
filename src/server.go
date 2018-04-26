package main

import (
	"net/http"
	"fmt"
    "log"
)


func rootPage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Your are on the main page")

}


func searchPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Your are on the search page")
}


func main() {
    // url handler
    http.HandleFunc("/", rootPage)
	http.HandleFunc("/search", searchPage)

    // start server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
    log.Println("Server listening on 8080")
}

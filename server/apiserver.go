package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
    "sync"
	"log"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// for json object
type Owner struct {
    Login string `json:"login"`
}

type Repository struct {
	Name string `json:name`
    Full_Name string `json:"full_name"`
    Html_url string `json:"html_url"`
    Description string `json:"description"`
	Owner Owner
    Languages_url string `json:"languages_url"`
    Languages map[string]int	// map Name/Lines
}

type Structure struct {
	Repositories []Repository `json:"items"`
}

var structure = new(Structure)
var repositories []Repository

// global map to store languages name and number of line for each
//var Lmap = make(map[string]int)
var mapLocker = struct{
    sync.RWMutex
    Lmap map[string]int
} {Lmap: make(map[string]int)}


// This program use the OAuth authentication.
// Please put your Github API token here
var token = ""

// Number of repositories that will be recovered by the API
var nb int

// Limit the number of parallel goroutines
var maxGoroutines = 10

/*
This function check if there is any error.
In case of error, it prints the error and exits the program
*/
func checkError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}


/*
This function sends the request to the github api and returns the response
OAuth authentication is used
*/
func request(url string) *http.Response {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Fatal(err)
    }

    req.SetBasicAuth(token, "x-oauth-basic")

    client := http.Client{}
    res, err := client.Do(req)
    if err != nil {
        log.Println("StatusCode:", res.StatusCode)
        log.Fatal(err)
    }

    return res
}


/*
This function is a goroutine.
It makes a request with the given url to the github API.
Then it parses the response to get the languages name and the number of line
*/
func parseLanguageRouting(url string, i int, wg *sync.WaitGroup) {
    // request http api
    res := request(url)

    // read json data
    jsonResponce, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    checkError(err)

    // parse string
    repositories[i].Languages = parseStringLanguage(string(jsonResponce))

    // for the barrier
    wg.Done()

}


/*
This function parses the string languages_url form the json response.
It maps the language name as key and the number of line as value
*/
func parseStringLanguage(str string)  map[string]int{

    if len(str) == 2 {  //avoid repo without any language
        return nil
    }

    var i = 0
    var mark1 = -1
    var mark2 = -1
    var language_name string
    var number_line int
    var err error
    var m map[string]int

    m = make(map[string]int)

    for i < len(str) {
        c := str[i]

        if c == '"' && mark1 == -1 {  // begining of the language name
            mark1 = i + 1
        }

        if c == '"' && mark1 != -1 && mark2 == -1 && i > mark1 {  // end of the language name
            mark2 = i
            language_name = str[mark1:mark2]
            //fmt.Printf("\t%s: ", language_name)
            mark1 = i + 2 // begining of the number
            mark2 = -1
            i += 1 // shift in the string, restart after the ':'
        }

        if c == '}' || (c == ',' && mark1 != -1 && mark2 == -1 && i > mark1) {
            mark2 = i

            // cast string to int
            number_line, err = strconv.Atoi(str[mark1:mark2])
            checkError(err)
            m[language_name] = number_line

            // make map safe for concurrent use
            mapLocker.Lock()
            mapLocker.Lmap[language_name] = mapLocker.Lmap[language_name] + number_line
            mapLocker.Unlock()

            mark1 = i + 2 // begining of the language name
            mark2 = -1
            i += 1 // shift in the string, restart after the '""'
        }

        i += 1
    }

    return m

}


/*
This function makes the query to the api,parses the json responses,
and starts goroutines to treats json.
All goroutines wait for others.
*/
func apiRequest(url string, search bool) {


    // do request to http api
    var res = request(url)
	log.Printf("Request made to github api: %s\n", url)

	// read the json
	jsonResponce, err := ioutil.ReadAll(res.Body)

    checkError(err)
	res.Body.Close()

	// Parse json data
	if search {
		// use this because the json structure isn't the same
		err = json.Unmarshal([]byte(jsonResponce), structure)
		repositories = structure.Repositories
	} else {
		err = json.Unmarshal([]byte(jsonResponce), &repositories)
	}
    checkError(err)

	// iterate over each repository
	nb = len(repositories)
	log.Printf("%d repositories recovered\n", nb)

	// set the barrier for the nb goroutines
    var wg sync.WaitGroup
    wg.Add(nb)

	guard := make(chan struct{}, maxGoroutines)
	for i := 0; i < nb; i++ {
        // run goroutine for request and parse responses
		guard <-struct {} {} // block if already filled -> max 10 at the same times
		go func (url string, index int, waitgroup *sync.WaitGroup) {
			parseLanguageRouting(url, index, waitgroup)
			<- guard
		} (repositories[i].Languages_url, i, &wg)
    }

	// wait for all goroutines
    wg.Wait()
}


/*
This function prints the result from the basic query (the home page)
*/
func printHomePage(w http.ResponseWriter) {
	printHeader(w)

	fmt.Fprintf(w, "<h1>Liste des dépôts Github</h1>")
	fmt.Fprintf(w, "<ul>")
	for i := 0; i < nb; i++ {
		fmt.Fprintf(w, "<li><a target=\"_blank\" href=\"%s\">%s</a></li> ", repositories[i].Html_url,repositories[i].Full_Name)
	}
	fmt.Fprintf(w, "</ul>")

	fmt.Fprintf(w, "<hr>")
	fmt.Fprintf(w, "<h1>Statistiques sur les langages</h1>")

	// TODO optimize these loops
	for k := range mapLocker.Lmap { //each language
		fmt.Fprintf(w, "%d lines", mapLocker.Lmap[k]) // language name : language number lines
		fmt.Fprintf(w, "<ul>")
		for i := 0; i < nb; i++ {   // each repository
			for kk := range repositories[i].Languages { // each language of each repository
				if strings.EqualFold(kk, k) { // test if the repository has the concerning language
					fmt.Fprintf(w, "<li><a target=\"_blank\" href=\"%s\">%s</a> : ", repositories[i].Html_url,repositories[i].Full_Name)	// repo name
					fmt.Fprintf(w, "%d lines </li>", repositories[i].Languages[kk])	// repo number lines
				}
			}
		}
		fmt.Fprintf(w, "</ul>")
	}

	fmt.Fprintf(w, "</body></html>")
}


/*
This function prints the result from the search query
*/
func printSearchPage(w http.ResponseWriter, search string) {
	printHeader(w)

	fmt.Fprintf(w, "<h1>Liste des dépôts Github utilisant le langage %s</h1>", search)

	// TODO optimize these loops
	for k := range mapLocker.Lmap { //each language
		if strings.EqualFold(k, search) {
			fmt.Fprintf(w, "%d lines", mapLocker.Lmap[k]) // language name : language number lines
			fmt.Fprintf(w, "<ul>")
			for i := 0; i < nb; i++ {   // each repository
				for kk := range repositories[i].Languages { // each language of each repository
					if strings.EqualFold(kk, k) { // test if the repository has the concerning language
						fmt.Fprintf(w, "<li><a target=\"_blank\" href=\"%s\">%s</a> : ", repositories[i].Html_url,repositories[i].Full_Name)	// repo name
						fmt.Fprintf(w, "%d lines </li>", repositories[i].Languages[kk])	// repo number lines
					}
				}
			}
			fmt.Fprintf(w, "</ul>")
		}
	}

	fmt.Fprintf(w, "</body></html>")
}


/*
This function prints the commun header of the html page
*/
func printHeader(w http.ResponseWriter) {
	str :=`
		<html>
		<head>
			<title>Webserver GO GitHub API</title>
		</head>
		<body>
		<span>Faire une recherche par langage : </span>
	    <br>
	    <form action="/search" method="GET">
	    	<input type="text" name="language">
	    	<input type="submit" value="Rechercher">
	    </form>
		`
	fmt.Fprintf(w, str)
}


/*
This function is the handler for the home page
It makes the query to the api.
Then it print the result on the page
*/
func homePage(w http.ResponseWriter, r *http.Request) {

	log.Printf("New request from %s on %s", r.RemoteAddr, r.URL.Path)

    // Get datas from the github API
	// TODO : get 100 LAST public repositories
	var N = 130000000 // start point. TODO : find another method...
    var url = fmt.Sprintf("%s%d", "https://api.github.com/repositories?since=", N)
    apiRequest(url, false)

	// display the page
	printHomePage(w)


	log.Println("Page successfully loaded with datas")
}


/*
This function is the handler for the search page
It takes the language passed as parameter in the url
and makes the query to the api.
Then it print the result on the page
*/
func searchPage(w http.ResponseWriter, r *http.Request)  {
	log.Printf("New request from %s on %s", r.RemoteAddr, r.URL.Path)

	search :=r.FormValue("language")
	log.Printf("Searching for %s", search)

	var url = "https://api.github.com/search/repositories?q=language:" + search
    apiRequest(url, true)

	printSearchPage(w, search)

	log.Println("Page successfully loaded with datas")
}


func main() {
	// declare router
	r := mux.NewRouter()

	// url handler
	r.HandleFunc("/", homePage).Methods("GET")
	r.HandleFunc("/search", searchPage).Methods("GET")

	// start server
	http.ListenAndServe(":8080", r)
	log.Println("Server listenning on 8080")


}

package main

import (
	"fmt"
	"net/http"
	"html/template"
	"encoding/json"
	"io/ioutil"
    "sync"
	"log"
	"strconv"

	"github.com/gorilla/mux"
)

// data structure for the html template
type Repo struct {
	FullName string
	Link string
	Description string
}
type Datas struct {
	Repos []Repo
}

// json object
type Owner struct {
    Login string `json:"login"`
}
type Repository struct {
    Full_Name string `json:"full_name"`
    Html_url string `json:"html_url"`
    Description string `json:"description"`
    Languages_url string `json:"languages_url"`
    Languages map[string]int
}
var repositories []Repository

// global map to store languages name and number of line for each
//var Lmap = make(map[string]int)
var mapLocker = struct{
    sync.RWMutex
    Lmap map[string]int
} {Lmap: make(map[string]int)}


// This program use the OAuth authentication.
// Please put your Github API token here
var token = "c78019fdfdff8d99b179407dee3a56308ff1f1f2"

// Number of repositories that will be recovered by the API
var nb int

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

func apiRequest() {
    // TODO : get 100 last public repositories
	var N = 130000000 // start point. TODO : find another method...
    var repository_url = fmt.Sprintf("%s%d", "https://api.github.com/repositories?since=", N)

    // do request to http api
    var res = request(repository_url)
	log.Printf("Request made to github api...\n")

	// read the json
	jsonResponce, err := ioutil.ReadAll(res.Body)
    checkError(err)
	res.Body.Close()

	// Parse json data
	err = json.Unmarshal([]byte(jsonResponce), &repositories)
    checkError(err)

	// iterate over each repository
	nb = len(repositories)
	log.Printf("%d repositories recovered\n", nb)

	// set the barrier for the nb goroutines
    var wg sync.WaitGroup
    wg.Add(nb)

	for i := 0; i < nb; i++ {
        // run goroutine for request and parse responses
        go parseLanguageRouting(repositories[i].Languages_url, i, &wg)
    }

	// wait for all goroutines
    wg.Wait()
}


func homePage(w http.ResponseWriter, r *http.Request) {

	log.Printf("New request from %s on %s", r.RemoteAddr, r.URL.Path)
    // Get datas from the github API
    apiRequest()

	// mapLocker.Lmap contains global [language_name; number_line]
	// repository[i].Languages contains individual [language_name; number_line]

	// fill the data structure for the template
	Repos := make([]Repo, nb)
	for i := 0; i < nb; i++ {   // each repository
		Repos[i].FullName = repositories[i].Full_Name
		Repos[i].Link = repositories[i].Html_url
		Repos[i].Description = repositories[i].Description
	}
	data := Datas {
        Repos,
    }

    // Parse and execute template with datas
    tmpl := template.Must(template.ParseFiles("templates/homepage.html"))


    err := tmpl.Execute(w, data)
	checkError(err)
	log.Println("Page successfully loaded with datas")
}


func main() {
	// declare router
	r := mux.NewRouter()

	// url handler
	r.HandleFunc("/", homePage)

	// start server
	http.ListenAndServe(":8080", r)
	log.Println("Server listenning on 8080")
}
